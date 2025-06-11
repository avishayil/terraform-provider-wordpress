// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &WordpressProvider{}
var _ provider.ProviderWithFunctions = &WordpressProvider{}
var _ provider.ProviderWithEphemeralResources = &WordpressProvider{}

type WordpressProvider struct {
	version string
}

type WordpressProviderModel struct {
	SSHTarget  types.String `tfsdk:"ssh_target"`
	RemotePath types.String `tfsdk:"remote_path"`
	AllowRoot  types.Bool   `tfsdk:"allow_root"`
}

func (p *WordpressProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wordpress"
	resp.Version = p.version
}

func (p *WordpressProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ssh_target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The SSH target for remote WordPress execution. E.g., 'docker:container-name' or 'user@host'.",
			},
			"remote_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The path to the WordPress installation on the remote system.",
			},
			"allow_root": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to add --allow-root to WP-CLI commands.",
			},
		},
	}
}

func (p *WordpressProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data WordpressProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := &WPConfig{
		SSHTarget:  defaultStringIfUnset(data.SSHTarget, ""),
		RemotePath: defaultStringIfUnset(data.RemotePath, ""),
		AllowRoot:  defaultBoolIfUnset(data.AllowRoot, false),
	}

	resp.ResourceData = cfg
	resp.DataSourceData = cfg
}

func (p *WordpressProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPluginResource,
	}
}

func (p *WordpressProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *WordpressProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *WordpressProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WordpressProvider{version: version}
	}
}
