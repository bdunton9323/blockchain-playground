package contract

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

// returns the contract ID
// privKeyHex - the key without the 0x
func DeployContract(privKeyHex string, nodeUrl string) (string, error) {

	client, err := ethclient.Dial(nodeUrl)
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
