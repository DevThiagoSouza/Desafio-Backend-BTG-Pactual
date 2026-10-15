package test_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DevThiagoSouza/src/application/services"
	"github.com/DevThiagoSouza/src/domain"
)

type fakeOrderRepository struct {
	saveFn                func(context.Context, *domain.Order) error
	findByCustomerCodeFn  func(context.Context, string) ([]domain.Order, error)
	totalByOrderIDFn      func(context.Context, string) (float64, error)
	countByCustomerCodeFn func(context.Context, string) (int64, error)
}

func (f *fakeOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	return f.saveFn(ctx, order)
}

func (f *fakeOrderRepository) FindByCustomerCode(ctx context.Context, customerCode string) ([]domain.Order, error) {
	return f.findByCustomerCodeFn(ctx, customerCode)
}

func (f *fakeOrderRepository) TotalByOrderID(ctx context.Context, orderID string) (float64, error) {
	return f.totalByOrderIDFn(ctx, orderID)
}

func (f *fakeOrderRepository) CountByCustomerCode(ctx context.Context, customerCode string) (int64, error) {
	return f.countByCustomerCodeFn(ctx, customerCode)
}

func validOrder() domain.Order {
	return domain.Order{
		ID: "order-1", OrderCode: "code-1", CustomersCode: "customer-1",
		Items: []domain.Item{{ID: "item-1", Product: "product", Quantity: 1, Price: 10}},
	}
}

func TestOrderServiceCreateOrder(t *testing.T) {
	var saved *domain.Order
	repository := &fakeOrderRepository{
		saveFn: func(_ context.Context, order *domain.Order) error {
			saved = order
			return nil
		},
	}
	service := services.NewOrderService(repository)

	if err := service.CreateOrder(context.Background(), validOrder()); err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if saved == nil || saved.ID != "order-1" {
		t.Fatalf("repository did not receive the expected order: %+v", saved)
	}
}

func TestOrderServiceCreateOrderValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*domain.Order)
		wantErr string
	}{
		{"missing order id", func(order *domain.Order) { order.ID = "" }, "order id is required"},
		{"missing order code", func(order *domain.Order) { order.OrderCode = "" }, "order code is required"},
		{"missing customer code", func(order *domain.Order) { order.CustomersCode = "" }, "customer code is required"},
		{"missing item id", func(order *domain.Order) { order.Items[0].ID = "" }, "item id is required"},
		{"invalid quantity", func(order *domain.Order) { order.Items[0].Quantity = 0 }, "item quantity must be greater than zero"},
		{"negative price", func(order *domain.Order) { order.Items[0].Price = -1 }, "item price cannot be negative"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeOrderRepository{
				saveFn: func(context.Context, *domain.Order) error {
					t.Fatal("Save must not be called for invalid orders")
					return nil
				},
			}
			order := validOrder()
			test.mutate(&order)
			err := services.NewOrderService(repository).CreateOrder(context.Background(), order)
			if err == nil || err.Error() != test.wantErr {
				t.Fatalf("CreateOrder() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestOrderServicePropagatesRepositoryError(t *testing.T) {
	expected := errors.New("repository failed")
	repository := &fakeOrderRepository{
		saveFn: func(context.Context, *domain.Order) error { return expected },
	}

	err := services.NewOrderService(repository).CreateOrder(context.Background(), validOrder())
	if !errors.Is(err, expected) {
		t.Fatalf("CreateOrder() error = %v, want %v", err, expected)
	}
}

func TestOrderServiceQueries(t *testing.T) {
	repository := &fakeOrderRepository{
		findByCustomerCodeFn: func(_ context.Context, code string) ([]domain.Order, error) {
			return []domain.Order{{CustomersCode: code}}, nil
		},
		totalByOrderIDFn: func(_ context.Context, id string) (float64, error) {
			if id != "order-1" {
				t.Fatalf("unexpected order id %q", id)
			}
			return 12.5, nil
		},
		countByCustomerCodeFn: func(_ context.Context, code string) (int64, error) {
			if code != "customer-1" {
				t.Fatalf("unexpected customer code %q", code)
			}
			return 2, nil
		},
	}
	service := services.NewOrderService(repository)

	orders, err := service.GetOrdersByCustomer(context.Background(), "customer-1")
	if err != nil || len(orders) != 1 {
		t.Fatalf("GetOrdersByCustomer() = %v, %v", orders, err)
	}
	total, err := service.GetOrderTotal(context.Background(), "order-1")
	if err != nil || total != 12.5 {
		t.Fatalf("GetOrderTotal() = %v, %v", total, err)
	}
	count, err := service.GetOrderCountByCustomer(context.Background(), "customer-1")
	if err != nil || count != 2 {
		t.Fatalf("GetOrderCountByCustomer() = %v, %v", count, err)
	}
}

func TestOrderServiceRejectsEmptyQueryValues(t *testing.T) {
	repository := &fakeOrderRepository{}
	service := services.NewOrderService(repository)

	if _, err := service.GetOrdersByCustomer(context.Background(), ""); err == nil {
		t.Fatal("GetOrdersByCustomer() should reject an empty customer code")
	}
	if _, err := service.GetOrderTotal(context.Background(), ""); err == nil {
		t.Fatal("GetOrderTotal() should reject an empty order id")
	}
	if _, err := service.GetOrderCountByCustomer(context.Background(), ""); err == nil {
		t.Fatal("GetOrderCountByCustomer() should reject an empty customer code")
	}
}
