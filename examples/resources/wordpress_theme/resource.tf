# WordPress Theme Resource Example

resource "wordpress_theme" "default" {
  name   = "twentytwentyfour"
  active = true # Install and activate the Twenty Twenty-Four theme
}