# WordPress User Data Source Example

data "wordpress_user" "admin" {
  username = "admin"
}

output "admin_email" {
  value = data.wordpress_user.admin.email
}

output "admin_role" {
  value = data.wordpress_user.admin.role
}
