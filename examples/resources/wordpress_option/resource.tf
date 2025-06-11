# WordPress Option Resource Example

resource "wordpress_option" "site_name" {
  name  = "blogname"
  value = "My Awesome Terraform Site" # Set the site name
}

resource "wordpress_option" "timezone" {
  name  = "timezone_string"
  value = "Europe/Paris" # Set the timezone to Europe/Paris
}