package contract

import (
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

// returns the contract ID
func InstallContract() string {

	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		log.Fatalf("Error dialing the node: %v", err)
	}

	// TODO: load in the key at startup time
	//       This is a throwaway key for local dev so I don't care that it's in git
	auth, err := bind.NewTransactorWithChainID(strings.NewReader("/home/bdunton/.ssh/ethereum_testing"), "superSecurePassword", big.NewInt(15))

	if err != nil {
		log.Fatalf("Could not connect to Ethereum: %v", err)
	}

	DeployStorage(auth, client)
	return "Installing the contract"
}
