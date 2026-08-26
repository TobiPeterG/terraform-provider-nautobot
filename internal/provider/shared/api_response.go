package shared

import "fmt"

// ValidateAPIObjectID protects resource reads from malformed responses and
// from accidentally writing state for a different object than was requested.
func ValidateAPIObjectID(objectName, requestedID string, returnedID *string) error {
	if returnedID == nil || *returnedID == "" {
		return fmt.Errorf("%s response has no id", objectName)
	}
	if requestedID != "" && *returnedID != requestedID {
		return fmt.Errorf("%s response returned id %q while %q was requested", objectName, *returnedID, requestedID)
	}
	return nil
}

func ValidateReturnedObjectID(objectName, requestedID, returnedID string) error {
	return ValidateAPIObjectID(objectName, requestedID, &returnedID)
}
