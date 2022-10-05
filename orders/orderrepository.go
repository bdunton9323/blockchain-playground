package orders

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

// A DTO object representing a row in the database
type Order struct {
	OrderId       string
	ItemId        string
	ItemName      string
	Price         int64
	DeliveryPrice int64
	TokenAddress  string
	TokenId       int64
	Delivered     bool
}

type OrderRepository interface {
	GetOrder(orderId string) (*Order, error)
	CreateOrder(order *Order) (string, error)
	MarkOrderDelivered(orderId string) error
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
var allFields = "order_id, item_id, item_name, price, delivery_price, token_address, token_id, delivered"

// Construct a new repository connected to MariaDB
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

// Returns the order with the given ID from the database. If not found, then nil.
func (repo *MariaDBOrderRepository) GetOrder(orderId string) (*Order, error) {
	// I know this is vulnerable to SQL injection, but it's fine for a demo
	query := fmt.Sprintf("select %s from %s where order_id = '%s'",
		allFields, ordersTable, orderId)

	results, err := repo.runQuery(query)
	if err != nil {
		return nil, err
	}

	if !results.Next() {
		return nil, nil
	}

	var order Order
	err = results.Scan(
		&order.OrderId,
		&order.ItemId,
		&order.ItemName,
		&order.Price,
		&order.DeliveryPrice,
		&order.TokenAddress,
		&order.TokenId,
		&order.Delivered)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// Writes the given order to the database.
func (repo *MariaDBOrderRepository) CreateOrder(order *Order) error {
	values := []string{
		fmt.Sprint(order.OrderId),
		order.ItemId,
		order.ItemName,
		fmt.Sprint(order.Price),
		fmt.Sprint(order.DeliveryPrice),
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

	insert, err := repo.runQuery(query)
	if err != nil {
		log.Errorf("query returned error: %v", err)
		return err
	}
	insert.Close()

	return nil
}

// Sets the 'delivered' field for the given order ID. Returns an error if the order doesn't exist.
func (repo *MariaDBOrderRepository) MarkOrderDelivered(orderId string) error {
	query := fmt.Sprintf("update orders set delivered = '1' where order_id = '%s'", orderId)
	_, err := repo.runQuery(query)
	if err != nil {
		return err
	}
	return nil
}

// Runs the given query against the database
func (repo *MariaDBOrderRepository) runQuery(query string) (*sql.Rows, error) {
	log.Infof("running query [%s]", query)
	return repo.conn.Query(query)
}
