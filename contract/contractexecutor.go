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
		Client:           client,
		ServerPrivateKey: privKey,
	}, nil
}

// returns (tokenId, token address, error)
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

// Purches the token from the owner
func (_exec *DeliveryContractExecutor) BuyNFT(addressHex string, tokenId int64, buyerPrivateKey string, price int64) error {

	privKey, err := crypto.HexToECDSA(buyerPrivateKey)
	if err != nil {
		return err
	}

	nonce, err := _exec.getNonce(privKey)
	if err != nil {
		return err
	}

	txOpts := bind.NewKeyedTransactor(privKey)
	txOpts.Nonce = nonce

	// the amount of Ether being sent in the request, in wei
	txOpts.Value = big.NewInt(price)

	// Quorum is gasless, so this doesn't work there
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return err
	// }
	// txOpts.GasLimit = uint64(300000) // in gas units
	// txOpts.GasPrice = gasPrice

	contractAddress := common.HexToAddress(addressHex)
	contractInstance, err := NewDeliveryToken(contractAddress, _exec.Client)
	if err != nil {
		return err
	}

	buyerAddress := _exec.getAddressFromKey(privKey)

	// print a before balance so we can see that ether was actually transferred
	_exec.printBalances(buyerAddress, &contractAddress)

	tx, err := contractInstance.Buy(txOpts)
	if err != nil {
		return errors.New(fmt.Sprintf("Error paying for delivery: %v", err))
	}
	log.Infof("Tx sent with ID [%s]", tx.Hash().Hex())

	// TODO: this is not going to work because it's not waiting for the transaction to be mined
	// TODO: my contract is not transferring the balance to the seller
	// print an after balance so we can see that ether was actually transferred
	_exec.printBalances(buyerAddress, &contractAddress)

	return nil
}

// Returns the owner of the token
func (_exec *DeliveryContractExecutor) GetOwner(contractAddress string, privKeyHex string) (string, error) {

	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return "", err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce

	address := common.HexToAddress(contractAddress)
	contractInstance, err := NewDeliveryToken(address, _exec.Client)
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

// Destroys the token
func (_exec *DeliveryContractExecutor) BurnContract(contractAddress string, privKeyHex string) error {

	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce

	address := common.HexToAddress(contractAddress)
	contractInstance, err := NewDeliveryToken(address, _exec.Client)
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

func (_exec *DeliveryContractExecutor) printBalances(customerAddress *common.Address, contractAddress *common.Address) {
	customerBalance, err1 := _exec.Client.BalanceAt(context.Background(), *customerAddress, nil)
	contractBalance, err2 := _exec.Client.BalanceAt(context.Background(), *contractAddress, nil)
	if err1 == nil && err2 == nil {
		log.Infof("Customer (address [%s]) balance: %d", customerAddress.Hex(), customerBalance)
		log.Infof("Contract (address [%s]) balance: %d", contractAddress.Hex(), contractBalance)
	} else {
		log.Errorf("Failed to get ether balance of [%s] or [%s]. Reason: [%v]",
			customerAddress.Hex(),
			contractAddress.Hex(),
			func() string {
				if err1 != nil {
					return err1.Error()
				} else {
					return err2.Error()
				}
			}())
	}
}

// Converts a private key to the address
func (_exec *DeliveryContractExecutor) getAddressFromKey(privateKey *ecdsa.PrivateKey) *common.Address {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &address
}

// There is some functionalilty in the ABI bindings for subscribing to an events channel to watch for events
// emitted from the blockchain. It is not supported with the RPC connection interface I am using, since there
// is no long-lived connection there. I am leaving the code here regardless.
// ----
//
// func waitForMint(contractInstance *DeliveryToken) {
// 	log.Info("Waiting for the emitted event")
//
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go listenForMintedEvent(context.Background(), &wg, contractInstance)
// 	wg.Wait()
// }
//
// func listenForMintedEvent(ctx context.Context, wg *sync.WaitGroup, contract *DeliveryToken) {
// 	defer wg.Done()
//
// 	events := make(chan *DeliveryTokenNFTMinted)
//
// 	start := uint64(0)
// 	watchOpts := &bind.WatchOpts{
// 		Start: &start,
// 		Context: ctx,
// 	}
// 	subscription, err := contract.DeliveryTokenFilterer.WatchNFTMinted(watchOpts, events)
// 	if err != nil {
// 		log.Errorf("Could not watch for token minting event: %v", err)
// 	}
// 	defer subscription.Unsubscribe()
// }
