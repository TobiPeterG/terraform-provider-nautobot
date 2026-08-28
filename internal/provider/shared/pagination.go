package shared

import "fmt"

const DataSourcePageSize int32 = 200

type Page[T any] struct {
	Items   []T
	HasNext bool
}

// CollectPages owns offset bookkeeping and termination checks. Endpoint-specific
// request construction and object mapping remain at the call site.
func CollectPages[T any](fetch func(limit, offset int32) (Page[T], error)) ([]T, error) {
	items := make([]T, 0)
	var offset int32
	for {
		page, err := fetch(DataSourcePageSize, offset)
		if err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		if !page.HasNext {
			return items, nil
		}
		if len(page.Items) == 0 {
			return nil, fmt.Errorf("pagination response reported a next page but returned no items at offset %d", offset)
		}
		offset += int32(len(page.Items))
	}
}
