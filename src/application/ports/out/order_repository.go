package out

import (
	"context"

	"github.com/DevThiagoSouza/src/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	FindByCustomerCode(ctx context.Context, customerCode string) ([]domain.Order, error)
	TotalByOrderID(ctx context.Context, orderID string) (float64, error)
	CountByCustomerCode(ctx context.Context, customerCode string) (int64, error)
}
