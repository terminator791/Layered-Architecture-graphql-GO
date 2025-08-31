package http

import (
	"encoding/json"
	"net/http"

	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/business/service"
	"github.com/terminator791/Layered-Architecture-graphql-GO/services/order-service/internal/domain/models"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input models.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.CreateOrder(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if order == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}