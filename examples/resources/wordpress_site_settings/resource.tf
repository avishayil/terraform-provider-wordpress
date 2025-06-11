# WordPress Site Settings Resource Example

resource "wordpress_site_settings" "basic" {
  site_name        = "My Terraform Site"
  site_description = "IaC all the way"
  admin_email      = "admin@example.com"
  timezone         = "Europe/Tel_Aviv"
  date_format      = "F j, Y"
  time_format      = "g:i a"
  start_of_week    = "1"
}