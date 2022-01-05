package db

import (
	"context"
	"errors"
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/models"
	"storage_service/internal/pkg/conf"

	"github.com/lib/pq"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var migrationsApplied bool

// ------------------------------WalletsDAO------------------------------

type PostgresStorageItemsDAO struct {
	db                *pgxpool.Pool
	storageItemsTable string
}

func (dao *PostgresStorageItemsDAO) GetListByProductIDs(ctx context.Context, prodIDs []uint) ([]*models.StorageItem, error) {
	rows, err := dao.db.Query(ctx, "storage_items_list_by_prod_ids", pq.Array(prodIDs))
	if err != nil {
		return nil, err
	}

	items := make([]*models.StorageItem, 0, 10)

	for rows.Next() {
		var item models.StorageItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Count,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

func (dao *PostgresStorageItemsDAO) GetListByOrderID(ctx context.Context, orderID uint) ([]*models.StorageItem, error) {
	rows, err := dao.db.Query(ctx, "storage_items_list_by_order_id", orderID)
	if err != nil {
		return nil, err
	}

	items := make([]*models.StorageItem, 0, 10)

	for rows.Next() {
		var item models.StorageItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Count,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

func (dao *PostgresStorageItemsDAO) UpdateCountBulk(ctx context.Context, items []*models.StorageItem) error {
	tx, err := dao.db.Begin(ctx)
	if err != nil {
		return err
	}

	for _, v := range items {
		if _, err := tx.Exec(ctx, "update_storage_item_count", v.Count, v.ProductID); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (dao *PostgresStorageItemsDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func (dao *PostgresStorageItemsDAO) Close() {
	dao.db.Close()
}

func NewPostgresStorageItemsDAO(ctx context.Context, config *conf.Config) *PostgresStorageItemsDAO {
	dbConn := GetPostgresConnection(ctx, config.WalletDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"storage_items_list_by_prod_ids": `SELECT id, product_id, count 
			FROM storage_items WHERE product_id=ANY($1::bigint[]);`,
		"storage_items_list_by_order_id": `SELECT id, product_id, count 
			FROM storage_items WHERE product_id=ANY($1::bigint[]);`,
		"update_storage_item_count": `UPDATE storage_items 
			SET count=$1::int WHERE product_id=$2::int;`,
	}

	submitPreparedStatements(ctx, queriesMap, dbConn)

	return &PostgresStorageItemsDAO{
		db:                dbConn,
		storageItemsTable: config.StorageDatabase.StorageItemsTable,
	}
}

// ------------------------------TransactionsDAO------------------------------

type PostgresTransactionsDAO struct {
	db                *pgxpool.Pool
	transactionsTable string
}

func (dao *PostgresTransactionsDAO) GetByOrderID(ctx context.Context, orderID uint) (*models.StorageTransaction, error) {
	var trans models.StorageTransaction

	err := dao.db.QueryRow(ctx, "transaction_by_id", orderID).Scan(
		&trans.ID,
		&trans.OrderID,
		&trans.Type,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, in.ErrTransNotFound
	}

	rows, err := dao.db.Query(ctx, "transaction_items_by_trans_id", trans.ID)
	if err != nil {
		return nil, err
	}

	items := make([]*models.StorageTransactionItem, 0, 10)

	for rows.Next() {
		var item models.StorageTransactionItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.TransactionID,
			&item.Count,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	trans.Items = items

	return &trans, rows.Err()
}

func (dao *PostgresTransactionsDAO) GetItemsByOrderID(ctx context.Context, orderID uint) ([]*models.StorageTransactionItem, error) {
	rows, err := dao.db.Query(ctx, "transaction_by_order_id", orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, in.ErrTransNotFound
	}

	items := make([]*models.StorageTransactionItem, 0, 10)

	for rows.Next() {
		var item models.StorageTransactionItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.TransactionID,
			&item.Count,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

func (dao *PostgresTransactionsDAO) Create(
	ctx context.Context,
	data *in.CreateStorageTransactionDTO,
) (*models.StorageTransaction, error) {
	tx, err := dao.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	var trans models.StorageTransaction

	err = tx.QueryRow(
		ctx,
		"create_transaction",
		data.OrderID,
		data.Type,
	).Scan(
		&trans.ID,
		&trans.OrderID,
		&trans.Type,
	)
	if err != nil {
		return nil, err
	}

	items := make([]*models.StorageTransactionItem, 0, 10)

	for _, v := range data.Items {
		var item models.StorageTransactionItem

		err = tx.QueryRow(
			ctx,
			"create_transaction_item",
			v.ProductID,
			trans.ID,
			trans.OrderID,
			v.Count,
		).Scan(
			&item.ID,
			&item.ProductID,
			&item.TransactionID,
			&item.Count,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	errTx := tx.Commit(ctx)
	if errTx != nil {
		return nil, errTx
	}

	trans.Items = items

	return &trans, err
}

func (dao *PostgresTransactionsDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func (dao *PostgresTransactionsDAO) Close() {
	dao.db.Close()
}

func NewPostgresStorageTransDAO(ctx context.Context, config *conf.Config) *PostgresTransactionsDAO {
	dbConn := GetPostgresConnection(ctx, config.WalletDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"transaction_by_id": `SELECT id, order_id, type 
			FROM storage_transactions 
			WHERE id=$1::bigint;`,
		"transaction_items_by_trans_id": `SELECT id, product_id, transaction_id, count 
			FROM storage_transaction_items 
			WHERE transaction_id=$1::bigint;`,
		"transaction_by_order_id": `SELECT id, order_id, type 
			FROM storage_transactions WHERE order_id=$1::bigint;`,
		"create_transaction": `INSERT INTO storage_transactions(order_id, type) VALUES($1::bigint, $2::smallint) RETURNING id, order_id, type;`,
		"create_transaction_item": `INSERT INTO 
			storage_transaction_items(product_id, transaction_id, order_id, count) 
			VALUES($1::int, $2::bigint, $3::bigint, $4::int)
			RETURNING id, product_id, transaction_id, count;`,
	}

	submitPreparedStatements(ctx, queriesMap, dbConn)

	return &PostgresTransactionsDAO{
		db:                dbConn,
		transactionsTable: config.StorageDatabase.TransactionsTable,
	}
}
