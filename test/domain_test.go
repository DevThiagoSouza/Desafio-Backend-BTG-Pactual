package test_test

import (
	"testing"

	"github.com/DevThiagoSouza/src/domain"
)

func TestOrderTotal(t *testing.T) {
	order := domain.Order{
		Items: []domain.Item{
			{Quantity: 2, Price: 10.50},
			{Quantity: 3, Price: 4.25},
		},
	}

	if got, want := order.Total(), 33.75; got != want {
		t.Fatalf("Total() = %v, want %v", got, want)
	}
}

func TestOrderTotalWithoutItems(t *testing.T) {
	if got := (domain.Order{}).Total(); got != 0 {
		t.Fatalf("Total() = %v, want 0", got)
	}
}
