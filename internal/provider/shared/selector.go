package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

type SelectorField struct {
	Name  string
	Value string
}

// SelectorSpec is the single definition of the fields accepted by a singular
// data source. It supplies both Terraform configuration validation and the
// defensive validation performed before an API request.
type SelectorSpec struct {
	NaturalKey []string
	Qualifiers []string
}

func (s SelectorSpec) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{SelectorConfigValidator{spec: s}}
}

func (s SelectorSpec) Validate(id string, values map[string]string) error {
	naturalKey := make([]SelectorField, 0, len(s.NaturalKey))
	for _, name := range s.NaturalKey {
		naturalKey = append(naturalKey, SelectorField{Name: name, Value: values[name]})
	}
	qualifiers := make([]SelectorField, 0, len(s.Qualifiers))
	for _, name := range s.Qualifiers {
		qualifiers = append(qualifiers, SelectorField{Name: name, Value: values[name]})
	}
	return ValidateIDOrNaturalKey(id, naturalKey, qualifiers...)
}

// ValidateIDOrNaturalKey enforces the selector contract used by singular data
// sources. A lookup is either by ID, or by every field in the natural key.
// Qualifiers are optional parts of a natural-key lookup (for example a nullable
// tenant), but they may not be combined with an ID lookup.
func ValidateIDOrNaturalKey(id string, naturalKey []SelectorField, qualifiers ...SelectorField) error {
	allNaturalFields := append(append([]SelectorField{}, naturalKey...), qualifiers...)
	providedNaturalField := false
	for _, field := range allNaturalFields {
		if field.Value != "" {
			providedNaturalField = true
			break
		}
	}

	selectorNames := make([]string, 0, len(naturalKey))
	missingNames := make([]string, 0, len(naturalKey))
	for _, field := range naturalKey {
		selectorNames = append(selectorNames, "`"+field.Name+"`")
		if field.Value == "" {
			missingNames = append(missingNames, "`"+field.Name+"`")
		}
	}
	selectorDescription := strings.Join(selectorNames, " and ")

	if id != "" {
		if providedNaturalField {
			return fmt.Errorf("`id` cannot be combined with natural-key fields; provide either `id`, or %s", selectorDescription)
		}
		return nil
	}
	if len(missingNames) > 0 {
		if !providedNaturalField {
			return fmt.Errorf("provide either `id`, or %s", selectorDescription)
		}
		return fmt.Errorf("natural-key fields must be provided together; missing %s", strings.Join(missingNames, " and "))
	}

	return nil
}

// ExactMatchError prevents singular data sources from silently selecting the
// first item returned by a list endpoint.
func ExactMatchError(objectName, selectorDescription string, count int) error {
	switch count {
	case 0:
		return fmt.Errorf("no %s found matching %s", objectName, selectorDescription)
	case 1:
		return nil
	default:
		return fmt.Errorf("expected exactly one %s matching %s, found %d", objectName, selectorDescription, count)
	}
}
