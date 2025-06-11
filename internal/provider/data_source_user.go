package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewUserDataSource() datasource.DataSource {
	return &wordpressUserDataSource{}
}

type wordpressUserDataSource struct {
	config *WPConfig
}

type wordpressUserDataSourceModel struct {
	Username    types.String `tfsdk:"username"`
	Email       types.String `tfsdk:"email"`
	Role        types.String `tfsdk:"role"`
	DisplayName types.String `tfsdk:"display_name"`
	FirstName   types.String `tfsdk:"first_name"`
	LastName    types.String `tfsdk:"last_name"`
}

func (d *wordpressUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *wordpressUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads metadata about a WordPress user by username.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username to query.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "The user’s email address.",
			},
			"role": schema.StringAttribute{
				Computed:    true,
				Description: "The user’s role.",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The user’s display name.",
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: "The user’s first name.",
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: "The user’s last name.",
			},
		},
	}
}

func (d *wordpressUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*WPConfig)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", "Expected *WPConfig")
		return
	}
	d.config = cfg
}

func (d *wordpressUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	cfg := d.config

	var data wordpressUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()

	// First check if the user exists at all
	_, err := RunWPWithOutput(cfg, "user", "get", username)
	if err != nil {
		resp.Diagnostics.AddError(
			"User not found",
			fmt.Sprintf("User %s does not exist or cannot be accessed: %v", username, err),
		)
		return
	}

	// Enhanced field reader with better error handling
	readField := func(field string) string {
		val, err := RunWPWithOutput(cfg, "user", "get", username, "--field="+field)
		if err != nil {
			// Try user meta as fallback
			metaVal, metaErr := RunWPWithOutput(cfg, "user", "meta", "get", username, field)
			if metaErr == nil && metaVal != "" {
				return strings.TrimSpace(metaVal)
			}

			// Log but don't error out
			fmt.Printf("DEBUG: Could not fetch field %s for user %s: %v\n", field, username, err)
			return ""
		}
		return strings.TrimSpace(val)
	}

	// Populate fields
	data.Email = types.StringValue(readField("user_email"))
	data.Role = types.StringValue(readField("roles"))
	data.DisplayName = types.StringValue(readField("display_name"))
	data.FirstName = types.StringValue(readField("first_name"))
	data.LastName = types.StringValue(readField("last_name"))

	resp.State.Set(ctx, &data)
}
