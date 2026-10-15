package domain

type Item struct {
	ID       string  `json:"id"`
	Product  string  `json:"produto"`
	Quantity int     `json:"quantidade"`
	Price    float64 `json:"preco"`
}