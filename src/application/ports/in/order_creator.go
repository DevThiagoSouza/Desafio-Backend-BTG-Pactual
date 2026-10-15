package in

import (
	"context"

	"github.com/DevThiagoSouza/src/domain"
)

type OrderCreator interface {
	CreateOrder(ctx context.Context, order domain.Order) error
}
