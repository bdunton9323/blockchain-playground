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
func InstallContract(keyHexValue string) (string, error) {

	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error dialing the node: %v", err))
	}

	key, err := hex.DecodeString(keyHexValue)
	if err != nil {
		log.Errorf("Could not decode hex: %v", err)
	}

	privateKey, err := crypto.ToECDSA(key)
	if err != nil {
		log.Errorf("Could not create private key: %v", err)
	}

	// TODO: this is deprecated. It wants NewKeyedTransactorWithChainID, but I don't know what to use for chainId
	transactOps := bind.NewKeyedTransactor(privateKey)

	if err != nil {
		log.Errorf("Could not connect to Ethereum: %v", err)
	}

	address, _, _, err := DeployStorage(transactOps, client)
	if err != nil {
		log.Errorf("Error deploying contract: %v", err)
	} else {
		log.Errorf("Deployed contract with address [%v]", address.String())
	}

	return address.String(), nil
}
