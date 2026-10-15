package in

import (
	"context"

	"github.com/DevThiagoSouza/src/domain"
)

type OrderQuery interface {
	GetOrdersByCustomer(ctx context.Context, customerCode string) ([]domain.Order, error)
	GetOrderTotal(ctx context.Context, orderID string) (float64, error)
	GetOrderCountByCustomer(ctx context.Context, customerCode string) (int64, error)
}
