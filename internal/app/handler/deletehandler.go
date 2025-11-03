package handler

import (
	"context"
	"net/http"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

//go:generate mockgen -source=deletehandler.go -destination=mocks/data_delete_mock.gen.go -package=mocks
type DataDelete interface {
	DeleteData(ctx context.Context, dataID, userID string) error
}

type DeleteHandler struct {
	deleter DataDelete
}

func NewDeleteHandler(deleter DataDelete) *DeleteHandler {
	return &DeleteHandler{deleter: deleter}
}

func (h *DeleteHandler) DeleteDataHandler(w http.ResponseWriter, r *http.Request) {
	dataID := r.PathValue("id")
	h.DeleteData(w, r, dataID)
}

func (h *DeleteHandler) DeleteData(w http.ResponseWriter, r *http.Request, dataID string) {
	userID, ok := r.Context().Value(domain.UserCtxKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if dataID == "" {
		http.Error(w, appErrors.ErrInvalidId.Error(), http.StatusBadRequest)
		return
	}

	if err := h.deleter.DeleteData(r.Context(), dataID, userID); err != nil {
		if err == appErrors.ErrNotFoundData {
			http.Error(w, appErrors.ErrNotFoundData.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete data", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
