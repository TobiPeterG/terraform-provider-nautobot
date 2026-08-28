package shared

import (
	"context"
	"fmt"

	nb "github.com/nautobot/go-nautobot/v3"
)

// StatusResolver loads statuses once for a collection read. This avoids one API
// request per object while still treating missing status references as errors.
type StatusResolver struct {
	names map[string]string
}

func NewStatusResolver(ctx context.Context, client *nb.APIClient) (*StatusResolver, error) {
	names := make(map[string]string)
	statuses, err := CollectPages(func(limit, offset int32) (Page[nb.Status], error) {
		page, response, err := client.ExtrasAPI.ExtrasStatusesList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return Page[nb.Status]{}, fmt.Errorf("list statuses: %s", HTTPError(err, response))
		}
		return Page[nb.Status]{
			Items:   page.Results,
			HasNext: page.Next.IsSet() && page.Next.Get() != nil && *page.Next.Get() != "",
		}, nil
	})
	if err != nil {
		return nil, err
	}
	for _, status := range statuses {
		if status.Id == nil || *status.Id == "" {
			return nil, fmt.Errorf("status list returned an item with no id")
		}
		if status.Name == "" {
			return nil, fmt.Errorf("status %s returned no name", *status.Id)
		}
		names[*status.Id] = status.Name
	}
	return &StatusResolver{names: names}, nil
}

func (r *StatusResolver) Name(id string) (string, error) {
	if id == "" {
		return "", nil
	}
	name, ok := r.names[id]
	if !ok {
		return "", fmt.Errorf("status name not found for ID %s", id)
	}
	return name, nil
}
