// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewOptionResource() resource.Resource {
	return &wordpressOptionResource{}
}

type wordpressOptionResource struct {
	config *WPConfig
}

type wordpressOptionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *wordpressOptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_option"
}

func (r *wordpressOptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the WordPress option (e.g., 'blogname', 'timezone_string').",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The value to assign to the option.",
			},
		},
	}
}

func (r *wordpressOptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*WPConfig)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", "Expected *WPConfig")
		return
	}
	r.config = cfg
}

func (r *wordpressOptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cfg := r.config

	var plan wordpressOptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := RunWP(cfg, "option", "set", plan.Name.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to set option", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (r *wordpressOptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	cfg := r.config

	var state wordpressOptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := RunWPWithOutput(cfg, "option", "get", state.Name.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Value = types.StringValue(strings.TrimSpace(output))
	resp.State.Set(ctx, &state)
}

func (r *wordpressOptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	cfg := r.config

	var plan wordpressOptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := RunWP(cfg, "option", "set", plan.Name.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update option", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (r *wordpressOptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	cfg := r.config

	var state wordpressOptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := RunWP(cfg, "option", "delete", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete option", err.Error())
		return
	}
}
