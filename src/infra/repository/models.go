package repository

type itemModel struct {
	ID       string  `gorm:"column:id;primaryKey"`
	Product  string  `gorm:"column:product;not null"`
	Quantity int     `gorm:"column:quantity;not null"`
	Price    float64 `gorm:"column:price;not null"`
	OrderID  string  `gorm:"column:order_id;not null"`
}

func (itemModel) TableName() string {
	return "items"
}

type orderModel struct {
	ID            string      `gorm:"column:id;primaryKey"`
	OrderCode     string      `gorm:"column:order_code;not null"`
	CustomersCode string      `gorm:"column:customers_code;not null"`
	Items         []itemModel `gorm:"foreignKey:OrderID;references:ID"`
}

func (orderModel) TableName() string {
	return "orders"
}
