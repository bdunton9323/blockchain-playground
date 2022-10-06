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

// instance variables needed by the contract executor
type DeliveryContractExecutor struct {
	Client           *ethclient.Client
	ServerPrivateKey *ecdsa.PrivateKey
	ContractAddress  *common.Address
	ContractInstance *DeliveryContract
	// where the vendor receives payments and minted delivery tokens
	VendorAddress *common.Address
}

// Creates a new DeliveryContractExecutor that can interact with a delivery contract.
// This will either create a new instance of the contract or use an existing address.
//
// contractAddress - optional. If not given, this will deploy a new instance of the contract.
func NewDeliveryContractExecutor(
	nodeUrl string,
	serverPrivateKey string,
	contractAddress *string,
) (*DeliveryContractExecutor, error) {

	client, err := ethclient.Dial("http://172.13.3.1:8545")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not connect to ethereum node: %v", err))
	}

	privKey, err := crypto.HexToECDSA(serverPrivateKey)
	if err != nil {
		return nil, err
	}

	vendorPublicKey := privKey.Public()
	pubKey := vendorPublicKey.(*ecdsa.PublicKey)
	vendorAddress := crypto.PubkeyToAddress(*pubKey)

	executor := DeliveryContractExecutor{
		Client:           client,
		ServerPrivateKey: privKey,
		VendorAddress:    &vendorAddress,
	}

	// Either look up the existing contract or deploy a new one
	if contractAddress != nil && len(*contractAddress) != 0 {
		addr := common.HexToAddress(*contractAddress)
		contractInstance, err := NewDeliveryContract(addr, client)
		if err != nil {
			return nil, err
		}
		executor.ContractAddress = &addr
		executor.ContractInstance = contractInstance
	} else {
		newAddr, contract, err := executor.deployContract()
		if err != nil {
			return nil, err
		}
		executor.ContractAddress = newAddr
		executor.ContractInstance = contract
	}

	log.Info("Done initializing")
	log.Infof("    Vendor has an address of [%s]", vendorAddress.Hex())
	log.Infof("    Contract deployed with address [%s]", executor.ContractAddress.Hex())

	return &executor, nil
}

func (_exec *DeliveryContractExecutor) deployContract() (*common.Address, *DeliveryContract, error) {
	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce

	contractAddress, tx, tokenContract, err := DeployDeliveryContract(txOpts, _exec.Client)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Error deploying token contract: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to create contract", tx.Hash().Hex())

	err = _exec.waitForMining(tx.Hash(), 30)
	if err != nil {
		return nil, nil, err
	}

	return &contractAddress, tokenContract, nil
}

// returns (tokenId, contract address, error)
func (_exec *DeliveryContractExecutor) MintNFT(purchase *Purchase) (*big.Int, string, error) {
	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return nil, "", err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce
	recipientAddress := common.HexToAddress(purchase.RecipientAddress)

	log.Infof("Minting a token token to be delivered to [%v] for a cost of [%v]",
		recipientAddress.Hex(),
		purchase.PurchasePrice.Int64()+purchase.DeliveryPrice.Int64())

	tx, err := _exec.ContractInstance.MintToken(
		txOpts,
		recipientAddress,
		purchase.DeliveryPrice,
		purchase.PurchasePrice,
		purchase.OrderId)

	if err != nil {
		return nil, "", err
	}

	err = _exec.waitForMining(tx.Hash(), 30)
	if err != nil {
		return nil, "", err
	}

	tokenId, err := _exec.ContractInstance.GetTokenIdForOrder(nil, purchase.OrderId)
	if err != nil {
		return nil, "", err
	}

	return tokenId, _exec.ContractAddress.Hex(), nil
}

// Deposit ether from the customer into the contract
func (_exec *DeliveryContractExecutor) PayForGoods(
	tokenId int64,
	buyerPrivateKey string,
	price int64,
) error {
	privKey, err := crypto.HexToECDSA(buyerPrivateKey)
	if err != nil {
		return err
	}

	txOpts, buyerAddress, err := _exec.buildTxOptsForPrivateKey(privKey)

	// the amount of Ether being sent in the request, in wei
	txOpts.Value = big.NewInt(price)

	_exec.printBalance("customer", buyerAddress)
	_exec.printBalance("vendor", _exec.VendorAddress)
	_exec.printBalance("contract", _exec.ContractAddress)

	tx, err := _exec.ContractInstance.PayForGoods(txOpts, big.NewInt(tokenId))
	if err != nil {
		return errors.New(fmt.Sprintf("Error paying for delivery: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to pay [%d] for the order", tx.Hash().Hex(), price)
	err = _exec.waitForMining(tx.Hash(), 30)
	if err != nil {
		return err
	}

	_exec.printBalance("customer", buyerAddress)
	_exec.printBalance("vendor", _exec.VendorAddress)
	_exec.printBalance("contract", _exec.ContractAddress)
	return nil
}

func (_exec *DeliveryContractExecutor) DeliverOrderAndPayVendor(
	tokenId int64,
	buyerPrivateKey string,
	deliveryPrice int64,
) error {
	err := _exec.transferTokenToCustomer(tokenId, buyerPrivateKey, deliveryPrice)
	if err != nil {
		return err
	}

	err = _exec.payVendor(tokenId)
	if err != nil {
		log.Errorf("Error paying vendor: %v", err.Error())
	}

	return nil
}

// Purches the token from the owner
func (_exec *DeliveryContractExecutor) transferTokenToCustomer(
	tokenId int64,
	buyerPrivateKey string,
	deliveryPrice int64,
) error {

	privKey, err := crypto.HexToECDSA(buyerPrivateKey)
	if err != nil {
		return err
	}

	txOpts, buyerAddress, err := _exec.buildTxOptsForPrivateKey(privKey)

	// the amount of Ether being sent in the request, in wei
	txOpts.Value = big.NewInt(deliveryPrice)

	// I'm using Quorum, which is configured to be gasless, so this doesn't work
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return err
	// }
	// txOpts.GasLimit = uint64(300000) // in gas units
	// txOpts.GasPrice = gasPrice

	// print a balance before and after so we can see that ether was actually transferred
	_exec.printBalance("customer", buyerAddress)
	_exec.printBalance("vendor", _exec.VendorAddress)
	_exec.printBalance("contract", _exec.ContractAddress)

	tx, err := _exec.ContractInstance.Buy(txOpts, big.NewInt(tokenId))
	if err != nil {
		return errors.New(fmt.Sprintf("Error paying for delivery: %v", err))
	}
	log.Infof("Tx sent with ID [%s] to buy token [%d]", tx.Hash().Hex(), tokenId)
	err = _exec.waitForMining(tx.Hash(), 30)
	if err != nil {
		return err
	}

	_exec.printBalance("customer", buyerAddress)
	_exec.printBalance("vendor", _exec.VendorAddress)
	_exec.printBalance("contract", _exec.ContractAddress)
	return nil
}

// transfer the money from the contract to the vendor
func (_exec *DeliveryContractExecutor) payVendor(tokenId int64) error {
	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce

	_exec.printBalance("vendor", _exec.VendorAddress)
	_exec.printBalance("contract", _exec.ContractAddress)
	tx, err := _exec.ContractInstance.Withdraw(txOpts, big.NewInt(tokenId))
	if err != nil {
		return err
	}
	err = _exec.waitForMining(tx.Hash(), 30)
	if err != nil {
		return err
	}

	_exec.printBalance("vendor", _exec.VendorAddress)
	_exec.printBalance("contract", _exec.ContractAddress)
	return nil
}

// Returns address of the the token's current owner
func (_exec *DeliveryContractExecutor) GetOwner(tokenId int64) (string, error) {
	ownerAddress, err := _exec.ContractInstance.OwnerOf(nil, big.NewInt(tokenId))

	if err != nil {
		return "", err
	}
	owner := ownerAddress.Hex()
	return owner, nil
}

// Checks whether the token has been transferred to the customer
func (_exec *DeliveryContractExecutor) IsDelivered(tokenId int64) (bool, error) {
	// the order has been delivered if the token does not reside at the vendor's address

	address, err := _exec.GetOwner(tokenId)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Token [%d] does not exist or has been burned", tokenId))
	}

	return address != _exec.getAddressFromKey(_exec.ServerPrivateKey).Hex(), nil
}

// Destroys the token
func (_exec *DeliveryContractExecutor) BurnDeliveryToken(orderId string) error {

	nonce, err := _exec.getNonce(_exec.ServerPrivateKey)
	if err != nil {
		return err
	}

	txOpts := bind.NewKeyedTransactor(_exec.ServerPrivateKey)
	txOpts.Nonce = nonce

	_, err = _exec.ContractInstance.BurnTokenByOrderId(txOpts, orderId)

	if err != nil {
		log.Errorf("Failed to burn token: %v", err)
		return err
	}

	return nil
}

func (_exec *DeliveryContractExecutor) buildTxOptsForPrivateKey(
	privateKey *ecdsa.PrivateKey,
) (*bind.TransactOpts, *common.Address, error) {
	nonce, err := _exec.getNonce(privateKey)
	if err != nil {
		return nil, nil, err
	}

	txOpts := bind.NewKeyedTransactor(privateKey)
	txOpts.Nonce = nonce

	return txOpts, _exec.getAddressFromKey(privateKey), nil
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

func (_exec *DeliveryContractExecutor) printBalance(label string, address *common.Address) {
	balance, err := _exec.Client.BalanceAt(context.Background(), *address, nil)
	if err != nil {
		log.Info("Could not get balance for [%s]: [%v]", label, err.Error())
		return
	}
	log.Infof("%s has a balance of [%d]", label, balance)
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

// When a transaction is sent to the blockchain, it is pending until it actually gets incorporated into a block.
// By watching the transaction receipt, we can be sure the result of the transaction will be visible in the
// next call.
func (_exec *DeliveryContractExecutor) waitForMining(txHash common.Hash, maxWaitSeconds int) error {
	isMined := false
	startTime := time.Now()
	for !isMined && int(time.Since(startTime).Seconds()) < maxWaitSeconds {
		receipt, err := _exec.Client.TransactionReceipt(context.Background(), txHash)

		isMined = err == nil && receipt != nil && receipt.BlockNumber != nil && receipt.BlockNumber.Uint64() > 0

		if !isMined {
			time.Sleep(2 * time.Second)
		}
	}

	if !isMined {
		return errors.New(fmt.Sprintf("Transaction [%s] was not mined after 30 seconds", txHash.Hex()))
	}

	log.Infof("Tx [%s] was mined", txHash.Hex())
	return nil
}

// There is some functionalilty in the ABI bindings for subscribing to an events channel to watch for events
// emitted from the blockchain. It is not supported with the RPC connection interface I am using, since there
// is no long-lived connection there. I am leaving the code here regardless.
// ----
//
// func waitForMint(contractInstance *DeliveryContract) {
// 	log.Info("Waiting for the emitted event")
//
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go listenForMintedEvent(context.Background(), &wg, contractInstance)
// 	wg.Wait()
// }
//
// func listenForMintedEvent(ctx context.Context, wg *sync.WaitGroup, contract *DeliveryContract) {
// 	defer wg.Done()
//
// 	events := make(chan *DeliveryContractNFTMinted)
//
// 	start := uint64(0)
// 	watchOpts := &bind.WatchOpts{
// 		Start: &start,
// 		Context: ctx,
// 	}
// 	subscription, err := contract.DeliveryContractFilterer.WatchNFTMinted(watchOpts, events)
// 	if err != nil {
// 		log.Errorf("Could not watch for token minting event: %v", err)
// 	}
// 	defer subscription.Unsubscribe()
// }
