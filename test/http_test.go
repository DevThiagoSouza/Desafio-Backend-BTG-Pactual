package test_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DevThiagoSouza/src/domain"
	httpapi "github.com/DevThiagoSouza/src/infra/http"
	"github.com/gin-gonic/gin"
)

type fakeOrderAPI struct {
	order  domain.Order
	orders []domain.Order
	err    error
}

func (f *fakeOrderAPI) CreateOrder(_ context.Context, order domain.Order) error {
	f.order = order
	return f.err
}

func (f *fakeOrderAPI) GetOrdersByCustomer(context.Context, string) ([]domain.Order, error) {
	return f.orders, f.err
}

func (f *fakeOrderAPI) GetOrderTotal(context.Context, string) (float64, error) {
	return 25.5, f.err
}

func (f *fakeOrderAPI) GetOrderCountByCustomer(context.Context, string) (int64, error) {
	return 3, f.err
}

func newTestRouter(api *fakeOrderAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return httpapi.NewRouter(httpapi.NewOrderHandler(api, api))
}

func TestCreateOrderHandler(t *testing.T) {
	api := &fakeOrderAPI{}
	body, _ := json.Marshal(validOrder())
	request := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newTestRouter(api).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if api.order.ID != "order-1" {
		t.Fatalf("handler sent unexpected order: %+v", api.order)
	}
}

func TestCreateOrderHandlerRejectsInvalidJSON(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("{"))
	request.Header.Set("Content-Type", "application/json")

	newTestRouter(&fakeOrderAPI{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestOrderQueryHandlers(t *testing.T) {
	api := &fakeOrderAPI{orders: []domain.Order{validOrder()}}
	tests := []struct {
		name string
		path string
		want int
	}{
		{"orders", "/customers/customer-1/orders", http.StatusOK},
		{"total", "/orders/order-1/total", http.StatusOK},
		{"count", "/customers/customer-1/orders/count", http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			newTestRouter(api).ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestHandlersReturnBadRequestWhenApplicationFails(t *testing.T) {
	api := &fakeOrderAPI{err: errors.New("application failed")}
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create", http.MethodPost, "/orders", `{"id":"order-1"}`},
		{"orders", http.MethodGet, "/customers/customer-1/orders", ""},
		{"total", http.MethodGet, "/orders/order-1/total", ""},
		{"count", http.MethodGet, "/customers/customer-1/orders/count", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			if test.method == http.MethodPost {
				request.Header.Set("Content-Type", "application/json")
			}

			newTestRouter(api).ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}
