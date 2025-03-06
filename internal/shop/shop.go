package shop

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/imotkin/avito-task/internal/auth"
	me "github.com/imotkin/avito-task/internal/myerrors"
)

type Service struct {
	logger *slog.Logger
	repo   Repository
	auth   auth.Authorizer
}

func NewService(repo Repository, auth auth.Authorizer, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{repo: repo, auth: auth, logger: logger}
}

type Repository interface {
	BuyProduct(ctx context.Context, userID uuid.UUID, item string) error
	SendCoin(ctx context.Context, transfer Transfer) error
	UserInfo(ctx context.Context, userID uuid.UUID) (*User, error)

	AddUser(ctx context.Context, data auth.LoginData) (uuid.UUID, error)
	HasUser(ctx context.Context, data auth.LoginData) (uuid.UUID, bool, error)
	HasUserID(ctx context.Context, userID uuid.UUID) (bool, error)
}

type Product struct {
	ID    int
	Title string
	Price uint64
}

type User struct {
	ID          uuid.UUID       `json:"-"`
	Name        string          `json:"-"`
	Coins       uint64          `json:"coins"`
	Inventory   json.RawMessage `json:"inventory"`
	CoinHistory History         `json:"coinHistory"`
}

type Transfer struct {
	Sender   uuid.UUID `json:"fromUser"`
	Receiver uuid.UUID `json:"toUser"`
	Amount   uint64    `json:"amount"`
}

func (t Transfer) Valid() error {
	switch {
	case t.Receiver == uuid.Nil:
		return me.ErrEmptyReceiver
	case t.Amount == 0:
		return me.ErrNullAmount
	case t.Sender != uuid.Nil:
		return me.ErrInvalidSender
	default:
		return nil
	}
}

type Item struct {
	Type     string `json:"type"`
	Quantity uint64 `json:"quantity"`
}

type History struct {
	Received json.RawMessage `json:"received"`
	Sent     json.RawMessage `json:"sent"`
}
