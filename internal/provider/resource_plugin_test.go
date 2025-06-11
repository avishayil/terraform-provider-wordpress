// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

//go:build acceptance
// +build acceptance

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestWordpressPluginResource_Metadata(t *testing.T) {
	res := &wordpressPluginResource{}
	resp := &resource.MetadataResponse{}
	req := resource.MetadataRequest{ProviderTypeName: "wordpress"}
	res.Metadata(context.Background(), req, resp)
	assert.Equal(t, "wordpress_plugin", resp.TypeName)
}

func TestWordpressPluginResource_Schema(t *testing.T) {
	res := &wordpressPluginResource{}
	resp := &resource.SchemaResponse{}
	res.Schema(context.Background(), resource.SchemaRequest{}, resp)

	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "active")
}

func TestWordpressPluginResource_Configure(t *testing.T) {
	res := &wordpressPluginResource{}

	// Valid
	wpCfg := &WPConfig{
		SSHTarget:  "foo",
		RemotePath: "bar",
		AllowRoot:  true,
	}
	resp := &resource.ConfigureResponse{}
	req := resource.ConfigureRequest{ProviderData: wpCfg}
	res.Configure(context.Background(), req, resp)
	assert.Equal(t, wpCfg, res.config)
	assert.False(t, resp.Diagnostics.HasError())

	// Nil ProviderData
	res2 := &wordpressPluginResource{}
	resp2 := &resource.ConfigureResponse{}
	req2 := resource.ConfigureRequest{ProviderData: nil}
	res2.Configure(context.Background(), req2, resp2)
	assert.Nil(t, res2.config)
	assert.False(t, resp2.Diagnostics.HasError())

	// Invalid type
	res3 := &wordpressPluginResource{}
	resp3 := &resource.ConfigureResponse{}
	req3 := resource.ConfigureRequest{ProviderData: 123}
	res3.Configure(context.Background(), req3, resp3)
	assert.Nil(t, res3.config)
	assert.True(t, resp3.Diagnostics.HasError())
	assert.Contains(t, resp3.Diagnostics.Errors()[0].Summary(), "Unexpected Provider Data Type")
}

func TestIsPluginActive_StatusLine(t *testing.T) {
	assert.True(t, isPluginActive("Status: active\n"))
	assert.False(t, isPluginActive("Status: inactive\n"))
	assert.True(t, isPluginActive("status:    active  \nother stuff"))
	assert.False(t, isPluginActive("status:   inactive\nother stuff"))
}

func TestIsPluginActive_NameWithStatus(t *testing.T) {
	assert.True(t, isPluginActive("hello-dolly (active)\n"))
	assert.True(t, isPluginActive("akismet [active]\n"))
	assert.False(t, isPluginActive("hello-dolly (inactive)\n"))
	assert.False(t, isPluginActive("akismet inactive\n"))
	assert.True(t, isPluginActive("akismet active.\n"))
	assert.True(t, isPluginActive("plugin foobar active,\n"))
}

func TestIsPluginActive_Unrecognized(t *testing.T) {
	assert.False(t, isPluginActive("plugin foobar is something\n"))
}
