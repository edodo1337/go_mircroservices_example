package db

import (
	"context"
	"errors"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/models"
	"wallet_service/internal/pkg/conf"

	"github.com/jackc/pgx/v4"
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

func (dao *PostgresWalletsDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func NewPostgresWalletsDAO(ctx context.Context, config *conf.Config) *PostgresWalletsDAO {
	dbConn := GetPostgresConnection(ctx, config.WalletDatabaseURI())

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
			WHERE id=$2::bigint
			RETURNING id, user_id, balance;`,
	}

	submitPreparedStatements(ctx, queriesMap, dbConn)

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

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, in.ErrTransNotFound
	}

	return &trans, err
}

func (dao *PostgresTransactionsDAO) Create(
	ctx context.Context,
	data *in.CreateWalletTransactionDTO,
) (*models.WalletTransaction, error) {
	var trans models.WalletTransaction
	err := dao.db.QueryRow(
		ctx,
		"create_transaction",
		data.WalletID,
		data.OrderID,
		data.Cost,
		data.Type,
	).Scan(
		&trans.ID,
		&trans.WalletID,
		&trans.OrderID,
		&trans.Cost,
		&trans.Type,
	)

	return &trans, err
}

func (dao *PostgresTransactionsDAO) HealthCheck(ctx context.Context) error {
	if err := dao.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func NewPostgresWalletTransDAO(ctx context.Context, config *conf.Config) *PostgresTransactionsDAO {
	dbConn := GetPostgresConnection(ctx, config.WalletDatabaseURI())

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
			VALUES ($1::bigint, $2::bigint, $3::decimal, $4::smallint)
			RETURNING id, wallet_id, order_id, cost, type;`,
	}

	submitPreparedStatements(ctx, queriesMap, dbConn)

	return &PostgresTransactionsDAO{
		db:                dbConn,
		transactionsTable: config.WalletDatabase.TransactionsTable,
	}
}
