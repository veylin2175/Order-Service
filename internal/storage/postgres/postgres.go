package postgres

import (
	"L0/internal/config"
	"L0/internal/models"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

type OrderStorage interface {
	SaveOrder(order models.Order) error
	GetOrder(orderUID string) (*models.Order, error)
	GetAllOrders() ([]models.Order, error)
}

// InitDB создает подключение к бд
func InitDB(cfg *config.Config) (*Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("couldn't connect to the database: %v", err)
		return nil, err
	}

	return &Storage{db: db}, nil
}

// SaveOrder сохраняет заказ в БД
func (s *Storage) SaveOrder(order models.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %v", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}(tx)

	_, err = tx.Exec(`
			INSERT INTO orders (
                    order_uid, track_number, entry, locale, internal_signature,
		            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert a new handler: %v", err)
	}

	_, err = tx.Exec(`
			INSERT INTO deliveries (
			        order_uid, name, phone, zip, city, address, region, email
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (order_uid) DO UPDATE SET
					name = EXCLUDED.name,
					phone = EXCLUDED.phone,
					zip = EXCLUDED.zip,
					city = EXCLUDED.city,
					address = EXCLUDED.address,
					region = EXCLUDED.region,
					email = EXCLUDED.email`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert a new delivery: %v", err)
	}

	_, err = tx.Exec(`
			INSERT INTO payments (
			        order_uid, transaction, request_id, currency, provider,
			        amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (order_uid) DO UPDATE SET
					transaction = EXCLUDED.transaction,
					request_id = EXCLUDED.request_id,
					currency = EXCLUDED.currency,
					provider = EXCLUDED.provider,
					amount = EXCLUDED.amount,
					payment_dt = EXCLUDED.payment_dt,
					bank = EXCLUDED.bank,
					delivery_cost = EXCLUDED.delivery_cost,
					goods_total = EXCLUDED.goods_total,
					custom_fee = EXCLUDED.custom_fee`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert a new payment: %v", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`
				INSERT INTO items (
						order_uid, chrt_id, track_number, price, rid,
						name, sale, size, total_price, nm_id, brand, status)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert a new item: %v", err)
		}
	}

	return tx.Commit()
}

// GetOrder получает заказ из БД
func (s *Storage) GetOrder(orderUID string) (*models.Order, error) {
	var order models.Order

	err := s.db.QueryRow(`
			SELECT 
			    	order_uid, track_number, entry, locale, internal_signature,
			    	customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
			FROM orders WHERE order_uid = $1`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get handler: %v", err)
	}

	err = s.db.QueryRow(`
			SELECT
			    	name, phone, zip, city, address, region, email
			FROM deliveries WHERE order_uid = $1`, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %v", err)
	}

	err = s.db.QueryRow(`
			SELECT
			    	transaction, request_id, currency, provider, amount,
			    	payment_dt, bank, delivery_cost, goods_total, custom_fee
			FROM payments WHERE order_uid = $1`, orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %v", err)
	}

	rows, err := s.db.Query(`
			SELECT
			    	chrt_id, track_number, price, rid, name,
			    	sale, size, total_price, nm_id, brand, status
			FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get item: %v", err)
		}
	}

	return &order, nil
}

// GetAllOrders получает список всех заказов из БД
func (s *Storage) GetAllOrders() ([]models.Order, error) {
	rows, err := s.db.Query(`SELECT order_uid FROM orders`)
	if err != nil {
		log.Printf("failed to get handler uid: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var orders []models.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("failed to scan handler uid: %v", err)
		}

		order, err := s.GetOrder(orderUID)
		if err != nil {
			log.Printf("failed to get handler %s: %v", orderUID, err)
			continue
		}
		orders = append(orders, *order)
	}

	return orders, nil
}

// Close закрывает соединение с БД
func (s *Storage) Close() error {
	return s.db.Close()
}
