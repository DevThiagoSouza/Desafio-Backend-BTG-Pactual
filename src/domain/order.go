package domain

type Order struct {
	ID            string `json:"id"`
	OrderCode     string `json:"codigoPedido"`
	CustomersCode string `json:"codigoCliente"`
	Items         []Item `json:"itens"`
}

func (o Order) Total() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += float64(item.Quantity) * item.Price
	}
	return total
}