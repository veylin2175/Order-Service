package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"L0/internal/models"
	"L0/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler создает новый хендлер
func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

// GetOrder получает заказ
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := chi.URLParam(r, "orderUID")
	if orderUID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "orderUID is required"})
		return
	}

	order, err := h.service.GetOrder(orderUID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, map[string]string{"error": "handler not found"})
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, order)
}

// CreateOrder создает новый заказ
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.SaveOrder(&order); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders := h.service.GetAllOrders()
	render.JSON(w, r, orders)
}
