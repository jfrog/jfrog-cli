
#Azure Generic vNet Module
resource "azurerm_resource_group" "network" {
  count    = var.module_enabled ? 1 : 0
  name     = var.short_region != " " ? var.short_region : "${var.deploy_name}-${var.region}"
  location = var.region

  tags = {
    environment = var.environment
  }
}

resource "azurerm_virtual_network" "vnet" {
  count               = var.module_enabled ? 1 : 0
  name                = "${var.deploy_name}-${var.region}"
  location            = var.region
  address_space       = [var.vpc_cidr]
  resource_group_name = azurerm_resource_group.network[0].name

  tags = {
    environment = var.environment
    costcenter  = "${var.deploy_name}-${var.region}"
  }
}

resource "azurerm_subnet" "subnet" {
  count                = var.module_enabled ? length(var.subnet_names) : 0
  name                 = var.subnet_names[count.index]
  virtual_network_name = azurerm_virtual_network.vnet[0].name
  resource_group_name  = azurerm_resource_group.network[0].name
  address_prefixes     = [var.subnet_prefixes[count.index]]
  enforce_private_link_endpoint_network_policies = var.subnet_names[count.index] == "private" && var.enforce_private_subnet
  
}