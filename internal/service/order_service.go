package service

import (
	"L0/internal/cache"
	"L0/internal/models"
	"L0/internal/storage/postgres"
)

type OrderService struct {
	cache   *cache.OrderCache
	storage postgres.Storage
}

func New(storage postgres.Storage) *OrderService {
	service := &OrderService{
		cache:   cache.New(),
		storage: storage,
	}

	// Preload cache on the start
	_ = service.cache.Preload(func() ([]models.Order, error) {
		return storage.GetAllOrders()
	})

	return service
}

func (s *OrderService) GetOrder(orderUID string) (*models.Order, error) {
	// Checking the cache
	if order, exists := s.cache.Get(orderUID); exists {
		return &order, nil
	}

	// Checking the db if not exists in the cache
	order, err := s.storage.GetOrder(orderUID)
	if err != nil {
		return nil, err
	}

	// Saving to the cache
	s.cache.Set(*order)
	return order, nil
}

func (s *OrderService) SaveOrder(order models.Order) error {
	// Saving to the db
	if err := s.storage.SaveOrder(order); err != nil {
		return err
	}

	// Updating the cache
	s.cache.Set(order)
	return nil
}
