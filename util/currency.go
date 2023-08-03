package util

// 所有支援的幣種
const (
	USD = "USD"
	EUR = "EUR"
	TWD = "TWD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, TWD:
		return true
	}
	return false
}
