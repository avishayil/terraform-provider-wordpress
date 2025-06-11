// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/stretchr/testify/assert"
)

func TestWordpressProvider_Metadata(t *testing.T) {
	wp := &WordpressProvider{version: "test-version"}
	resp := &provider.MetadataResponse{}
	wp.Metadata(context.Background(), provider.MetadataRequest{}, resp)
	assert.Equal(t, "wordpress", resp.TypeName)
	assert.Equal(t, "test-version", resp.Version)
}

func TestWordpressProvider_Schema(t *testing.T) {
	wp := &WordpressProvider{}
	resp := &provider.SchemaResponse{}
	wp.Schema(context.Background(), provider.SchemaRequest{}, resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "ssh_target")
	assert.Contains(t, resp.Schema.Attributes, "remote_path")
	assert.Contains(t, resp.Schema.Attributes, "allow_root")
}

func TestWordpressProvider_Resources(t *testing.T) {
	wp := &WordpressProvider{}
	res := wp.Resources(context.Background())
	assert.Len(t, res, 1)
	assert.NotNil(t, res[0])
}

func TestWordpressProvider_DataSources(t *testing.T) {
	wp := &WordpressProvider{}
	ds := wp.DataSources(context.Background())
	assert.Empty(t, ds)
}

func TestWordpressProvider_Functions(t *testing.T) {
	wp := &WordpressProvider{}
	fns := wp.Functions(context.Background())
	assert.Empty(t, fns)
}

func TestWordpressProvider_EphemeralResources(t *testing.T) {
	wp := &WordpressProvider{}
	ers := wp.EphemeralResources(context.Background())
	assert.Empty(t, ers)
}

func TestNew(t *testing.T) {
	pfn := New("foo")
	prov := pfn()
	wp, ok := prov.(*WordpressProvider)
	assert.True(t, ok)
	assert.Equal(t, "foo", wp.version)
}
