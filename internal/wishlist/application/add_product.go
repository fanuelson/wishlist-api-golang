package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fanuelson/wishlist-api/internal/wishlist/domain"
)

type AddProductToWishlistUseCaseImpl struct {
	repo   WishlistRepository
	tx     TransactionManager
	policy domain.LimitPolicy
	now    func() time.Time
}

func NewAddProductToWishlistUseCase(repo WishlistRepository, tx TransactionManager, policy domain.LimitPolicy) AddProductToWishlistUseCase {
	return &AddProductToWishlistUseCaseImpl{
		repo:   repo,
		tx:     tx,
		policy: policy,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc AddProductToWishlistUseCaseImpl) Execute(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) (domain.Item, error) {
	item, err := domain.NewItem(customerID, productID, uc.now())
	if err != nil {
		return domain.Item{}, err
	}

	err = uc.tx.WithinCustomerLock(ctx, customerID, func(ctx context.Context) error {
		exists, err := uc.repo.Exists(ctx, customerID, productID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		count, err := uc.repo.CountByCustomer(ctx, customerID)
		if err != nil {
			return err
		}
		if !uc.policy.Allows(count) {
			return domain.ErrLimitExceeded
		}

		if _, err := uc.repo.Insert(ctx, item); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrLimitExceeded) {
			return domain.Item{}, fmt.Errorf("wishlist holds at most %d items: %w", uc.policy.Max, domain.ErrLimitExceeded)
		}
		return domain.Item{}, err
	}

	return item, nil
}
