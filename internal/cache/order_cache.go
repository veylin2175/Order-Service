package cache

import (
	"L0/internal/models"
	"sync"
)

type OrderCache struct {
	mu    sync.RWMutex
	items map[string]models.Order
}

func New() *OrderCache {
	return &OrderCache{
		items: make(map[string]models.Order),
	}
}

// Set adds a handler to the cache
func (c *OrderCache) Set(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[order.OrderUID] = order
}

// Get returns an handler by its UID
func (c *OrderCache) Get(orderUID string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[orderUID]
	return item, exists
}

// GetAll returns all the orders
func (c *OrderCache) GetAll() []models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	orders := make([]models.Order, 0, len(c.items))
	for _, order := range c.items {
		orders = append(orders, order)
	}

	return orders
}

// Preload loads data from the db on the start
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
