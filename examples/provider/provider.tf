# WordPress Provider Configuration

terraform {
  required_providers {
    wordpress = {
      source = "avishayil/wordpress"
    }
  }
}

provider "wordpress" {
  ssh_target  = "user@example.com" # SSH target for WordPress host
  remote_path = "/var/www/html"    # Path to WordPress installation
  allow_root  = true               # Use --allow-root flag with WP-CLI
}