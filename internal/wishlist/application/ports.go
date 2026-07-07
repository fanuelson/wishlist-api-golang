package application

import (
	"context"

	"github.com/fanuelson/wishlist-api/internal/wishlist/domain"
)

// IN
type AddProductToWishlistUseCase interface {
	Execute(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) (domain.Item, error)
}

// OUT
type WishlistRepository interface {
	CountByCustomer(ctx context.Context, customerID domain.CustomerID) (int, error)
	Insert(ctx context.Context, item domain.Item) (inserted bool, err error)
	Delete(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) (deleted bool, err error)
	Exists(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) (bool, error)
	FindAllByCustomer(ctx context.Context, customerID domain.CustomerID) ([]domain.Item, error)
}

type TransactionManager interface {
	WithinCustomerLock(ctx context.Context, customerID domain.CustomerID, fn func(ctx context.Context) error) error
}
