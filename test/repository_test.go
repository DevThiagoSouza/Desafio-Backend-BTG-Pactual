package test_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DevThiagoSouza/src/domain"
	"github.com/DevThiagoSouza/src/infra/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockGORM(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	return db, mock
}

func expectInsert(t *testing.T, mock sqlmock.Sqlmock, table string, values int) {
	t.Helper()
	queryArgs := make([]driver.Value, values)
	for index := range queryArgs {
		queryArgs[index] = sqlmock.AnyArg()
	}
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "` + table + `"`)).
		WithArgs(queryArgs...).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func TestItemsRepositorySaveAndFind(t *testing.T) {
	db, mock := newMockGORM(t)
	repositoryUnderTest := repository.NewItemsRepository(db, nil)
	item := domain.Item{ID: "item-1", Product: "book", Quantity: 2, Price: 12.5}

	mock.ExpectBegin()
	expectInsert(t, mock, "items", 5)
	mock.ExpectCommit()
	if err := repositoryUnderTest.Save(context.Background(), &item); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	mock.ExpectQuery(`SELECT .* FROM "items"`).
		WithArgs("item-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "product", "quantity", "price", "order_id"}).
			AddRow("item-1", "book", 2, 12.5, "order-1"))

	items, err := repositoryUnderTest.FindByID(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if len(items) != 1 || items[0] != item {
		t.Fatalf("FindByID() = %+v, want %+v", items, []domain.Item{item})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestItemsRepositorySaveItemsRollsBackOnError(t *testing.T) {
	db, mock := newMockGORM(t)
	repositoryUnderTest := repository.NewItemsRepository(db, nil)
	items := []domain.Item{
		{ID: "item-1", Product: "book", Quantity: 1, Price: 5},
		{ID: "item-2", Product: "pen", Quantity: 1, Price: 2},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "items"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "items"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(assertionError("insert failed"))
	mock.ExpectRollback()

	if err := repositoryUnderTest.SaveItems(context.Background(), items); err == nil {
		t.Fatal("SaveItems() should return the insert error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestItemsRepositoryTotalByOrderID(t *testing.T) {
	db, mock := newMockGORM(t)
	repositoryUnderTest := repository.NewItemsRepository(db, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(quantity * price), 0) FROM "items"`)).
		WithArgs("order-1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(30.5))

	total, err := repositoryUnderTest.TotalByOrderID(context.Background(), "order-1")
	if err != nil || total != 30.5 {
		t.Fatalf("TotalByOrderID() = %v, %v", total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestOrderRepositorySaveAndFind(t *testing.T) {
	db, mock := newMockGORM(t)
	repositoryUnderTest := repository.NewOrderRepository(db, nil)
	order := validOrder()

	mock.ExpectBegin()
	expectInsert(t, mock, "orders", 3)
	expectInsert(t, mock, "items", 5)
	mock.ExpectCommit()
	if err := repositoryUnderTest.Save(context.Background(), &order); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	mock.ExpectQuery(`SELECT .* FROM "orders"`).
		WithArgs("customer-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_code", "customers_code"}).
			AddRow("order-1", "code-1", "customer-1"))
	mock.ExpectQuery(`SELECT .* FROM "items"`).
		WithArgs("order-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "product", "quantity", "price", "order_id"}).
			AddRow("item-1", "product", 1, 10, "order-1"))

	orders, err := repositoryUnderTest.FindByCustomerCode(context.Background(), "customer-1")
	if err != nil {
		t.Fatalf("FindByCustomerCode() error = %v", err)
	}
	if len(orders) != 1 || len(orders[0].Items) != 1 {
		t.Fatalf("FindByCustomerCode() = %+v", orders)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestOrderRepositoryTotalsAndCount(t *testing.T) {
	db, mock := newMockGORM(t)
	repositoryUnderTest := repository.NewOrderRepository(db, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(quantity * price), 0) FROM "items"`)).
		WithArgs("order-1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(10.0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orders"`)).
		WithArgs("customer-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	total, err := repositoryUnderTest.TotalByOrderID(context.Background(), "order-1")
	if err != nil || total != 10 {
		t.Fatalf("TotalByOrderID() = %v, %v", total, err)
	}
	count, err := repositoryUnderTest.CountByCustomerCode(context.Background(), "customer-1")
	if err != nil || count != 1 {
		t.Fatalf("CountByCustomerCode() = %v, %v", count, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

type assertionError string

func (e assertionError) Error() string { return string(e) }
