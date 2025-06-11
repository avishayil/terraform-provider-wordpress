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
