package graphql

import "testing"

func TestGraphQLResponseHasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "missing value", value: nil, want: false},
		{name: "empty error list", value: []any{}, want: false},
		{name: "error list", value: []any{map[string]any{"message": "invalid field"}}, want: true},
		{name: "unexpected error shape", value: map[string]any{"message": "invalid field"}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := graphQLResponseHasErrors(test.value); got != test.want {
				t.Fatalf("graphQLResponseHasErrors(%v) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}
