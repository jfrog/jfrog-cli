output "instance_admin_password" {
  value = random_string.password.*.result
}

