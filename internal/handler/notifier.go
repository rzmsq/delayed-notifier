package handler

import (
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func PostNotification(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("body request close error: ", err)
		}
	}()

	var request models.CreateNotificationRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		slog.Error("json decode error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validator.New().Struct(request)
	if err != nil {
		slog.Error("validate error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = service.CreateNotification(&request)
	if err != nil {
		slog.Error("create notification error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetNotify(w http.ResponseWriter, r *http.Request) {}

func DeleteNotify(w http.ResponseWriter, r *http.Request) {}
