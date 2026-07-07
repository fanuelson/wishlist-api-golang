package wishlist

import "context"

type WishlistRepository interface {
	CountByCustomer(ctx context.Context, customerID CustomerID) (int, error)
	Insert(ctx context.Context, item Item) (inserted bool, err error)
	Delete(ctx context.Context, customerID CustomerID, productID ProductID) (deleted bool, err error)
	Exists(ctx context.Context, customerID CustomerID, productID ProductID) (bool, error)
	FindAllByCustomer(ctx context.Context, customerID CustomerID) ([]Item, error)
}

type TransactionManager interface {
	WithinCustomerLock(ctx context.Context, customerID CustomerID, fn func(ctx context.Context) error) error
}
