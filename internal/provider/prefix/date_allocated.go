package prefix

import (
	"time"

	nb "github.com/nautobot/go-nautobot/v3"
)

func prefixDateAllocated(value string) (nb.NullableTime, error) {
	var nullableTime nb.NullableTime
	if value == "" {
		nullableTime.Set(nil)
		return nullableTime, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nb.NullableTime{}, err
	}
	nullableTime.Set(&parsed)
	return nullableTime, nil
}
