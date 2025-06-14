---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "wordpress Provider"
subcategory: ""
description: |-
  
---

# wordpress Provider



## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `remote_path` (String) The path to the WordPress installation on the remote system.
- `ssh_target` (String) The SSH target for remote WordPress execution. E.g., 'docker:container-name' or 'user@host'.

### Optional

- `allow_root` (Boolean) Whether to add --allow-root to WP-CLI commands.
