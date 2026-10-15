package repository

import (
	"context"
	"fmt"

	portsout "github.com/DevThiagoSouza/src/application/ports/out"
	"github.com/DevThiagoSouza/src/domain"
	"gorm.io/gorm"
)

type ItemsRepository struct {
	DB     *gorm.DB
	Logger portsout.Logger
}

func NewItemsRepository(db *gorm.DB, logger portsout.Logger) *ItemsRepository {
	return &ItemsRepository{DB: db, Logger: logger}
}

func (r *ItemsRepository) logInfo(ctx context.Context, msg string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Info(ctx, msg, args...)
	}
}

func (r *ItemsRepository) logError(ctx context.Context, msg string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Error(ctx, msg, args...)
	}
}

func (r *ItemsRepository) Save(ctx context.Context, item *domain.Item) error {
	r.logInfo(ctx, "saving item", "item_id", item.ID, "product", item.Product)

	if err := r.DB.WithContext(ctx).Create(toItemModel(item, "")).Error; err != nil {
		r.logError(ctx, "failed to save item", "error", err)
		return fmt.Errorf("failed to save item: %w", err)
	}
	return nil
}

func (r *ItemsRepository) SaveItems(ctx context.Context, items []domain.Item) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Create(toItemModel(&item, "")).Error; err != nil {
				return fmt.Errorf("failed to save item in transaction: %w", err)
			}
		}
		return nil
	})
}

func (r *ItemsRepository) FindByID(ctx context.Context, id string) ([]domain.Item, error) {
	var models []itemModel
	if err := r.DB.WithContext(ctx).Where("id = ?", id).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	items := make([]domain.Item, 0, len(models))
	for _, model := range models {
		items = append(items, toDomainItem(model))
	}
	return items, nil
}

func (r *ItemsRepository) TotalByOrderID(ctx context.Context, orderID string) (float64, error) {
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

func toItemModel(item *domain.Item, orderID string) itemModel {
	return itemModel{
		ID: item.ID, Product: item.Product, Quantity: item.Quantity,
		Price: item.Price, OrderID: orderID,
	}
}

func toDomainItem(model itemModel) domain.Item {
	return domain.Item{
		ID: model.ID, Product: model.Product, Quantity: model.Quantity, Price: model.Price,
	}
}
