// Package shared contains small, provider-wide building blocks used by object
// families. It deliberately does not contain generic Terraform CRUD machinery;
// resource lifecycle code stays explicit in each family package.
//
// State normalization policy:
//   - optional strings and relationship IDs use an empty string when Nautobot
//     omits them, matching resource defaults and avoiding perpetual plans;
//   - optional lists use an empty list;
//   - absent timestamps remain Terraform null because no useful default exists;
//   - API lookup and conversion failures are diagnostics, never silent zero values.
package shared
