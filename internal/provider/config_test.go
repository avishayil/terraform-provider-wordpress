// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDefaultStringIfUnset(t *testing.T) {
	def := "default-value"

	// Null string
	val := types.StringNull()
	assert.Equal(t, def, defaultStringIfUnset(val, def), "should return default for null")

	// Unknown string
	val = types.StringUnknown()
	assert.Equal(t, def, defaultStringIfUnset(val, def), "should return default for unknown")

	// Concrete string
	val = types.StringValue("actual-value")
	assert.Equal(t, "actual-value", defaultStringIfUnset(val, def), "should return actual value")
}

func TestDefaultBoolIfUnset(t *testing.T) {
	def := true

	// Null bool
	val := types.BoolNull()
	assert.Equal(t, def, defaultBoolIfUnset(val, def), "should return default for null")

	// Unknown bool
	val = types.BoolUnknown()
	assert.Equal(t, def, defaultBoolIfUnset(val, def), "should return default for unknown")

	// Concrete bool (true)
	val = types.BoolValue(true)
	assert.Equal(t, true, defaultBoolIfUnset(val, false), "should return true for true value")

	// Concrete bool (false)
	val = types.BoolValue(false)
	assert.Equal(t, false, defaultBoolIfUnset(val, true), "should return false for false value")
}
