package db

import (
	"context"
	"fmt"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/app/models"
	"registry_service/internal/pkg/conf"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

var migrationsApplied bool

// ------------------------------OrdersDAO------------------------------

type PostgresOrdersDAO struct {
	db          *pgxpool.Pool
	ordersTable string
}

func (dao *PostgresOrdersDAO) CreateOrder(ctx context.Context, data *in.CreateOrderDTO) (*models.Order, error) {
	var order models.Order

	err := dao.db.QueryRow(ctx, "create_order", data.UserID, models.Pending).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.RejectedReason,
		&order.CreatedAt,
	)

	return &order, err
}

func (dao *PostgresOrdersDAO) GetOrdersListByUserID(ctx context.Context, userID uint) ([]*models.Order, error) {
	rows, err := dao.db.Query(ctx, "orders_list_by_user_id", userID)
	if err != nil {
		return nil, err
	}

	orders := make([]*models.Order, 0, 10)

	for rows.Next() {
		var order models.Order

		err = rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.RejectedReason,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, rows.Err()
}

func (dao *PostgresOrdersDAO) GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error) {
	panic("not implemented")
}

func (dao *PostgresOrdersDAO) DeleteOrder(ctx context.Context, orderID uint) error {
	_, err := dao.db.Exec(ctx, "delete_order", orderID)

	return err
}

func (dao *PostgresOrdersDAO) UpdateOrderStatus(
	ctx context.Context,
	orderID uint,
	status models.OrderStatus,
	reasonCode uint8,
) (*models.Order, error) {
	var order models.Order

	err := dao.db.QueryRow(ctx, "update_order_status", status, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.RejectedReason,
		&order.CreatedAt,
	)

	return &order, err
}

func (dao *PostgresOrdersDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func NewPostgresOrdersDAO(ctx context.Context, config *conf.Config) *PostgresOrdersDAO {
	dbConn := GetPostgresConnection(ctx, config.RegistryDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"create_order": `INSERT INTO orders(user_id, status) 
			VALUES($1::bigint, $2::smallint) 
			RETURNING id, user_id, status, rejected_reason, created_at;`,
		"orders_list_by_user_id": `SELECT id, user_id, status, 
			rejected_reason, created_at FROM orders WHERE user_id=$1::bigint;`,
		"update_order_status": `UPDATE orders SET status=$1::smallint 
			WHERE id=$2::bigint 
			returning id, user_id, status, rejected_reason, created_at;`,
		"delete_order": `DELETE FROM orders WHERE id=$1::bigint;`,
	}

	tx, err := dbConn.Begin(ctx)
	if err != nil {
		panic(err)
	}

	for k, v := range queriesMap {
		if _, err := tx.Prepare(ctx, k, v); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	return &PostgresOrdersDAO{
		db:          dbConn,
		ordersTable: config.RegistryDatabase.OrdersTable,
	}
}

// ---------------------------- OrderItemsDAO----------------------------

type PostgresOrderItemsDAO struct {
	db              *pgxpool.Pool
	orderItemsTable string
}

func (dao *PostgresOrderItemsDAO) CreateOrderItemsBulk(
	ctx context.Context,
	orderID uint,
	items []*in.CreateOrderItemDTO,
) ([]*models.OrderItem, error) {
	var query strings.Builder

	query.WriteString(`INSERT INTO order_items(order_id, product_id, count, product_price) VALUES `)

	for ind, v := range items {
		query.WriteString(
			fmt.Sprintf(
				"(%v::bigint, %v::bigint, %v::smallint, %v::decimal) ",
				orderID, v.ProductID, v.Count, v.ProductPrice,
			),
		)

		if ind != len(items)-1 {
			query.WriteString(",")
		}
	}

	query.WriteString(" RETURNING id, order_id, product_id, count, product_price;")

	fmt.Println("query", query.String())

	rows, err := dao.db.Query(ctx, query.String())
	if err != nil {
		return nil, err
	}

	orderItems := make([]*models.OrderItem, 0, 10)

	for rows.Next() {
		var orderItem models.OrderItem

		err = rows.Scan(
			&orderItem.ID,
			&orderItem.OrderID,
			&orderItem.ProductID,
			&orderItem.Count,
			&orderItem.ProductPrice,
		)
		if err != nil {
			return nil, err
		}

		orderItems = append(orderItems, &orderItem)
	}

	return orderItems, rows.Err()
}

func (dao *PostgresOrderItemsDAO) CreateOrderItem(
	ctx context.Context,
	orderID uint,
	data *in.CreateOrderItemDTO,
) (*models.OrderItem, error) {
	panic("not implemented")
}

func (dao *PostgresOrderItemsDAO) GetOrderItemByID(ctx context.Context, orderItemID uint) (*models.OrderItem, error) {
	panic("not implemented")
}

func (dao *PostgresOrderItemsDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func NewPostgresOrderItemsDAO(ctx context.Context, config *conf.Config) *PostgresOrderItemsDAO {
	dbConn := GetPostgresConnection(ctx, config.RegistryDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"orders_list_by_user_id": `SELECT id, user_id, status, 
			rejected_reason, created_at FROM orders WHERE user_id=$1::bigint;`,
		"update_order_status": `UPDATE orders SET status=$1::smallint 
			WHERE id=$2::bigint 
			returning id, user_id, status, rejected_reason, created_at;`,
		"delete_order": `DELETE FROM orders WHERE id=$1::bigint;`,
	}

	tx, err := dbConn.Begin(ctx)
	if err != nil {
		panic(err)
	}

	for k, v := range queriesMap {
		if _, err := tx.Prepare(ctx, k, v); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	return &PostgresOrderItemsDAO{
		db:              dbConn,
		orderItemsTable: config.RegistryDatabase.OrderItemsTable,
	}
}

// ---------------------------- ProductPricesDAO----------------------------

type PostgresProductPricesDAO struct {
	db            *pgxpool.Pool
	productsTable string
}

func (dao *PostgresProductPricesDAO) GetProductPricesMap(ctx context.Context, productIDs []uint) (in.ProductPricesMap, error) {
	if len(productIDs) == 0 {
		return nil, in.ErrEmptyProductIDs
	}

	pricesMap := make(map[uint]float32)

	rows, err := dao.db.Query(ctx, "products_list")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var product models.Product

		err = rows.Scan(
			&product.ID,
			&product.Title,
			&product.Price,
		)
		if err != nil {
			return nil, err
		}

		pricesMap[product.ID] = product.Price
	}

	return pricesMap, rows.Err()
}

func (dao *PostgresProductPricesDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func NewPostgresProductPricesDAO(ctx context.Context, config *conf.Config) *PostgresProductPricesDAO {
	dbConn := GetPostgresConnection(ctx, config.RegistryDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"products_list": `SELECT id, title, price FROM products;`,
	}

	tx, err := dbConn.Begin(ctx)
	if err != nil {
		panic(err)
	}

	for k, v := range queriesMap {
		if _, err := tx.Prepare(ctx, k, v); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	return &PostgresProductPricesDAO{
		db:            dbConn,
		productsTable: config.RegistryDatabase.ProductsTable,
	}
}
