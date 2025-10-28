package handler

import (
	"bytes"
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/rabbit"
	"delayed-notifier/internal/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
)

type APIHandler struct {
	Pool        *rabbit.ChannelPool
	RedisClient *redis.Client
}

func (h *APIHandler) PostNotification(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("body request close error")
		}
	}()

	channel := h.Pool.Get()
	defer h.Pool.Put(channel)

	var request models.Notification

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("json decode")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validator.New().Struct(request)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("validate")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	notificationJSON, err := json.Marshal(request)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("json marshal error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if request.SendAt.Before(time.Now()) {
		zlog.Logger.Error().Msg("send_at time is in the past")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	request.ID = uuid.New().String()
	request.Status = models.StatusPending
	err = h.RedisClient.Set(r.Context(), request.ID, notificationJSON)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("store save")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = service.CreateNotification(&request, channel)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("create notification error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Notification created with id " + request.ID))
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("write response")
	}
}

func (h *APIHandler) GetNotify(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	notification, err := h.RedisClient.Get(r.Context(), id)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("store find error")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var notificationModel models.Notification
	err = json.NewDecoder(bytes.NewReader([]byte(notification))).Decode(&notificationModel)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("json decode error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(notificationModel.Status))
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("write response error")
	}
}

func (h *APIHandler) DeleteNotify(w http.ResponseWriter, r *http.Request) {}
