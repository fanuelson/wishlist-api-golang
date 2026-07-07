package wishlist

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	add    *AddProductToWishlistUseCase
	logger *slog.Logger
}

func NewHandler(
	add *AddProductToWishlistUseCase,
	logger *slog.Logger,
) *Handler {
	return &Handler{add: add, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /customers/{customerId}/wishlist/items", h.addItem)
}

type addItemRequest struct {
	ProductID string `json:"productId"`
}

type itemResponse struct {
	CustomerID string `json:"customerId"`
	ProductID  string `json:"productId"`
	AddedAt    string `json:"addedAt"`
}

type listResponse struct {
	CustomerID string         `json:"customerId"`
	Items      []itemResponse `json:"items"`
	Total      int            `json:"total"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) addItem(w http.ResponseWriter, r *http.Request) {
	customerID := CustomerID(r.PathValue("customerId"))

	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.add.Execute(r.Context(), customerID, ProductID(req.ProductID))
	if err != nil {
		h.mapError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, toItemResponse(item))
}

func (h *Handler) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCustomerID), errors.Is(err, ErrInvalidProductID):
		h.writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrItemNotFound):
		h.writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrLimitExceeded):
		h.writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		h.logger.Error("unhandled error serving request", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func toItemResponse(item Item) itemResponse {
	return itemResponse{
		CustomerID: string(item.CustomerID),
		ProductID:  string(item.ProductID),
		AddedAt:    item.AddedAt.Format("2006-01-02T15:04:05.000000Z07:00"),
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, errorResponse{Error: message})
}
