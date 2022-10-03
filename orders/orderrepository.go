package orders

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type OrderDB struct {
	url      string
	dbName   string
	username string
	password string
}

type Order struct {
	OrderId      string
	ItemId       string
	ItemName     string
	Price        int64
	TokenAddress string
	TokenId      int64
	Delivered    bool
}

type OrderRepository interface {
	GetOrder(orderId string) (*Order, error)
	CreateOrder(order *Order) (string, error)
	UpdateOrder(orderId string, order *Order) error
}

type MariaDBOrderRepository struct {
	OrderRepository

	host     string
	dbName   string
	username string
	password string
	conn     *sql.DB
}

var ordersTable = "orders"
var allFields = "order_id, item_id, item_name, price, token_address, token_id, delivered"

func NewMariaDBOrderRepository(host string, dbName string, username string, password string) (*MariaDBOrderRepository, error) {
	r := new(MariaDBOrderRepository)
	r.host = host
	r.dbName = dbName
	r.username = username
	r.password = password

	connUrl := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, dbName)

	db, err := sql.Open("mysql", connUrl)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not connect to database %s: %v", dbName, err.Error()))
	}
	r.conn = db
	return r, nil
}

func (repo *MariaDBOrderRepository) GetOrder(orderId string) (*Order, error) {
	// I know this is vulnerable to SQL injection, but it's fine for the demo
	results, err := repo.conn.Query(fmt.Sprintf("select %s from %s where order_id = %s",
		allFields, ordersTable, orderId))
	if err != nil {
		return nil, err
	}

	var order Order
	err = results.Scan(
		&order.OrderId,
		&order.ItemId,
		&order.ItemName,
		&order.Price,
		&order.TokenAddress,
		&order.TokenId,
		&order.Delivered)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (repo *MariaDBOrderRepository) CreateOrder(order *Order) error {
	values := []string{
		fmt.Sprint(order.OrderId),
		order.ItemId,
		order.ItemName,
		fmt.Sprint(order.Price),
		order.TokenAddress,
		fmt.Sprint(order.TokenId),
		func() string {
			if order.Delivered {
				return "1"
			} else {
				return "0"
			}
		}(),
	}

	query := fmt.Sprintf("insert into orders (%s) values('%s')",
		allFields, strings.Join(values, "','"))

	log.Debugf("running query [%s]", query)

	insert, err := repo.conn.Query(query)
	if err != nil {
		log.Debugf("query returned error: %v", err)
		return err
	}
	insert.Close()

	return nil
}
