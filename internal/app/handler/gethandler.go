package handler

import (
	"context"
	"encoding/json"
	"net/http"

	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"

	"github.com/Te8va/GophKeeper/internal/app/domain"
)

//go:generate mockgen -source=gethandler.go -destination=mocks/data_getter_mock.gen.go -package=mocks
type DataGetter interface {
	GetDataByID(ctx context.Context, dataID, userID string) (*domain.DataRecord, error)
	GetAllUserData(ctx context.Context, userID string) ([]*domain.DataRecord, error)
}

type GetterHandler struct {
	getter DataGetter
}

func NewGetterHandler(getter DataGetter) *GetterHandler {
	return &GetterHandler{getter: getter}
}

func (h *GetterHandler) GetDataHandler(w http.ResponseWriter, r *http.Request) {
	dataID := r.PathValue("id")
	h.GetData(w, r, dataID)
}

func (h *GetterHandler) GetData(w http.ResponseWriter, r *http.Request, dataID string) {
	userID, ok := r.Context().Value(domain.UserCtxKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if dataID == "" {
		http.Error(w, "Missing or invalid ID", http.StatusBadRequest)
		return
	}

	record, err := h.getter.GetDataByID(r.Context(), dataID, userID)
	if err != nil {
		http.Error(w, appErrors.ErrNotFoundData.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)
}

func (h *GetterHandler) GetAllDataHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.UserCtxKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	records, err := h.getter.GetAllUserData(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(records)
}
