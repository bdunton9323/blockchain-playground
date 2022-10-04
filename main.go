package main

import (
	"flag"
	"strings"

	"github.com/bdunton9323/blockchain-playground/contract"
	"github.com/bdunton9323/blockchain-playground/controllers"
	"github.com/bdunton9323/blockchain-playground/orders"
	log "github.com/sirupsen/logrus"
)

// I'll just hard code these here for now

// the URL of the ethereum node (I used Quorum running locally)
var ethNodeUrl = "http://172.13.3.1:8545"

var dbHost = "127.0.0.1:3306"
var dbName = "orderdb"
var dbUser = "db_user"
var dbPassword = "mysqlPassword"

func main() {
	privateKey := flag.String("privatekey", "", "The private key of the microservice (i.e. the vendor)")
	flag.Parse()

	// the signature code expects the bare hex string
	if strings.HasPrefix(*privateKey, "0x") {
		log.Fatal("Private key must not start with 0x")
	}

	orderRepo, err := orders.NewMariaDBOrderRepository(dbHost, dbName, dbUser, dbPassword)
	if err != nil {
		log.Fatal("Could not connect to database: %s", err.Error())
	}

	contractExecutor, err := contract.NewDeliveryContractExecutor(ethNodeUrl, *privateKey)
	if err != nil {
		log.Fatal("Could not build the contract executor: %s", err.Error())
	}

	var orderController = &controllers.OrderController{
		ServerPrivateKey: *privateKey,
		NodeUrl:          ethNodeUrl,
		OrderRepository:  orderRepo,
		ContractExecutor: contractExecutor,
	}
	controllers.NewApiRouter(orderController).Start()
}
