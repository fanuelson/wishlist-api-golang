package wishlist

import "time"

type Item struct {
	CustomerID CustomerID
	ProductID  ProductID
	AddedAt    time.Time
}

func NewItem(customerID CustomerID, productID ProductID, now time.Time) (Item, error) {
	if err := customerID.validate(); err != nil {
		return Item{}, err
	}
	if err := productID.validate(); err != nil {
		return Item{}, err
	}
	return Item{
		CustomerID: customerID,
		ProductID:  productID,
		AddedAt:    now,
	}, nil
}
