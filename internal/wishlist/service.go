package wishlist

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const DefaultMaxItemsPerWishlist = 20

type Repository interface {
	AddWithinLimit(ctx context.Context, item Item, maxItems int) (added bool, err error)
	Remove(ctx context.Context, customerID CustomerID, productID ProductID) (removed bool, err error)
	Exists(ctx context.Context, customerID CustomerID, productID ProductID) (bool, error)
	FindAllByCustomer(ctx context.Context, customerID CustomerID) ([]Item, error)
}

type Service struct {
	repo     Repository
	maxItems int
	now      func() time.Time
}

func NewService(repo Repository, maxItems int) *Service {
	return &Service{
		repo:     repo,
		maxItems: maxItems,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) AddItem(ctx context.Context, customerID CustomerID, productID ProductID) (Item, error) {
	item, err := NewItem(customerID, productID, s.now())
	if err != nil {
		return Item{}, err
	}

	if _, err := s.repo.AddWithinLimit(ctx, item, s.maxItems); err != nil {
		if errors.Is(err, ErrLimitExceeded) {
			return Item{}, fmt.Errorf("wishlist holds at most %d items: %w", s.maxItems, ErrLimitExceeded)
		}
		return Item{}, err
	}

	return item, nil
}

func (s *Service) RemoveItem(ctx context.Context, customerID CustomerID, productID ProductID) error {
	if err := customerID.validate(); err != nil {
		return err
	}
	if err := productID.validate(); err != nil {
		return err
	}

	removed, err := s.repo.Remove(ctx, customerID, productID)
	if err != nil {
		return err
	}
	if !removed {
		return ErrItemNotFound
	}
	return nil
}

func (s *Service) ContainsItem(ctx context.Context, customerID CustomerID, productID ProductID) (bool, error) {
	if err := customerID.validate(); err != nil {
		return false, err
	}
	if err := productID.validate(); err != nil {
		return false, err
	}
	return s.repo.Exists(ctx, customerID, productID)
}

func (s *Service) ListItems(ctx context.Context, customerID CustomerID) ([]Item, error) {
	if err := customerID.validate(); err != nil {
		return nil, err
	}
	return s.repo.FindAllByCustomer(ctx, customerID)
}
