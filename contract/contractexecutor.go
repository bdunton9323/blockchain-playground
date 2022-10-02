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

// return (tokenId, token address, error)
func MintNFT(privKeyHex string, nodeUrl string, purchasePrice int64) (*big.Int, *string, error) {
	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, nil, err
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("error casting public key to ECDSA: %v", err))
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, nil, err
	}
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return nil, nil, err
	// }

	txOpts := bind.NewKeyedTransactor(privKey)
	txOpts.Nonce = big.NewInt(int64(nonce))
	// txOpts.Value = big.NewInt(0)     // in wei
	// txOpts.GasLimit = uint64(300000) // in units
	//txOpts.GasPrice = 0 // has to be 0 in quorum, apparently?

	// Deploy the contract
	address, tx, tokenContract, err := DeployMyToken(txOpts, client)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Error deploying token contract: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to create contract", tx.Hash().Hex())

	// address := common.HexToAddress(addressHex)
	// //storageInstance, err := NewStorage(address, client)
	// tokenContract, err := NewMyToken(address, client)
	// if err != nil {
	// 	return nil, err
	// }

	// use the contract to mint the token

	nonce++
	txOpts.Nonce = big.NewInt(int64(nonce))

	tokenId, err := mintToken(tokenContract, txOpts, purchasePrice)

	contractAddress := address.Hex()
	return tokenId, &contractAddress, nil
}

func mintToken(tokenContract *MyToken, txOpts *bind.TransactOpts, purchasePrice int64) (*big.Int, error) {

	waitForContractAndMintToken(tokenContract, txOpts, purchasePrice)

	tokenId, err := waitForTokenId(tokenContract)
	if err != nil {
		return nil, err
	}

	return tokenId, nil
}

func waitForContractAndMintToken(tokenContract *MyToken, txOpts *bind.TransactOpts, purchasePrice int64) error {
	tx, err := tokenContract.MintToken(txOpts, big.NewInt(purchasePrice))

	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value in contract: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to mint a token", tx.Hash().Hex())

	return err
}

// func withRetries(numRetries int, retriable func()) {
// 	numTries := 0
// 	for err != nil && numTries < numRetries {
// 		err := retriable()
// 		time.Sleep(2 * time.Second)
// 		numTries++
// 	}
// }

// The transaction to deploy the contract has to be mined by the 
// blockchain before we can use it. I'm sure there is a better way
// to do this, but until I am more familiar with Go and the ABI
// bindings, this will have to do.
func waitForTokenId(tokenContract *MyToken) (*big.Int, error) {
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

// func waitForEvent(contractInstance *MyToken) {
// 	log.Info("Waiting for the emitted event")

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go listenForMintedEvent(context.Background(), &wg, contractInstance)
// 	wg.Wait()
// }

// func listenForMintedEvent(ctx context.Context, wg *sync.WaitGroup, contract *MyToken) {
// 	defer wg.Done()

// 	events := make(chan *MyTokenNFTMinted)
// 	watchOpts := &bind.WatchOpts{
// 		// TODO: if this is reading from the latest, is there a chance I will miss my event?
// 		Start: nil,
// 		Context: ctx,
// 	}
// 	subscription, err := contract.MyTokenFilterer.WatchNFTMinted(watchOpts, events)
// 	if err != nil {
// 		log.Errorf("Could not watch for token minting event: %v", err)
// 	}

// 	defer subscription.Unsubscribe()

// }

func BuyNFT(addressHex string, privKeyHex string, nodeUrl string) error {
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
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(privKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address := common.HexToAddress(addressHex)
	//storageInstance, err := NewStorage(address, client)
	contractInstance, err := NewMyToken(address, client)
	if err != nil {
		return err
	}

	// TODO: get the token ID from somewhere
	tx, err := contractInstance.Buy(auth, big.NewInt(3))
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value in contract: %v", err))
	}

	log.Infof("Tx sent with ID [%s]", tx.Hash().Hex())

	return nil
}
