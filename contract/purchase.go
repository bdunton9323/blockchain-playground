package contract

import "math/big"

type Purchase struct {
	OrderId       string
	PurchasePrice *big.Int
	DeliveryPrice *big.Int
	// the ethereum address of the customer receiving the order, as a hex string
	RecipientAddress string
}
