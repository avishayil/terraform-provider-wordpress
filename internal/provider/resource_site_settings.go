package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSiteSettingsResource() resource.Resource {
	return &wordpressSiteSettingsResource{}
}

type wordpressSiteSettingsResource struct {
	config *WPConfig
}

type wordpressSiteSettingsModel struct {
	SiteName        types.String `tfsdk:"site_name"`
	SiteDescription types.String `tfsdk:"site_description"`
	AdminEmail      types.String `tfsdk:"admin_email"`
	Timezone        types.String `tfsdk:"timezone"`
	DateFormat      types.String `tfsdk:"date_format"`
	TimeFormat      types.String `tfsdk:"time_format"`
	StartOfWeek     types.String `tfsdk:"start_of_week"`
}

func (r *wordpressSiteSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_settings"
}

func (r *wordpressSiteSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages general WordPress site settings like title, description, timezone, and more.",
		Attributes: map[string]schema.Attribute{
			"site_name": schema.StringAttribute{
				Optional:    true,
				Description: "The site's name (blogname).",
			},
			"site_description": schema.StringAttribute{
				Optional:    true,
				Description: "The site's tagline (blogdescription).",
			},
			"admin_email": schema.StringAttribute{
				Optional:    true,
				Description: "The site administrator email.",
			},
			"timezone": schema.StringAttribute{
				Optional:    true,
				Description: "The timezone (e.g., 'Europe/London').",
			},
			"date_format": schema.StringAttribute{
				Optional:    true,
				Description: "The format for displaying dates.",
			},
			"time_format": schema.StringAttribute{
				Optional:    true,
				Description: "The format for displaying times.",
			},
			"start_of_week": schema.StringAttribute{
				Optional:    true,
				Description: "Numeric day the week starts on (0=Sunday, 1=Monday).",
			},
		},
	}
}

func (r *wordpressSiteSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wordpressSiteSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cfg := r.config

	var plan wordpressSiteSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := applySiteSettings(cfg, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to apply site settings", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (r *wordpressSiteSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	cfg := r.config

	var state wordpressSiteSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readSiteSetting := func(option string) (string, error) {
		out, err := RunWPWithOutput(cfg, "option", "get", option)
		return out, err
	}

	readInto := func(field *types.String, option string) {
		val, err := readSiteSetting(option)
		if err == nil {
			*field = types.StringValue(strings.TrimSpace(val))
		}
	}

	readInto(&state.SiteName, "blogname")
	readInto(&state.SiteDescription, "blogdescription")
	readInto(&state.AdminEmail, "admin_email")
	readInto(&state.Timezone, "timezone_string")
	readInto(&state.DateFormat, "date_format")
	readInto(&state.TimeFormat, "time_format")
	readInto(&state.StartOfWeek, "start_of_week")

	resp.State.Set(ctx, &state)
}

func (r *wordpressSiteSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	cfg := r.config

	var plan wordpressSiteSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := applySiteSettings(cfg, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update site settings", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (r *wordpressSiteSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: general settings don't really "delete" — we just leave them in place.
}

func applySiteSettings(cfg *WPConfig, settings *wordpressSiteSettingsModel) error {
	pairs := []struct {
		name   types.String
		option string
	}{
		{settings.SiteName, "blogname"},
		{settings.SiteDescription, "blogdescription"},
		{settings.AdminEmail, "admin_email"},
		{settings.Timezone, "timezone_string"},
		{settings.DateFormat, "date_format"},
		{settings.TimeFormat, "time_format"},
		{settings.StartOfWeek, "start_of_week"},
	}

	for _, pair := range pairs {
		if !pair.name.IsNull() && !pair.name.IsUnknown() {
			if err := RunWP(cfg, "option", "set", pair.option, pair.name.ValueString()); err != nil {
				return fmt.Errorf("setting %s failed: %w", pair.option, err)
			}
		}
	}
	return nil
}
