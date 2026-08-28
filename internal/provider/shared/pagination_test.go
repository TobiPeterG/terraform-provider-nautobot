package shared

import (
	"strings"
	"testing"
)

func TestCollectPages(t *testing.T) {
	t.Parallel()
	var offsets []int32
	items, err := CollectPages(func(limit, offset int32) (Page[int], error) {
		if limit != DataSourcePageSize {
			t.Fatalf("unexpected limit %d", limit)
		}
		offsets = append(offsets, offset)
		if offset == 0 {
			return Page[int]{Items: []int{1, 2}, HasNext: true}, nil
		}
		return Page[int]{Items: []int{3}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || len(offsets) != 2 || offsets[1] != 2 {
		t.Fatalf("unexpected items/offsets: items=%v offsets=%v", items, offsets)
	}
}

func TestCollectPagesRejectsEmptyIntermediatePage(t *testing.T) {
	t.Parallel()

	_, err := CollectPages(func(_, _ int32) (Page[int], error) {
		return Page[int]{HasNext: true}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "next page but returned no items") {
		t.Fatalf("expected an empty intermediate page error, got %v", err)
	}
}
