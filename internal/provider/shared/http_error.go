package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

func IsNotFoundResponse(resp *http.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusNotFound
}

// HTTPErrorAsError adapts HTTPError for helpers whose callback contract
// requires an error instead of a diagnostic detail string.
func HTTPErrorAsError(err error, resp *http.Response) error {
	return fmt.Errorf("%s", HTTPError(err, resp))
}

// HTTPError returns exactly two lines when possible:
// 1) "<status>"
// 2) "<message> (<METHOD> <URL>)"
func HTTPError(err error, resp *http.Response) string {
	var statusLine string
	if resp != nil && resp.Status != "" {
		statusLine = resp.Status
	}
	if statusLine == "" && err != nil {
		statusLine = err.Error()
	}

	var reqInfo string
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		reqInfo = fmt.Sprintf("%s %s", resp.Request.Method, resp.Request.URL.String())
	}

	// Try to extract a clean human message from the body
	var message string
	if resp != nil && resp.Body != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) > 0 {
			var anyJSON any
			if json.Unmarshal(bodyBytes, &anyJSON) == nil {
				if msg := BestJSONMessage(anyJSON); msg != "" {
					message = msg
				} else {
					message = strings.TrimSpace(string(bodyBytes))
				}
			} else {
				message = strings.TrimSpace(string(bodyBytes))
			}
		}
	}

	// Fallback to the original error text if we didn't get a body message
	if message == "" && err != nil && statusLine != err.Error() {
		message = err.Error()
	}

	// Append request info at the end of the message line
	if reqInfo != "" {
		if message == "" {
			message = fmt.Sprintf("(%s)", reqInfo)
		} else {
			message = fmt.Sprintf("%s (%s)", message, reqInfo)
		}
	}

	// Produce the final two-line string
	switch {
	case statusLine != "" && message != "":
		return statusLine + "\n" + message
	case statusLine != "":
		return statusLine
	default:
		return message
	}
}

// BestJSONMessage tries to extract a single, human-friendly sentence from common DRF error shapes,
// e.g. {"detail":"..."} or {"name":["..."]} or {"non_field_errors":["..."]}. Field names are stripped.
func BestJSONMessage(v any) string {
	// 1) Exact DRF "detail"
	if m, ok := v.(map[string]any); ok {
		if d, ok := m["detail"]; ok {
			if s := Stringify(d); strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}

	// Prefer non-field errors if present
	if m, ok := v.(map[string]any); ok {
		if nfe, ok := m["non_field_errors"]; ok {
			msgs := CollectStrings(nfe)
			if len(msgs) > 0 {
				return strings.Join(msgs, " | ")
			}
		}
	}

	// Single field with messages -> return the first message only
	if m, ok := v.(map[string]any); ok {
		all := []string{}
		for _, key := range SortedMapKeys(m) {
			all = append(all, CollectStrings(m[key])...)
		}
		if len(all) == 1 {
			return strings.TrimSpace(all[0])
		}
		if len(all) > 1 {
			return strings.Join(all, " | ")
		}
	}

	// Top-level list of messages
	if arr, ok := v.([]any); ok {
		msgs := CollectStrings(arr)
		if len(msgs) > 0 {
			return strings.Join(msgs, " | ")
		}
	}

	return ""
}

// CollectStrings walks typical DRF error shapes and returns only the message strings,
// ignoring field names (so {"name":["msg"]} -> ["msg"]).
func CollectStrings(v any) []string {
	out := []string{}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s != "" {
			out = append(out, s)
		}
	case []any:
		for _, it := range t {
			out = append(out, CollectStrings(it)...)
		}
	case map[string]any:
		for _, key := range SortedMapKeys(t) {
			out = append(out, CollectStrings(t[key])...)
		}
	}
	return out
}

func SortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func Stringify(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		b, _ := json.Marshal(s)
		return string(b)
	}
}
