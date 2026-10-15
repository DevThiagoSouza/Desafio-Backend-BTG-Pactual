package services

import (
	"context"
	"errors"

	portsin "github.com/DevThiagoSouza/src/application/ports/in"
	portsout "github.com/DevThiagoSouza/src/application/ports/out"
	"github.com/DevThiagoSouza/src/domain"
)

type OrderService struct {
	orderRepo portsout.OrderRepository
}

var _ portsin.OrderQuery = (*OrderService)(nil)

func NewOrderService(orderRepo portsout.OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

func (s *OrderService) CreateOrder(ctx context.Context, order domain.Order) error {
	if order.ID == "" {
		return errors.New("order id is required")
	}
	if order.OrderCode == "" {
		return errors.New("order code is required")
	}
	if order.CustomersCode == "" {
		return errors.New("customer code is required")
	}

	for _, item := range order.Items {
		if item.ID == "" {
			return errors.New("item id is required")
		}
		if item.Quantity <= 0 {
			return errors.New("item quantity must be greater than zero")
		}
		if item.Price < 0 {
			return errors.New("item price cannot be negative")
		}
	}

	return s.orderRepo.Save(ctx, &order)
}

func (s *OrderService) GetOrdersByCustomer(ctx context.Context, customerCode string) ([]domain.Order, error) {
	if customerCode == "" {
		return nil, errors.New("customer code is required")
	}
	return s.orderRepo.FindByCustomerCode(ctx, customerCode)
}

func (s *OrderService) GetOrderTotal(ctx context.Context, orderID string) (float64, error) {
	if orderID == "" {
		return 0, errors.New("order id is required")
	}
	return s.orderRepo.TotalByOrderID(ctx, orderID)
}

func (s *OrderService) GetOrderCountByCustomer(ctx context.Context, customerCode string) (int64, error) {
	if customerCode == "" {
		return 0, errors.New("customer code is required")
	}
	return s.orderRepo.CountByCustomerCode(ctx, customerCode)
}
