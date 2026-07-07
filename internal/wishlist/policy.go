package wishlist

type LimitPolicy struct {
	max int
}

func NewLimitPolicy(max int) LimitPolicy {
	return LimitPolicy{max: max}
}

func (p LimitPolicy) Allows(currentCount int) bool {
	return currentCount < p.max
}
