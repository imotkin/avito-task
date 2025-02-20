package shop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/imotkin/avito-task/internal/auth"
	me "github.com/imotkin/avito-task/internal/myerrors"
)

var timeout = time.Second * 3

func (s *Service) Authorize(w http.ResponseWriter, r *http.Request) {
	var data auth.LoginData

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		me.Error(w, r, "invalid JSON", http.StatusBadRequest)
		return
	}

	if data.Username == "" {
		me.Error(w, r, "empty username", http.StatusBadRequest)
		return
	}

	if data.Password == "" {
		me.Error(w, r, "empty password", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	id, ok, err := s.repo.HasUser(ctx, data)
	if err != nil {
		if errors.Is(err, me.ErrInvalidPassword) {
			me.Error(w, r, "invalid user password", http.StatusBadRequest)
			return
		}

		me.Error(w, r, "failed to check user data", http.StatusInternalServerError)
		return
	}

	if !ok {
		id, err = s.repo.AddUser(ctx, data)
		if err != nil {
			me.Error(w, r, "failed to create a new user", http.StatusInternalServerError)
			return
		}
	}

	token, err := s.auth.CreateToken(id, data.Username)
	if err != nil {
		me.Error(w, r, "failed to create JWT token", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, token)
}

func (s *Service) UserInfo(w http.ResponseWriter, r *http.Request) {
	id, err := s.auth.ParseToken(r)
	if err != nil {
		me.Error(w, r, "invalid JWT token", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	user, err := s.repo.UserInfo(ctx, id)
	if err != nil {
		me.Error(w, r, "failed to get user info", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, user)
}

func (s *Service) BuyProduct(w http.ResponseWriter, r *http.Request) {
	item := chi.URLParam(r, "item")

	if item == "" {
		me.Error(w, r, "empty item name", http.StatusBadRequest)
		return
	}

	if !IsProduct(item) {
		me.Error(w, r, "invalid item name", http.StatusBadRequest)
		return
	}

	id, err := s.auth.ParseToken(r)
	if err != nil {
		me.Error(w, r, "invalid JWT token", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	err = s.repo.BuyProduct(ctx, id, item)
	if err != nil {
		me.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

func (s *Service) SendCoin(w http.ResponseWriter, r *http.Request) {
	senderID, err := s.auth.ParseToken(r)
	if err != nil {
		me.Error(w, r, "invalid JWT token", http.StatusBadRequest)
		return
	}

	var transfer Transfer

	err = render.DecodeJSON(r.Body, &transfer)
	if err != nil {
		me.Error(w, r, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = transfer.Valid()
	if err != nil {
		me.Error(w, r, err, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	ok, err := s.repo.HasUserID(ctx, transfer.Receiver)
	if err != nil || !ok {
		me.Error(w, r, "sender is not found", http.StatusBadRequest)
		return
	}

	err = s.repo.SendCoin(ctx, Transfer{
		Sender:   senderID,
		Receiver: transfer.Receiver,
		Amount:   transfer.Amount,
	})
	if err != nil {
		if errors.Is(err, me.ErrLowBalance) {
			me.Error(w, r, err, http.StatusBadRequest)
			return
		}

		me.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}
