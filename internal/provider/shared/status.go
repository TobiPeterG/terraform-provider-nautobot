package shared

import (
	"context"
	"fmt"

	nb "github.com/nautobot/go-nautobot/v3"
)

func GetStatusName(ctx context.Context, client *nb.APIClient, statusID string) (string, error) {
	status, _, err := client.ExtrasAPI.ExtrasStatusesRetrieve(ctx, statusID).Execute()
	if err != nil {
		return "", err
	}
	if status.Name != "" {
		return status.Name, nil
	}
	return "", fmt.Errorf("status name not found for ID %s", statusID)
}

func GetStatusID(ctx context.Context, client *nb.APIClient, statusName string) (string, error) {
	statuses, _, err := client.ExtrasAPI.ExtrasStatusesList(ctx).Name([]string{statusName}).Execute()
	if err != nil {
		return "", err
	}
	if len(statuses.Results) != 1 {
		return "", ExactMatchError("status", fmt.Sprintf("name %q", statusName), len(statuses.Results))
	}
	if statuses.Results[0].Id == nil || *statuses.Results[0].Id == "" {
		return "", fmt.Errorf("status %s returned no id", statusName)
	}
	return *statuses.Results[0].Id, nil
}
