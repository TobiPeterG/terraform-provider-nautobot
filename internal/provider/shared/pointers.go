package shared

func StringPtr(value string) *string {
	return &value
}

func Int32Ptr(value int) *int32 {
	converted := int32(value)
	return &converted
}

// DerefString safely dereferences a string pointer, returning an empty string for nil.
func DerefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
