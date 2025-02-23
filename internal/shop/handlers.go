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
		s.logger.Error("Auth. Invalid JSON", "error", err)
		me.Error(w, r, "invalid JSON", http.StatusBadRequest)
		return
	}

	if data.Username == "" {
		s.logger.Error("Auth. Empty username")
		me.Error(w, r, "empty username", http.StatusBadRequest)
		return
	}

	if data.Password == "" {
		s.logger.Error("Auth. Empty user password", "user", data.Username)
		me.Error(w, r, "empty password", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	id, ok, err := s.repo.HasUser(ctx, data)
	if err != nil {
		if errors.Is(err, me.ErrInvalidPassword) {
			s.logger.Error("Auth. Invalid password", "user", data.Username)
			me.Error(w, r, "invalid user password", http.StatusBadRequest)
			return
		}

		s.logger.Error("Auth. Check user presence", "user", data.Username, "error", err)
		me.Error(w, r, "failed to check user data", http.StatusInternalServerError)
		return
	}

	if !ok {
		id, err = s.repo.AddUser(ctx, data)
		if err != nil {
			s.logger.Error("Auth. Add a new user", "user", data.Username, "error", err)
			me.Error(w, r, "failed to create a new user", http.StatusInternalServerError)
			return
		}
	}

	token, err := s.auth.CreateToken(id, data.Username)
	if err != nil {
		s.logger.Error("Auth. Create JWT token", "user", data.Username, "error", err)
		me.Error(w, r, "failed to create JWT token", http.StatusInternalServerError)
		return
	}

	s.logger.Info("Auth. OK", "user", data.Username)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, token)
}

func (s *Service) UserInfo(w http.ResponseWriter, r *http.Request) {
	id, err := s.auth.ParseToken(r)
	if err != nil {
		s.logger.Error("User info. Parse token", "error", err)
		me.Error(w, r, "invalid JWT token", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	user, err := s.repo.UserInfo(ctx, id)
	if err != nil {
		s.logger.Error("User info. Call repository", "error", err)
		me.Error(w, r, "failed to get user info", http.StatusInternalServerError)
		return
	}

	s.logger.Info("User info. OK", "user", user)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, user)
}

func (s *Service) BuyProduct(w http.ResponseWriter, r *http.Request) {
	id, err := s.auth.ParseToken(r)
	if err != nil {
		s.logger.Error("Buy product. Parse token", "error", err)
		me.Error(w, r, "invalid JWT token", http.StatusBadRequest)
		return
	}

	item := chi.URLParam(r, "item")

	if item == "" {
		s.logger.Error("Buy product. Empty item", "buyer", id)
		me.Error(w, r, "empty item name", http.StatusBadRequest)
		return
	}

	if !IsProduct(item) {
		s.logger.Error("Buy product. Invalid item", "buyer", id, "item", item)
		me.Error(w, r, "invalid item name", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	err = s.repo.BuyProduct(ctx, id, item)
	if err != nil {
		s.logger.Error("Buy product. Call repository", "buyer", id, "error", err)
		me.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.logger.Info("Buy product. OK", "buyer", id)
	render.Status(r, http.StatusOK)
}

func (s *Service) SendCoin(w http.ResponseWriter, r *http.Request) {
	senderID, err := s.auth.ParseToken(r)
	if err != nil {
		s.logger.Error("Send coin. Parse token", "error", err)
		me.Error(w, r, "invalid JWT token", http.StatusBadRequest)
		return
	}

	var transfer Transfer

	err = render.DecodeJSON(r.Body, &transfer)
	if err != nil {
		s.logger.Error("Send coin. Decode JSON body", "sender", senderID, "error", err)
		me.Error(w, r, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = transfer.Valid()
	if err != nil {
		s.logger.Error("Send coin. Invalid transfer", "sender", senderID, "error", err)
		me.Error(w, r, err, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	ok, err := s.repo.HasUserID(ctx, transfer.Receiver)
	if err != nil || !ok {
		s.logger.Error("Send coin. Get receiver ID",
			"sender", senderID, "receiver", transfer.Receiver, "error", err)
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
			s.logger.Error("Send coin. Low balance", "sender", senderID)
			me.Error(w, r, err, http.StatusBadRequest)
			return
		}

		s.logger.Error("Send coin. Call repository", "sender", senderID, "error", err)
		me.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.logger.Info("Send coin. OK",
		"sender", senderID,
		"receiver", transfer.Receiver,
		"amount", transfer.Amount)

	render.Status(r, http.StatusOK)
}
