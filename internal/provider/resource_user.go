// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewUserResource() resource.Resource {
	return &wordpressUserResource{}
}

type wordpressUserResource struct {
	config *WPConfig
}

type wordpressUserModel struct {
	Username    types.String `tfsdk:"username"`
	Email       types.String `tfsdk:"email"`
	Password    types.String `tfsdk:"password"`
	Role        types.String `tfsdk:"role"`
	DisplayName types.String `tfsdk:"display_name"`
	FirstName   types.String `tfsdk:"first_name"`
	LastName    types.String `tfsdk:"last_name"`
}

func (r *wordpressUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *wordpressUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a WordPress user account.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Unique username for the WordPress user.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email address of the user.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Login password for the user.",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "User role (e.g., administrator, editor, subscriber).",
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Description: "Display name for the user.",
			},
			"first_name": schema.StringAttribute{
				Optional:    true,
				Description: "First name of the user.",
			},
			"last_name": schema.StringAttribute{
				Optional:    true,
				Description: "Last name of the user.",
			},
		},
	}
}

func (r *wordpressUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wordpressUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cfg := r.config

	var plan wordpressUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{
		"user", "create",
		plan.Username.ValueString(),
		plan.Email.ValueString(),
		"--user_pass=" + plan.Password.ValueString(),
		"--role=" + plan.Role.ValueString(),
	}

	if !plan.DisplayName.IsNull() {
		args = append(args, "--display_name="+plan.DisplayName.ValueString())
	}
	if !plan.FirstName.IsNull() {
		args = append(args, "--first_name="+plan.FirstName.ValueString())
	}
	if !plan.LastName.IsNull() {
		args = append(args, "--last_name="+plan.LastName.ValueString())
	}

	if err := RunWP(cfg, args...); err != nil {
		resp.Diagnostics.AddError("Failed to create WordPress user", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (r *wordpressUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	cfg := r.config

	var state wordpressUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if user exists
	if _, err := RunWPWithOutput(cfg, "user", "get", state.Username.ValueString()); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.State.Set(ctx, &state)
}

func (r *wordpressUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	cfg := r.config

	var plan wordpressUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"user", "update", plan.Username.ValueString()}

	if !plan.Email.IsNull() {
		args = append(args, "--user_email="+plan.Email.ValueString())
	}
	if !plan.Role.IsNull() {
		args = append(args, "--role="+plan.Role.ValueString())
	}
	if !plan.DisplayName.IsNull() {
		args = append(args, "--display_name="+plan.DisplayName.ValueString())
	}
	if !plan.FirstName.IsNull() {
		args = append(args, "--first_name="+plan.FirstName.ValueString())
	}
	if !plan.LastName.IsNull() {
		args = append(args, "--last_name="+plan.LastName.ValueString())
	}

	// Password update requires a separate command
	if !plan.Password.IsNull() {
		if err := RunWP(cfg, "user", "update", plan.Username.ValueString(), "--user_pass="+plan.Password.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to update password", err.Error())
			return
		}
	}

	if err := RunWP(cfg, args...); err != nil {
		resp.Diagnostics.AddError("Failed to update WordPress user", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (r *wordpressUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	cfg := r.config

	var state wordpressUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := RunWP(cfg, "user", "delete", state.Username.ValueString(), "--yes"); err != nil {
		resp.Diagnostics.AddError("Failed to delete WordPress user", err.Error())
		return
	}
}
