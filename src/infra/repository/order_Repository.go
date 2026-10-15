package repository

import (
	"context"
	"fmt"

	portsout "github.com/DevThiagoSouza/src/application/ports/out"
	"github.com/DevThiagoSouza/src/domain"
	"gorm.io/gorm"
)

type OrderRepository struct {
	DB     *gorm.DB
	Logger portsout.Logger
}

func NewOrderRepository(db *gorm.DB, logger portsout.Logger) *OrderRepository {
	return &OrderRepository{DB: db, Logger: logger}
}

func (r *OrderRepository) logInfo(ctx context.Context, msg string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Info(ctx, msg, args...)
	}
}

func (r *OrderRepository) logError(ctx context.Context, msg string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Error(ctx, msg, args...)
	}
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	r.logInfo(ctx, "saving order", "order_id", order.ID, "customer_code", order.CustomersCode)

	if err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := orderModel{
			ID: order.ID, OrderCode: order.OrderCode, CustomersCode: order.CustomersCode,
		}
		if err := tx.Create(&model).Error; err != nil {
			r.logError(ctx, "failed to insert order", "order_id", order.ID, "error", err)
			return fmt.Errorf("failed to insert order: %w", err)
		}

		for _, item := range order.Items {
			itemModel := toItemModel(&item, order.ID)
			if err := tx.Create(&itemModel).Error; err != nil {
				r.logError(ctx, "failed to insert item", "order_id", order.ID, "item_id", item.ID, "error", err)
				return fmt.Errorf("failed to insert item: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	r.logInfo(ctx, "order saved", "order_id", order.ID)
	return nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	return r.Save(ctx, order)
}

func (r *OrderRepository) FindByCustomerCode(ctx context.Context, customerCode string) ([]domain.Order, error) {
	var models []orderModel
	if err := r.DB.WithContext(ctx).
		Preload("Items").
		Where("customers_code = ?", customerCode).
		Order("id").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}

	orders := make([]domain.Order, 0, len(models))
	for _, model := range models {
		order := domain.Order{
			ID: model.ID, OrderCode: model.OrderCode, CustomersCode: model.CustomersCode,
			Items: make([]domain.Item, 0, len(model.Items)),
		}
		for _, item := range model.Items {
			order.Items = append(order.Items, toDomainItem(item))
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, customerCode string) ([]domain.Order, error) {
	return r.FindByCustomerCode(ctx, customerCode)
}

func (r *OrderRepository) TotalByOrderID(ctx context.Context, orderID string) (float64, error) {
	var total float64
	if err := r.DB.WithContext(ctx).
		Model(&itemModel{}).
		Select("COALESCE(SUM(quantity * price), 0)").
		Where("order_id = ?", orderID).
		Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total: %w", err)
	}
	return total, nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context, orderID string) (float64, error) {
	return r.TotalByOrderID(ctx, orderID)
}

func (r *OrderRepository) CountByCustomerCode(ctx context.Context, customerCode string) (int64, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&orderModel{}).
		Where("customers_code = ?", customerCode).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}
	return count, nil
}
