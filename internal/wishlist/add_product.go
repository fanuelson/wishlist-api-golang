package wishlist

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type AddProductToWishlistUseCase struct {
	repo   WishlistRepository
	tx     TransactionManager
	policy LimitPolicy
	now    func() time.Time
}

func NewAddProductToWishlistUseCase(repo WishlistRepository, tx TransactionManager, policy LimitPolicy) *AddProductToWishlistUseCase {
	return &AddProductToWishlistUseCase{
		repo:   repo,
		tx:     tx,
		policy: policy,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc *AddProductToWishlistUseCase) Execute(ctx context.Context, customerID CustomerID, productID ProductID) (Item, error) {
	item, err := NewItem(customerID, productID, uc.now())
	if err != nil {
		return Item{}, err
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
			return ErrLimitExceeded
		}

		if _, err := uc.repo.Insert(ctx, item); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrLimitExceeded) {
			return Item{}, fmt.Errorf("wishlist holds at most %d items: %w", uc.policy.max, ErrLimitExceeded)
		}
		return Item{}, err
	}

	return item, nil
}
