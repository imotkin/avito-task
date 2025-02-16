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
			Transfer{Sender: uuid.New(), Amount: 100, Receiver: uuid.New()},
			me.ErrInvalidSender,
		},
		{
			Transfer{Receiver: uuid.New(), Amount: 0},
			me.ErrNullAmount,
		},
		{
			Transfer{Receiver: uuid.Nil},
			me.ErrEmptyReceiver,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.transfer.Valid(), tt.err)
		})
	}
}
