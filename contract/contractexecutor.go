package contract

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

type DeliveryContractExecutor struct {
	Client           *ethclient.Client
	ServerPrivateKey *ecdsa.PrivateKey
}

func NewDeliveryContractExecutor(nodeUrl string, serverPrivateKey string) (*DeliveryContractExecutor, error) {
	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not connect to ethereum node: %v", err))
	}

	privKey, err := crypto.HexToECDSA(serverPrivateKey)
	if err != nil {
		return nil, err
	}

	return &DeliveryContractExecutor{
		Client: client,
		ServerPrivateKey: privKey,
	}, nil
}


// return (tokenId, token address, error)
func (_exec *DeliveryContractExecutor) DeployContractAndMintNFT(
	privKeyHex string,
	nodeUrl string,
	purchasePrice int64,
	userAddress string) (*big.Int, string, error) {

	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return nil, "", err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce

	// Deploy the contract
	address, tx, tokenContract, err := DeployDeliveryToken(txOpts, _exec.Client)
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("Error deploying token contract: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to create contract", tx.Hash().Hex())

	// Need a new nonce for the next transaction
	txOpts.Nonce = nonce.Add(nonce, big.NewInt(1))
	tokenId, err := mintToken(tokenContract, txOpts, purchasePrice, common.HexToAddress(userAddress))

	contractAddress := address.Hex()
	return tokenId, contractAddress, nil
}

func mintToken(
	tokenContract *DeliveryToken,
	txOpts *bind.TransactOpts,
	purchasePrice int64,
	userAddress common.Address) (*big.Int, error) {

	log.Infof("Minting token to be purchased by [%v] for cost [%v]", userAddress.Hex(), purchasePrice)
	err := waitForContractAndMintToken(tokenContract, txOpts, purchasePrice, userAddress)
	if err != nil {
		return nil, err
	}

	tokenId, err := waitForTokenId(tokenContract)
	if err != nil {
		return nil, err
	}

	return tokenId, nil
}

// The transaction to deploy the contract has to be mined by the
// blockchain before we can use it. I'm sure there is a better way
// to do this, but until I am more familiar with Go and the ABI
// bindings, this will have to do.
func waitForContractAndMintToken(
	tokenContract *DeliveryToken,
	txOpts *bind.TransactOpts,
	purchasePrice int64,
	userAddress common.Address) error {

	tx, err := tokenContract.MintToken(txOpts, big.NewInt(purchasePrice), userAddress)

	numTries := 1
	for err != nil && numTries < 5 {
		numTries++
		tx, err = tokenContract.MintToken(txOpts, big.NewInt(purchasePrice), userAddress)
		if err != nil {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value in contract: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to mint a token", tx.Hash().Hex())

	return err
}

// The transaction to deploy the contract has to be mined by the
// blockchain before we can use it. I'm sure there is a better way
// to do this, but until I am more familiar with Go and the ABI
// bindings, this will have to do.
func waitForTokenId(tokenContract *DeliveryToken) (*big.Int, error) {
	tokenId, err := tokenContract.GetId(nil)
	numTries := 1
	for err != nil && numTries < 5 {
		time.Sleep(2 * time.Second)
		numTries++
		tokenId, err = tokenContract.GetId(nil)
	}
	if err != nil {
		log.Errorf("Error getting token ID. The transaction may not have been mined. %v", err)
		return nil, err
	}

	return tokenId, nil
}

// func waitForEvent(contractInstance *DeliveryToken) {
// 	log.Info("Waiting for the emitted event")

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go listenForMintedEvent(context.Background(), &wg, contractInstance)
// 	wg.Wait()
// }

// func listenForMintedEvent(ctx context.Context, wg *sync.WaitGroup, contract *DeliveryToken) {
// 	defer wg.Done()

// 	events := make(chan *DeliveryTokenNFTMinted)
// 	watchOpts := &bind.WatchOpts{
// 		// TODO: if this is reading from the latest, is there a chance I will miss my event?
// 		Start: nil,
// 		Context: ctx,
// 	}
// 	subscription, err := contract.DeliveryTokenFilterer.WatchNFTMinted(watchOpts, events)
// 	if err != nil {
// 		log.Errorf("Could not watch for token minting event: %v", err)
// 	}

// 	defer subscription.Unsubscribe()

// }

func BuyNFT(addressHex string, tokenId int64, privKeyHex string, nodeUrl string) error {
	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return err
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New(fmt.Sprintf("error casting public key to ECDSA: %v", err))
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return err
	// }

	auth := bind.NewKeyedTransactor(privKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Nonce = big.NewInt(int64(nonce))

	// TODO: get the price from the database
	auth.Value = big.NewInt(0) // in wei
	//auth.GasLimit = uint64(300000) // in units
	// auth.GasPrice = gasPrice

	address := common.HexToAddress(addressHex)
	//storageInstance, err := NewStorage(address, client)
	contractInstance, err := NewDeliveryToken(address, client)
	if err != nil {
		return err
	}

	tx, err := contractInstance.Buy(auth)
	if err != nil {
		return errors.New(fmt.Sprintf("Error paying for delivery: %v", err))
	}

	log.Infof("Tx sent with ID [%s]", tx.Hash().Hex())

	return nil
}

func GetOwner(contractAddress string, privKeyHex string) (string, error) {
	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return "", err
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New(fmt.Sprintf("error casting public key to ECDSA: %v", err))
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	txOpts := bind.NewKeyedTransactor(privKey)
	txOpts.Nonce = big.NewInt(int64(nonce))

	address := common.HexToAddress(contractAddress)
	contractInstance, err := NewDeliveryToken(address, client)
	if err != nil {
		return "", err
	}

	ownerAddress, err := contractInstance.GetOwner(nil)

	if err != nil {
		return "", err
	}
	owner := ownerAddress.Hex()
	return owner, nil
}

func BurnContract(contractAddress string, privKeyHex string) error {
	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return err
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New(fmt.Sprintf("error casting public key to ECDSA: %v", err))
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	txOpts := bind.NewKeyedTransactor(privKey)
	txOpts.Nonce = big.NewInt(int64(nonce))

	address := common.HexToAddress(contractAddress)
	contractInstance, err := NewDeliveryToken(address, client)
	if err != nil {
		return err
	}

	_, err = contractInstance.BurnToken(txOpts)

	if err != nil {
		log.Errorf("Failed to burn token: %v", err)
		return err
	}

	return nil
}


// Gets a nonce to use for the next transaction
func (_exec *DeliveryContractExecutor) getNonce(privateKey *ecdsa.PrivateKey) (*big.Int, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Error casting public key to ECDSA"))
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := _exec.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	return big.NewInt(int64(nonce)), nil
}