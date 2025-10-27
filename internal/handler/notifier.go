package handler

import (
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

type APIHandler struct {
	Connection *rabbitmq.Connection
}

func (h *APIHandler) PostNotification(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("body request close error")
		}
	}()

	var request models.CreateNotificationRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("json decode error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validator.New().Struct(request)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("validate error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = service.CreateNotification(&request, h.Connection)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("create notification error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *APIHandler) GetNotify(w http.ResponseWriter, r *http.Request) {}

func (h *APIHandler) DeleteNotify(w http.ResponseWriter, r *http.Request) {}
