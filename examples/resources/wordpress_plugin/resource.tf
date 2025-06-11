# WordPress Plugin Resource Example

resource "wordpress_plugin" "akismet" {
  name   = "akismet"
  active = true # Install and activate the Akismet plugin
}

resource "wordpress_plugin" "hello_dolly" {
  name   = "hello-dolly"
  active = false # Install but don't activate Hello Dolly
}