package prefix

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPrefixReferences(t *testing.T) {
	t.Parallel()

	const id = "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	for name, reference := range map[string]any{
		"parent": prefixParentReference(id),
		"RIR":    prefixRIRReference(id),
	} {
		encoded, err := json.Marshal(reference)
		if err != nil {
			t.Fatalf("marshal %s reference: %v", name, err)
		}
		if !strings.Contains(string(encoded), id) {
			t.Errorf("%s reference does not contain %q: %s", name, id, encoded)
		}
	}
}
