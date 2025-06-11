# WordPress User Resource Example

resource "wordpress_user" "admin" {
  username     = "admin"
  email        = "admin@example.com"
  password     = "SuperSecure123!"
  role         = "administrator"
  display_name = "Avishay Bar"
  first_name   = "Avishay"
  last_name    = "Bar"
}
