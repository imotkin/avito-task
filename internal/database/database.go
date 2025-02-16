package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/imotkin/avito-task/internal/auth"
	"github.com/imotkin/avito-task/internal/config"
	me "github.com/imotkin/avito-task/internal/myerrors"
	"github.com/imotkin/avito-task/internal/shop"
	"golang.org/x/crypto/pbkdf2"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var _ shop.Repository = &DatabaseRepo{}

type DatabaseRepo struct {
	db *sql.DB
}

func New(config *config.Config) (*DatabaseRepo, error) {
	db, err := sql.Open("postgres", config.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DatabaseRepo{db: db}, nil
}

func (r *DatabaseRepo) Connection() *sql.DB {
	return r.db
}

func (r *DatabaseRepo) BuyProduct(ctx context.Context, userID uuid.UUID, item string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	defer tx.Rollback()

	var product shop.Product

	err = tx.QueryRowContext(ctx,
		"SELECT id, price FROM shop.products WHERE title = $1",
		item,
	).Scan(&product.ID, &product.Price)
	if err != nil {
		return fmt.Errorf("get product data: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE shop.users SET coins = coins - $1 WHERE id = $2",
		product.Price, userID)
	if err != nil {
		return fmt.Errorf("update buyer balance: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO shop.inventory (holder, product)
		 VALUES ($1, $2)
		 ON CONFLICT (holder, product) DO UPDATE
         SET amount = shop.inventory.amount + 1`,
		userID, product.ID)
	if err != nil {
		return fmt.Errorf("update buyer inventory: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *DatabaseRepo) SendCoin(ctx context.Context, transfer shop.Transfer) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	defer tx.Rollback()

	var balance uint64

	err = tx.QueryRowContext(ctx,
		"SELECT coins FROM shop.users WHERE id = $1",
		transfer.Sender,
	).Scan(&balance)
	if err != nil {
		return fmt.Errorf("get sender balance: %w", err)
	}

	if balance < transfer.Amount {
		return me.ErrLowBalance
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE shop.users SET coins = coins - $1 WHERE id = $2",
		transfer.Amount, transfer.Sender)
	if err != nil {
		return fmt.Errorf("update sender balance: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE shop.users SET coins = coins + $1 WHERE id = $2",
		transfer.Amount, transfer.Receiver)
	if err != nil {
		return fmt.Errorf("update receiver balance: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO shop.transfers VALUES (DEFAULT, $1, $2, $3)",
		transfer.Sender, transfer.Receiver, transfer.Amount)
	if err != nil {
		return fmt.Errorf("add transfer record: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *DatabaseRepo) UserInfo(ctx context.Context, userID uuid.UUID) (*shop.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT u.coins, 
			COALESCE(
					(SELECT jsonb_agg(jsonb_build_object('type', p.title, 'quantity', i.amount))
					   FROM shop.inventory i
					   JOIN shop.products p ON i.product = p.id
					  WHERE i.holder = u.id), 
					'[]'::jsonb
			) AS inventory, 
			COALESCE(
				(SELECT jsonb_agg(jsonb_build_object('fromUser', sender, 'amount', amount)) 
				   FROM shop.transfers 
				  WHERE receiver = u.id), 
				'[]'::jsonb
			) AS received,
			COALESCE(
				(SELECT jsonb_agg(jsonb_build_object('toUser', receiver, 'amount', amount)) 
				   FROM shop.transfers 
				  WHERE sender = u.id), 
				'[]'::jsonb
			) AS sent
		FROM shop.users u
		WHERE u.id = $1`, userID)

	var user shop.User

	err := row.Scan(
		&user.Coins, &user.Inventory,
		&user.CoinHistory.Received, &user.CoinHistory.Sent)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	return &user, nil
}

func (r *DatabaseRepo) AddUser(ctx context.Context, data auth.LoginData) (uuid.UUID, error) {
	hashedPassword := base64.StdEncoding.EncodeToString(
		pbkdf2.Key([]byte(data.Password), []byte(data.Username), 10000, 32, sha256.New))

	row := r.db.QueryRowContext(ctx,
		`INSERT INTO shop.users (username, hashed_password) 
		 VALUES ($1, $2) RETURNING id`,
		data.Username, hashedPassword)

	var id uuid.UUID

	err := row.Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get created user id: %w", err)
	}

	return id, nil
}

func (r *DatabaseRepo) HasUser(ctx context.Context, data auth.LoginData) (uuid.UUID, bool, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, hashed_password, username 
		   FROM shop.users 
		  WHERE username = $1 LIMIT 1`,
		data.Username)

	var (
		id       uuid.UUID
		password string
		username string
	)

	err := row.Scan(&id, &password, &username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, false, nil
		}

		return uuid.Nil, false, fmt.Errorf("get user id: %w", err)
	}

	input := base64.StdEncoding.EncodeToString(
		pbkdf2.Key([]byte(data.Password), []byte(username), 10000, 32, sha256.New))
	if input != password {
		return uuid.Nil, false, me.ErrInvalidPassword
	}

	return id, true, nil
}

func (r *DatabaseRepo) HasUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 
			  FROM shop.users 
			 WHERE id = $1
		)`, userID)

	var ok bool

	err := row.Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("get user presence: %w", err)
	}

	return ok, nil
}
