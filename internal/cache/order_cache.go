package cache

import (
	"L0/internal/models"
	"sync"
)

type OrderCache struct {
	mu    sync.RWMutex
	items map[string]models.Order
}

// New создает новый объект кеша
func New() *OrderCache {
	return &OrderCache{
		items: make(map[string]models.Order),
	}
}

// Set добавляет хендлер в кеш
func (c *OrderCache) Set(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[order.OrderUID] = order
}

// Get возвращает хендлер по его UID
func (c *OrderCache) Get(orderUID string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[orderUID]
	return item, exists
}

// GetAll возвращает все заказы
func (c *OrderCache) GetAll() []models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	orders := make([]models.Order, 0, len(c.items))
	for _, order := range c.items {
		orders = append(orders, order)
	}

	return orders
}

// Preload выгружает данные из БД при старте
func (c *OrderCache) Preload(loader func() ([]models.Order, error)) error {
	orders, err := loader()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.items[order.OrderUID] = order
	}

	return nil
}
