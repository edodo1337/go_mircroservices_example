package db

import (
	"context"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/models"
	"wallet_service/internal/pkg/conf"

	"github.com/jackc/pgx/v4/pgxpool"
)

var migrationsApplied bool

// ------------------------------WalletsDAO------------------------------

type PostgresWalletsDAO struct {
	db           *pgxpool.Pool
	walletsTable string
}

func (dao *PostgresWalletsDAO) GetByUserID(ctx context.Context, userID uint) (*models.Wallet, error) {
	var wallet models.Wallet

	err := dao.db.QueryRow(ctx, "get_wallet_by_user_id", userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
	)

	return &wallet, err
}

func (dao *PostgresWalletsDAO) UpdateBalance(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error) {
	var updatedWallet models.Wallet
	err := dao.db.QueryRow(ctx, "update_wallet", wallet.Balance, wallet.ID).Scan(
		&updatedWallet.ID,
		&updatedWallet.UserID,
		&updatedWallet.Balance,
	)

	return &updatedWallet, err
}

func NewPostgresWalletsDAO(ctx context.Context, config *conf.Config) *PostgresWalletsDAO {
	dbConn := GetPostgresConnection(ctx, config.RegistryDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"get_wallet_by_user_id": `SELECT id, user_id, balance 
			FROM wallets WHERE user_id=$1::bigint;`,
		"update_wallet": `UPDATE wallets SET balance=$1::decimal
			WHERE id=$2::bigint;`,
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

	return &PostgresWalletsDAO{
		db:           dbConn,
		walletsTable: config.WalletDatabase.WalletsTable,
	}
}

// ------------------------------TransactionsDAO------------------------------

type PostgresTransactionsDAO struct {
	db                *pgxpool.Pool
	transactionsTable string
}

func (dao *PostgresTransactionsDAO) GetByOrderID(ctx context.Context, orderID uint) (*models.WalletTransaction, error) {
	var trans models.WalletTransaction

	err := dao.db.QueryRow(ctx, "get_transaction_by_order_id", orderID).Scan(
		&trans.ID,
		&trans.WalletID,
		&trans.OrderID,
		&trans.Cost,
		&trans.Type,
	)

	return &trans, err
}

func (dao *PostgresTransactionsDAO) Create(ctx context.Context, transData *in.CreateWalletTransactionDTO) (*models.WalletTransaction, error) {
	var trans models.WalletTransaction
	err := dao.db.QueryRow(
		ctx,
		"create_transaction",
		transData.WalletID,
		transData.OrderID,
		transData.Cost,
		transData.Type,
	).Scan(
		trans.ID,
		trans.WalletID,
		trans.OrderID,
		trans.Cost,
		trans.Type,
	)

	return &trans, err
}

func NewPostgresWalletTransDAO(ctx context.Context, config *conf.Config) *PostgresTransactionsDAO {
	dbConn := GetPostgresConnection(ctx, config.RegistryDatabaseURI())

	if !migrationsApplied {
		err := Migrate(ctx, dbConn)
		if err != nil {
			panic(err)
		}
	}

	defer func() { migrationsApplied = true }()

	queriesMap := map[string]string{
		"get_transaction_by_order_id": `SELECT id, wallet_id, order_id, cost, type 
			FROM wallet_transactions WHERE order_id=$1::bigint;`,
		"create_transaction": `INSERT INTO wallet_transactions(wallet_id, order_id, cost, type) 
			VALUES ($1::bigint, $2::bigint, $3::decimal, $4::smallint);`,
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

	return &PostgresTransactionsDAO{
		db:                dbConn,
		transactionsTable: config.WalletDatabase.TransactionsTable,
	}
}
