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

func NewThemeResource() resource.Resource {
	return &wordpressThemeResource{}
}

type wordpressThemeResource struct {
	config *WPConfig
}

type wordpressThemeModel struct {
	Name   types.String `tfsdk:"name"`
	Active types.Bool   `tfsdk:"active"`
}

func (r *wordpressThemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme"
}

func (r *wordpressThemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The theme slug (e.g., 'twentytwentyfour').",
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the theme should be active. Note: WordPress requires one active theme at all times.",
			},
		},
		Description: "Manages a WordPress theme. Note that WordPress requires one theme to be active at all times, which may override your active setting.",
	}
}

func (r *wordpressThemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wordpressThemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cfg := r.config

	var plan wordpressThemeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	themeName := plan.Name.ValueString()
	shouldBeActive := plan.Active.ValueBool()

	fmt.Printf("DEBUG: Installing theme %s, requested active=%t\n", themeName, shouldBeActive)

	// Install theme without activation flag
	if err := RunWP(cfg, "theme", "install", themeName); err != nil {
		resp.Diagnostics.AddError("Failed to install theme", err.Error())
		return
	}

	time.Sleep(2 * time.Second)

	// Check actual state after installation
	currentlyActive, err := isThemeActive(cfg, themeName)
	if err != nil {
		resp.Diagnostics.AddError("Failed to determine theme status", err.Error())
		return
	}

	fmt.Printf("DEBUG: After installation, theme %s active=%t (requested=%t)\n",
		themeName, currentlyActive, shouldBeActive)

	// Take action if needed
	if currentlyActive != shouldBeActive {
		if shouldBeActive {
			// Need to activate
			if err := RunWP(cfg, "theme", "activate", themeName); err != nil {
				resp.Diagnostics.AddError("Failed to activate theme", err.Error())
				return
			}
			currentlyActive = true
		} else {
			// Try deactivating by activating another theme
			output, err := RunWPWithOutput(cfg, "theme", "list", "--status=inactive", "--field=name")
			if err == nil && output != "" {
				themes := strings.Split(strings.TrimSpace(output), "\n")
				if len(themes) > 0 {
					fallbackTheme := themes[0]
					if err := RunWP(cfg, "theme", "activate", fallbackTheme); err == nil {
						currentlyActive = false
					}
				}
			}

			// If we couldn't deactivate, accept it and add a warning
			if currentlyActive {
				resp.Diagnostics.AddWarning(
					"Theme could not be deactivated",
					"WordPress requires one active theme. Theme will remain active.")
			}
		}

		time.Sleep(2 * time.Second)
	}

	// Important: Set the state to reflect reality, not what was requested
	plan.Active = types.BoolValue(currentlyActive)
	resp.State.Set(ctx, &plan)
}

func (r *wordpressThemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	cfg := r.config

	var state wordpressThemeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if active
	active, err := isThemeActive(cfg, state.Name.ValueString())
	if err != nil {
		// If we can't determine if it's active, try deleting anyway
		active = false
	}

	if active {
		// For test environments, activate the default theme if possible
		fallbackThemes := []string{"twentytwentyfour", "twentytwentythree", "twentytwentytwo"}
		activated := false

		for _, theme := range fallbackThemes {
			// Skip if it's the theme we're trying to delete
			if theme == state.Name.ValueString() {
				continue
			}

			// Check if fallback theme is installed
			err := RunWP(cfg, "theme", "is-installed", theme)
			if err == nil {
				// Try to activate it
				err = RunWP(cfg, "theme", "activate", theme)
				if err == nil {
					activated = true
					time.Sleep(2 * time.Second)
					break
				}
			}
		}

		// If we couldn't activate a fallback, warn but don't fail
		if !activated {
			resp.Diagnostics.AddWarning("Cannot delete active theme",
				"Theme is active and no fallback could be activated. Resource will be removed from state but may still exist in WordPress.")
			return
		}
	}

	if err := RunWP(cfg, "theme", "delete", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete theme", err.Error())
		return
	}
}

func (r *wordpressThemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	cfg := r.config

	var state wordpressThemeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if installed
	if err := RunWP(cfg, "theme", "is-installed", state.Name.ValueString()); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	active, err := isThemeActive(cfg, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to check theme status", err.Error())
		return
	}

	state.Active = types.BoolValue(active)
	resp.State.Set(ctx, &state)
}

func (r *wordpressThemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	cfg := r.config

	var plan wordpressThemeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state wordpressThemeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if plan.Active.ValueBool() && !state.Active.ValueBool() {
		// Activating a theme is supported
		if err := RunWP(cfg, "theme", "activate", plan.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to activate theme", err.Error())
			return
		}
		time.Sleep(2 * time.Second)
	} else if !plan.Active.ValueBool() && state.Active.ValueBool() {
		// Trying to deactivate - check if there's another theme we can activate
		output, err := RunWPWithOutput(cfg, "theme", "list", "--status=inactive", "--field=name")
		if err != nil || output == "" {
			resp.Diagnostics.AddWarning(
				"Cannot deactivate theme",
				"WordPress requires one active theme. This theme will remain active.")
		} else {
			// Try to activate another theme
			themes := strings.Split(strings.TrimSpace(output), "\n")
			if len(themes) > 0 {
				fallbackTheme := themes[0]
				if err := RunWP(cfg, "theme", "activate", fallbackTheme); err != nil {
					resp.Diagnostics.AddWarning(
						"Failed to deactivate theme",
						"Could not activate another theme to replace this one.")
				}
			}
		}
		time.Sleep(2 * time.Second)
	}

	// Get actual state after update attempt
	active, err := isThemeActive(cfg, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify theme status", err.Error())
		return
	}

	plan.Active = types.BoolValue(active)
	resp.State.Set(ctx, &plan)
}

func isThemeActive(cfg *WPConfig, theme string) (bool, error) {
	output, err := RunWPWithOutput(cfg, "theme", "status", theme)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.ToLower(strings.TrimSpace(line))
		if strings.HasPrefix(line, "status:") && strings.Contains(line, "active") {
			return true, nil
		}
	}
	return false, nil
}
