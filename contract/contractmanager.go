package contract

import (
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

	// client, err := ethclient.Dial("http://172.13.3.1:8545")
	// if err != nil {
	// 	return errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	// }

	// address := common.HexToAddress(addressHex)
	// storage, err := NewStorage(address, client)
	storage, err := getContract(addressHex)
	if err != nil {
		return errors.New(fmt.Sprintf("Error fetching contract: %v", err))
	}

	transactOpts, err := buildTransactor(privKeyHex)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building transactor for contract [%v]", addressHex))
	}

	_, err = storage.StorageTransactor.Set(transactOpts, newValue)
	return err
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
