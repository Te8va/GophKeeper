package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

//go:generate mockgen -source=savehandler.go -destination=mocks/data_saver_mock.gen.go -package=mocks
type DataSaver interface {
	CreateData(ctx context.Context, userID string, req *domain.DataRecord) error
	UpdateData(ctx context.Context, userID string, req *domain.DataRecord) error
}

type SaverHandler struct {
	saver DataSaver
}

func NewSaverHandler(saver DataSaver) *SaverHandler {
	return &SaverHandler{saver: saver}
}

func (h *SaverHandler) CreateDataHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.UserCtxKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.CreateDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Data is nil", http.StatusBadRequest)
		return
	}

	if req.Metadata == "" {
		http.Error(w, "Metadata is nil", http.StatusBadRequest)
		return
	}

	if !isValidDataType(req.Type) {
		http.Error(w, "Type is wrong", http.StatusBadRequest)
		return
	}

	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	record := &domain.DataRecord{
		Type:     req.Type,
		Metadata: req.Metadata,
		Data:     dataJSON,
	}

	err = h.saver.CreateData(r.Context(), userID, record)
	if err != nil {
		http.Error(w, "Failed to create data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func isValidDataType(dataType domain.DataType) bool {
	switch dataType {
	case domain.TypeLoginPassword, domain.TypeTextData, domain.TypeBinaryData, domain.TypeBankCard:
		return true
	default:
		return false
	}
}

func (h *SaverHandler) UpdateDataHandler(w http.ResponseWriter, r *http.Request) {
	dataID := r.PathValue("id")
	h.UpdateData(w, r, dataID)
}

func (h *SaverHandler) UpdateData(w http.ResponseWriter, r *http.Request, dataID string) {
	userID, ok := r.Context().Value(domain.UserCtxKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if dataID == "" {
		http.Error(w, "Missing or invalid ID", http.StatusBadRequest)
		return
	}

	var req domain.UpdateDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.ID = dataID

	if req.Data == nil {
		http.Error(w, "Data is required", http.StatusBadRequest)
		return
	}

	if req.Metadata == "" {
		http.Error(w, "Metadata is required", http.StatusBadRequest)
		return
	}

	if !isValidDataType(req.Type) {
		http.Error(w, "Invalid data type", http.StatusBadRequest)
		return
	}

	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	record := &domain.DataRecord{
		Type:     req.Type,
		Metadata: req.Metadata,
		Data:     dataJSON,
	}

	err = h.saver.UpdateData(r.Context(), userID, record)
	if err != nil {

		if err == appErrors.ErrNotFoundData {
			http.Error(w, appErrors.ErrNotFoundData.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update data", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
