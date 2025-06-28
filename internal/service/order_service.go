package service

import (
	"L0/internal/cache"
	"L0/internal/models"
	"L0/internal/storage/postgres"
	"errors"
	"fmt"
	"log"
)

type OrderService struct {
	cache   *cache.OrderCache
	storage postgres.OrderStorage
}

var (
	ErrOrderNotFound = errors.New("order not found")
)

// New создает новый OrderService с предзагрузкой кэша
func New(storage postgres.OrderStorage, cache *cache.OrderCache) *OrderService {
	service := &OrderService{
		cache:   cache,
		storage: storage,
	}

	// Предзагрузка кэша при старте
	go func() {
		if err := service.preloadCache(); err != nil {
			log.Printf("Cache preload failed: %v", err)
		}
	}()

	return service
}

// preloadCache загружает данные из БД в кэш
func (s *OrderService) preloadCache() error {
	orders, err := s.storage.GetAllOrders()
	if err != nil {
		return fmt.Errorf("get all orders failed: %w", err)
	}

	for _, order := range orders {
		s.cache.Set(order)
	}

	return nil
}

// GetAllOrders возвращает все заказы из кэша
func (s *OrderService) GetAllOrders() []models.Order {
	return s.cache.GetAll()
}

// GetOrder возвращает заказ по ID
func (s *OrderService) GetOrder(orderUID string) (*models.Order, error) {
	if order, exists := s.cache.Get(orderUID); exists {
		return &order, nil
	}

	order, err := s.storage.GetOrder(orderUID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	s.cache.Set(*order)
	return order, nil
}

// SaveOrder сохраняет заказ
func (s *OrderService) SaveOrder(order *models.Order) error {
	if err := s.storage.SaveOrder(*order); err != nil {
		return err
	}
	s.cache.Set(*order)
	return nil
}
