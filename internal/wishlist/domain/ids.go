package domain

import "github.com/fanuelson/wishlist-api/internal/stringutils"

type CustomerID string

type ProductID string

func (id CustomerID) validate() error {
	return stringutils.RequireNonBlank(string(id), ErrInvalidCustomerID)
}

func (id ProductID) validate() error {
	return stringutils.RequireNonBlank(string(id), ErrInvalidProductID)
}
