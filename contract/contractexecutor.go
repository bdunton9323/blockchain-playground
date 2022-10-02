package contract

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

func MintNFT(addressHex string, privKeyHex string, nodeUrl string) error {
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

	
	//waitForMining(client, tx.Hash())

	log.Infof("Tx sent with ID [%s]", tx.Hash().Hex())

	return nil
}

// func waitForMining(client *ethclient.Client, txHash common.Hash) {
// 	NewMyTokenFilterer
// 	receipt, error := client.TransactionReceipt(context.Background(), txHash)
// 	receipt.PostState
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
