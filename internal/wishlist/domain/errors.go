package domain

import "errors"

var (
	ErrInvalidCustomerID = errors.New("customer id must not be empty")
	ErrInvalidProductID  = errors.New("product id must not be empty")
	ErrItemNotFound      = errors.New("product not found in wishlist")
	ErrLimitExceeded     = errors.New("wishlist limit items exceeded")
)
