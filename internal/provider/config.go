// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// WPConfig holds the configuration for executing WP-CLI commands.
type WPConfig struct {
	SSHTarget  string
	RemotePath string
	AllowRoot  bool
}

// defaultStringIfUnset returns the default value if the input is null or unknown.
func defaultStringIfUnset(val types.String, def string) string {
	if val.IsNull() || val.IsUnknown() {
		return def
	}
	return val.ValueString()
}

// defaultBoolIfUnset returns the default value if the input is null or unknown.
func defaultBoolIfUnset(val types.Bool, def bool) bool {
	if val.IsNull() || val.IsUnknown() {
		return def
	}
	return val.ValueBool()
}
