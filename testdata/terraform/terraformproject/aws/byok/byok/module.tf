
data "azurerm_client_config" "current" {
}

data "azuread_group" "devops_infra_group" {
  display_name     = "DevOpsInfraAdmin"
  security_enabled = true
}

resource "azurerm_key_vault_key" "byok_vault_key" {
  name         = "byok-key-${var.customer_name}"
  key_vault_id =  var.key_vault_id
  key_type     = "RSA"
  key_size     = 2048

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey",
  ]
  depends_on = [
//    azurerm_key_vault_access_policy.byok_postgres_key_access_policy,
    data.azurerm_client_config.current
  ]
}



resource "random_password" "byok_database_password" {
  length      = 16
  min_lower   = 2
  min_numeric = 2
  min_special = 2
  min_upper   = 2
  number      = true
  special     = true
  upper       = true
}


resource "azurerm_postgresql_server" "byok_postgres" {
  count = var.dbs_count
  name  = "${var.deploy_name}-${var.region}-byok-${var.customer_name}-${count.index}"
  location                     = var.region
  resource_group_name          = var.resource_group_name
  sku_name                     = var.postgres_dbs[count.index]["sku"]
  storage_mb                   = var.postgres_dbs[count.index]["override_disk_size"]
  backup_retention_days        = var.postgres_dbs[count.index]["backup_retention_days"]
  geo_redundant_backup_enabled = "false"
  ssl_enforcement_enabled      = "true"
  administrator_login          = var.user_name
  administrator_login_password = var.user_password == "" ? random_password.byok_database_password.result : var.user_password
  version                      = var.postgres_dbs[count.index]["postgres_version"]
  create_mode                  = "Default"

  public_network_access_enabled = false

  tags = {
    Environment = var.environment
    Application = try(var.postgres_dbs[count.index]["tags"]["application"], "common")
  }
  identity {
    type = "SystemAssigned"
  }
  depends_on = [
    azurerm_key_vault_key.byok_vault_key,
    random_password.byok_database_password
  ]
}

resource "azurerm_management_lock" "byok_postgres_delete_lock" {
  count = var.dbs_count
  name       = "subscription-level"
  scope      = azurerm_postgresql_server.byok_postgres[count.index].id
  lock_level = "CanNotDelete"
  notes      = "Postgres accidental deletion protection is locked by terraform!"
  depends_on = [
    azurerm_postgresql_server.byok_postgres
  ]
}

resource "azurerm_key_vault_access_policy" "byok_postgres_key_access_policy" {
  count            = var.dbs_count
  key_vault_id     = var.key_vault_id
  tenant_id        = data.azurerm_client_config.current.tenant_id
  object_id        = azurerm_postgresql_server.byok_postgres[count.index].identity.0.principal_id

  key_permissions    = ["get", "unwrapkey", "wrapkey"]
  secret_permissions = ["get"]
  depends_on = [
    azurerm_postgresql_server.byok_postgres
  ]
}

resource "azurerm_postgresql_server_key" "byok_postgres_vault_key" {
  count            = var.dbs_count
  server_id        = azurerm_postgresql_server.byok_postgres[count.index].id
  key_vault_key_id = azurerm_key_vault_key.byok_vault_key.id
  depends_on = [
    azurerm_key_vault_key.byok_vault_key
  ]
}

resource "azurerm_storage_account" "byok_account" {
  count                    = var.byok_storage_enable
  name                     = var.customer_name
  resource_group_name      = var.resource_group_name
  location                 = var.region
  account_tier             = var.account_tier
  account_kind             = var.account_kind
  account_replication_type = var.account_replication_type
  allow_blob_public_access = var.allow_blob_public_access
  // enable_advanced_threat_protection = "${var.enable_advanced_threat_protection}"
  // enable_https_traffic_only = var.enable_https_traffic_only

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_storage_container" "byok_container" {
  count                 = var.byok_storage_enable
  name                  = "vhds"
  storage_account_name  = azurerm_storage_account.byok_account[count.index].name
  container_access_type = "private"
}

resource "azurerm_storage_account_customer_managed_key" "byok_storage" {
  count              = var.byok_storage_enable
  storage_account_id = azurerm_storage_account.byok_account[count.index].id
  key_vault_id       = var.key_vault_id
  key_name           = azurerm_key_vault_key.byok_vault_key.name
  key_version        = azurerm_key_vault_key.byok_vault_key.version
}

resource "azurerm_private_endpoint" "private_ep_ip" {
  count               = var.dbs_count
  name                = "${azurerm_postgresql_server.byok_postgres[count.index].name}-endpoint"
  location            = var.region
  resource_group_name = var.resource_group_name
  subnet_id           = var.data_subnet

  private_service_connection {
    name                           = "${azurerm_postgresql_server.byok_postgres[count.index].name}-privateserviceconnection"
    private_connection_resource_id = azurerm_postgresql_server.byok_postgres[count.index].id
    subresource_names              = ["postgresqlServer"]
    is_manual_connection           = false
  }
  depends_on = [azurerm_postgresql_server.byok_postgres]
}
resource "azurerm_private_dns_a_record" "private_ep_ip_dns_record" {
  count               = var.dbs_count
  name                = azurerm_postgresql_server.byok_postgres[count.index].name
  zone_name           = var.private_dns_name
  resource_group_name = var.resource_group_name
  ttl                 = 300
  records             = [azurerm_private_endpoint.private_ep_ip[count.index].private_service_connection[0].private_ip_address]

  depends_on = [azurerm_private_endpoint.private_ep_ip]
}

resource "azurerm_postgresql_configuration" "idle_in_transaction_session_timeout" {
  count               = var.dbs_count
  name                = "idle_in_transaction_session_timeout"
  resource_group_name = var.resource_group_name
  server_name         = azurerm_postgresql_server.byok_postgres[count.index].name
  value               = "300000"
}

resource "azurerm_postgresql_configuration" "connection_throttling" {
  count               = var.dbs_count
  name                = "connection_throttling"
  resource_group_name = var.resource_group_name
  server_name         = azurerm_postgresql_server.byok_postgres[count.index].name
  value               = "OFF"
}

resource "azurerm_postgresql_configuration" "log_min_duration_statement" {
  count               = var.dbs_count
  name                = "log_min_duration_statement"
  resource_group_name = var.resource_group_name
  server_name         = azurerm_postgresql_server.byok_postgres[count.index].name
  value               = "1000"
}
