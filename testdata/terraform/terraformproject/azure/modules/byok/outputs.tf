output "postgres_fqdn" {
  value = element(concat(azurerm_postgresql_server.byok_postgres.*.fqdn, [""]), 0)
}

output "postgres_server_name" {
  value = element(concat(azurerm_postgresql_server.byok_postgres.*.name, [""]), 0)
}

output "postgres_admin_username" {
  value = element(
    concat(
      azurerm_postgresql_server.byok_postgres.*.administrator_login,
      [""],
    ),
    0,
  )
}

output "postgres_admin_password" {
  value = element(
    concat(
      azurerm_postgresql_server.byok_postgres.*.administrator_login_password,
      [""],
    ),
    0,
  )
}

output "byok_vault_key" {
  value = azurerm_key_vault_key.byok_vault_key.id
}

output "byok_vault_key_versionless" {
  value = azurerm_key_vault_key.byok_vault_key.versionless_id
}