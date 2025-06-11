// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPluginResource() resource.Resource {
	return &wordpressPluginResource{}
}

type wordpressPluginResource struct {
	config *WPConfig
}

type wordpressPluginModel struct {
	Name   types.String `tfsdk:"name"`
	Active types.Bool   `tfsdk:"active"`
}

func (r *wordpressPluginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}

func (r *wordpressPluginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The WP‑CLI plugin slug (e.g., 'akismet').",
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the plugin should be activated.",
			},
		},
	}
}

func (r *wordpressPluginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wordpressPluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cfg := r.config

	var plan wordpressPluginModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build install command
	args := []string{"plugin", "install", plan.Name.ValueString()}
	if plan.Active.ValueBool() {
		args = append(args, "--activate")
	}

	fmt.Printf("DEBUG: Installing plugin %s, active=%t\n",
		plan.Name.ValueString(), plan.Active.ValueBool())

	// Install the plugin
	if err := RunWP(cfg, args...); err != nil {
		resp.Diagnostics.AddError("Failed to install plugin", err.Error())
		return
	}

	// Add a delay after installation
	time.Sleep(3 * time.Second)

	// Verify that the plugin was installed
	if err := RunWP(cfg, "plugin", "is-installed", plan.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Plugin not installed after install attempt",
			fmt.Sprintf("Failed to verify plugin %s is installed: %v", plan.Name.ValueString(), err))
		return
	}

	// Query plugin status
	output, err := RunWPWithOutput(cfg, "plugin", "status", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify plugin status",
			fmt.Sprintf("Command failed: %v\nOutput: %s", err, output))
		return
	}

	active := isPluginActive(output)
	fmt.Printf("DEBUG: After installation, plugin %s active=%t\n",
		plan.Name.ValueString(), active)

	// If we wanted it inactive but it's active, explicitly deactivate it
	if !plan.Active.ValueBool() && active {
		fmt.Printf("DEBUG: Plugin was activated by default, deactivating...\n")
		if err := RunWP(cfg, "plugin", "deactivate", plan.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to deactivate plugin", err.Error())
			return
		}

		// Wait for deactivation to take effect
		time.Sleep(3 * time.Second)

		output, err = RunWPWithOutput(cfg, "plugin", "status", plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to verify plugin status after deactivation",
				fmt.Sprintf("Command failed: %v\nOutput: %s", err, output))
			return
		}
		active = isPluginActive(output)
		fmt.Printf("DEBUG: After explicit deactivation, plugin active=%t\n", active)
	}

	plan.Active = types.BoolValue(active)
	resp.State.Set(ctx, &plan)
}

func (r *wordpressPluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	cfg := r.config

	var state wordpressPluginModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if plugin is installed
	if err := RunWP(cfg, "plugin", "is-installed", state.Name.ValueString()); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Get plugin status
	output, err := RunWPWithOutput(cfg, "plugin", "status", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get plugin status",
			fmt.Sprintf("Command failed: %v\nOutput: %s", err, output))
		return
	}

	active := isPluginActive(output)
	fmt.Printf("DEBUG: Read operation - plugin %s active=%t\n",
		state.Name.ValueString(), active)

	state.Active = types.BoolValue(active)
	resp.State.Set(ctx, &state)
}

func (r *wordpressPluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	cfg := r.config

	var plan wordpressPluginModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for comparison
	var state wordpressPluginModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only take action if active state is changing
	if !plan.Active.IsUnknown() && !plan.Active.IsNull() &&
		plan.Active.ValueBool() != state.Active.ValueBool() {

		fmt.Printf("DEBUG: Changing plugin %s active state from %t to %t\n",
			plan.Name.ValueString(), state.Active.ValueBool(), plan.Active.ValueBool())

		var err error
		if plan.Active.ValueBool() {
			err = RunWP(cfg, "plugin", "activate", plan.Name.ValueString())
			fmt.Printf("DEBUG: Activated plugin %s\n", plan.Name.ValueString())
		} else {
			err = RunWP(cfg, "plugin", "deactivate", plan.Name.ValueString())
			fmt.Printf("DEBUG: Deactivated plugin %s\n", plan.Name.ValueString())
		}

		if err != nil {
			resp.Diagnostics.AddError("Failed to update plugin activation", err.Error())
			return
		}

		// Add a delay to ensure WordPress has processed the change
		fmt.Printf("DEBUG: Waiting 3 seconds for WordPress to process the change...\n")
		time.Sleep(3 * time.Second)
	}

	// Allow time for the change to take effect
	time.Sleep(3 * time.Second)

	// Re-read status to reflect actual state
	output, err := RunWPWithOutput(cfg, "plugin", "status", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify plugin status",
			fmt.Sprintf("Command failed: %v\nOutput: %s", err, output))
		return
	}

	active := isPluginActive(output)
	fmt.Printf("DEBUG: After operation, plugin %s is active=%t\n",
		plan.Name.ValueString(), active)

	plan.Active = types.BoolValue(active)
	resp.State.Set(ctx, &plan)
}

func (r *wordpressPluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	cfg := r.config

	var state wordpressPluginModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := RunWP(cfg, "plugin", "delete", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete plugin", err.Error())
		return
	}
}

func isPluginActive(output string) bool {
	fmt.Printf("DEBUG: Raw plugin status output:\n%s\n", output)

	// First check for explicit status line
	for _, line := range strings.Split(output, "\n") {
		trimLine := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(trimLine, "status:") {
			status := strings.TrimSpace(strings.TrimPrefix(trimLine, "status:"))
			fmt.Printf("DEBUG: Found status line: '%s', extracted status: '%s'\n", trimLine, status)
			return status == "active"
		}
	}

	// Next check for plugin name with status
	for _, line := range strings.Split(output, "\n") {
		line = strings.ToLower(strings.TrimSpace(line))
		// Look for lines that might contain plugin info
		if strings.Contains(line, "plugin") || strings.Contains(line, "akismet") ||
			strings.Contains(line, "hello-dolly") {
			fmt.Printf("DEBUG: Found potential plugin line: '%s'\n", line)

			// Check for inactive status
			if strings.Contains(line, "inactive") {
				return false
			}
			// Check for active status, but make sure it's not just part of another word
			if (strings.Contains(line, " active") ||
				strings.Contains(line, "(active)") ||
				strings.Contains(line, "[active]") ||
				strings.Contains(line, "active,") ||
				strings.Contains(line, "active.")) &&
				!strings.Contains(line, "inactive") {
				return true
			}
		}
	}

	fmt.Printf("DEBUG: Could not determine plugin status from output, defaulting to inactive\n")
	return false
}
