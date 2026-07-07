package domain

var DefaultMaxItemsPerWishlist = 20

type LimitPolicy struct {
	Max int
}

func NewLimitPolicy(max int) LimitPolicy {
	return LimitPolicy{Max: max}
}

func (p LimitPolicy) Allows(currentCount int) bool {
	return currentCount < p.Max
}
