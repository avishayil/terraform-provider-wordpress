// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"terraform-provider-wordpress/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func protoV6Factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"wordpress": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

// Returns the docker SSH target from WP_CONTAINER_NAME env var.
func sshTarget() string {
	container := os.Getenv("WP_CONTAINER_NAME") + "-1"
	if container == "-1" {
		container = "terraform-provider-wordpress-wordpress-1" // fallback for local testing
	}
	return fmt.Sprintf("docker:%s", container)
}

func testConfig(pluginName string, active bool) string {
	return fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_plugin" "example" {
  name   = "%s"
  active = %t
}
`, sshTarget(), pluginName, active)
}

func TestAccWordpressPlugin_active(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testConfig("hello-dolly", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_plugin.example", "name", "hello-dolly"),
					resource.TestCheckResourceAttr("wordpress_plugin.example", "active", "true"),
				),
			},
		},
	})
}

func TestAccWordpressPlugin_inactive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testConfig("akismet", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_plugin.example", "name", "akismet"),
					resource.TestCheckResourceAttr("wordpress_plugin.example", "active", "false"),
				),
			},
		},
	})
}

func TestAccWordpressPlugin_invalid_plugin(t *testing.T) {
	config := testConfig("this-plugin-does-not-exist", true)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`Failed to install plugin`),
			},
		},
	})
}

func TestAccWordpressPlugin_bad_ssh_target(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "wordpress" {
  ssh_target  = "not-a-real-target"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_plugin" "example" {
  name   = "hello-dolly"
  active = true
}
`,
				ExpectError: regexp.MustCompile(`Failed to install plugin`),
			},
		},
	})
}

func TestAccWordpressPlugin_missing_name(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_plugin" "example" {
  active = true
}
`, sshTarget()),
				ExpectError: regexp.MustCompile(`The argument "name" is required`),
			},
		},
	})
}

func TestAccWordpressPlugin_deactivate_not_active(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				// Step 1: Install and deactivate the plugin
				Config: testConfig("classic-editor", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_plugin.example", "name", "classic-editor"),
					resource.TestCheckResourceAttr("wordpress_plugin.example", "active", "false"),
				),
			},
			{
				// Step 3: Reapply same deactivated state
				PreConfig: func() {
					time.Sleep(5 * time.Second)
				},
				Config: testConfig("classic-editor", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_plugin.example", "name", "classic-editor"),
					resource.TestCheckResourceAttr("wordpress_plugin.example", "active", "false"),
				),
			},
		},
	})
}

func TestAccWordpressTheme_installOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				// First ensure we have at least one other theme to be active
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_theme" "default" {
  name   = "twentytwentyfour"
  active = true
}
`, sshTarget()),
			},
			{
				// Now install our test theme without forcing active state
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_theme" "default" {
  name   = "twentytwentyfour"
  active = true
}

resource "wordpress_theme" "example" {
  name   = "twentytwentythree"
  # Let WordPress decide if it should be active or not
  depends_on = [wordpress_theme.default]
}
`, sshTarget()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_theme.example", "name", "twentytwentythree"),
					// Don't check active status as it depends on WordPress
				),
			},
		},
	})
}

func TestAccWordpressTheme_activateAndSwitch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_theme" "example" {
  name   = "twentytwentytwo"
  active = true
}
`, sshTarget()),
				Check: resource.TestCheckResourceAttr("wordpress_theme.example", "active", "true"),
			},
			{
				// fallback theme, non-destructive
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_theme" "fallback" {
  name   = "twentytwentyone"
  active = true
}
`, sshTarget()),
			},
		},
	})
}

func TestAccWordpressOption_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_option" "site_name" {
  name  = "blogname"
  value = "Terraform Site"
}
`, sshTarget()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_option.site_name", "name", "blogname"),
					resource.TestCheckResourceAttr("wordpress_option.site_name", "value", "Terraform Site"),
				),
			},
		},
	})
}

func TestAccWordpressSiteSettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_site_settings" "basic" {
  site_name        = "My Terraform Site"
  site_description = "IaC is awesome"
  admin_email      = "admin@example.com"
  timezone         = "Europe/London"
  date_format      = "F j, Y"
  time_format      = "g:i a"
  start_of_week    = "1"
}
`, sshTarget()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_site_settings.basic", "site_name", "My Terraform Site"),
					resource.TestCheckResourceAttr("wordpress_site_settings.basic", "timezone", "Europe/London"),
				),
			},
		},
	})
}

func TestAccWordpressUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_user" "admin" {
  username     = "terraform_admin"
  email        = "admin@terraform.dev"
  password     = "SecureP@ss123"
  role         = "administrator"
  display_name = "Terraform Admin"
  first_name   = "Terraform"
  last_name    = "Admin"
}
`, sshTarget()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wordpress_user.admin", "username", "terraform_admin"),
					resource.TestCheckResourceAttr("wordpress_user.admin", "role", "administrator"),
				),
			},
		},
	})
}

func TestAccWordpressUserDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					cfg := &provider.WPConfig{
						SSHTarget:  sshTarget(),
						RemotePath: "/var/www/html",
						AllowRoot:  true,
					}
					// Use meta update instead of user update for better compatibility
					_ = provider.RunWP(cfg, "user", "meta", "update", "admin", "first_name", "Terraform")
					_ = provider.RunWP(cfg, "user", "meta", "update", "admin", "last_name", "Admin")
					// Give WordPress time to process the updates
					time.Sleep(3 * time.Second)
				},
				Config: fmt.Sprintf(`
provider "wordpress" {
  ssh_target  = "%s"
  remote_path = "/var/www/html"
  allow_root  = true
}

data "wordpress_user" "admin" {
  username = "admin"
}
`, sshTarget()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wordpress_user.admin", "username", "admin"),
					resource.TestCheckResourceAttr("data.wordpress_user.admin", "first_name", "Terraform"),
					resource.TestCheckResourceAttr("data.wordpress_user.admin", "last_name", "Admin"),
				),
			},
		},
	})
}
