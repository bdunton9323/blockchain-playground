package contract

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

// returns the contract ID
// privKeyHex - the key without the 0x
func InstallContract(privKeyHex string) (string, error) {

	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	key, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not decode hex: %v", err))
	}

	privateKey, err := crypto.ToECDSA(key)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not create private key: %v", err))
	}

	// TODO: this is deprecated. It wants NewKeyedTransactorWithChainID, but I don't know what to use for chainId
	transactOps := bind.NewKeyedTransactor(privateKey)

	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not connect to Ethereum: %v", err))
	}

	// TODO: do I need to wait for it to be mined before returning, or does this wait automatically?
	address, _, _, err := DeployStorage(transactOps, client)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error deploying contract: %v", err))
	} else {
		log.Infof(fmt.Sprintf("Deployed contract with address [%v]", address.String()))
	}

	return address.String(), nil
}

func SetValue(newValue *big.Int, addressHex string, privKeyHex string) error {

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
	storageInstance, err := NewStorage(address, client)
	if err != nil {
		return err
	}

	tx, err := storageInstance.Set(auth, newValue)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value in contract: %v", err))
	}

	log.Infof("Tx sent with ID [%s]", tx.Hash().Hex())

	return nil
}

func GetValue(addressHex string) (*big.Int, error) {
	storage, err := getContract(addressHex)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error fetching contract with address [%v]", addressHex))
	}

	val, err := storage.StorageCaller.Get(nil)

	return val, nil
}

func getContract(addressHex string) (*Storage, error) {
	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	address := common.HexToAddress(addressHex)
	storage, err := NewStorage(address, client)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error fetching contract with address [%v]", addressHex))
	}
	return storage, nil
}

func buildTransactor(privKeyHex string) (*bind.TransactOpts, error) {
	key, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create private key: %v", err))
	}

	privateKey, err := crypto.ToECDSA(key)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create private key: %v", err))
	}
	return bind.NewKeyedTransactor(privateKey), nil
}
