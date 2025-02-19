package shop

import (
	"testing"

	"github.com/google/uuid"
	me "github.com/imotkin/avito-task/internal/myerrors"
	"github.com/stretchr/testify/assert"
)

func TestValidTransfer(t *testing.T) {
	tests := []struct {
		transfer Transfer
		err      error
	}{
		{
			transfer: Transfer{Sender: uuid.New(), Amount: 100, Receiver: uuid.New()},
			err:      me.ErrInvalidSender,
		},
		{
			transfer: Transfer{Receiver: uuid.New(), Amount: 0},
			err:      me.ErrNullAmount,
		},
		{
			transfer: Transfer{Receiver: uuid.Nil},
			err:      me.ErrEmptyReceiver,
		},
		{
			transfer: Transfer{Receiver: uuid.New(), Amount: 100},
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.transfer.Valid(), tt.err)
		})
	}
}
