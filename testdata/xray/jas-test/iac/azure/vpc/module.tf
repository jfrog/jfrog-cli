
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
#  service_endpoints = [
#    "Microsoft.KeyVault"
#  ]
  
  dynamic "delegation"{
    for_each =var.subnet_names[count.index] == "flexible-dbs" ? ["exec"] : []
    content { 
      name = "dlg-Microsoft.DBforPostgreSQL-flexibleServers"
      service_delegation {
        name = "Microsoft.DBforPostgreSQL/flexibleServers"
         actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action"
      ]
    }
  }
  }

  enforce_private_link_endpoint_network_policies = var.subnet_names[count.index] == "data"
  enforce_private_link_service_network_policies = var.subnet_names[count.index] == "private" && var.enforce_pl_svc_net_private
  lifecycle {
    ignore_changes = [
      service_endpoints,
      delegation[0].name
    ]
  }
}


resource "azurerm_private_dns_zone" "postgres_private_dns" {
  count               = var.module_enabled ? 1 : 0
  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = azurerm_resource_group.network[0].name
}

resource "random_string" "postgres_private_dns_net_link_name" {
  count   = var.module_enabled ? 1 : 0
  length  = 8
  special = false
  number  = false
  upper   = false
}

resource "azurerm_private_dns_zone_virtual_network_link" "postgres_private_dns_net_link" {
  count                 = var.module_enabled ? 1 : 0
  name                  = random_string.postgres_private_dns_net_link_name[0].result
  resource_group_name   = azurerm_resource_group.network[0].name
  private_dns_zone_name = azurerm_private_dns_zone.postgres_private_dns[0].name
  virtual_network_id    = azurerm_virtual_network.vnet[0].id
}

//resource "azurerm_network_security_group" "nsg" {
//  count               = "${var.module_enabled ? length(var.subnet_names) : 0}"
//  name                = "${var.subnet_names[count.index]}-sg"
//  location            = "${var.region}"
//  resource_group_name = "${var.deploy_name}-${var.region}"
//}
//
//resource "azurerm_subnet_network_security_group_association" "nsg" {
//  count                     = "${var.module_enabled ? length(var.subnet_names) : 0}"
//  subnet_id                 = "${element(azurerm_subnet.subnet.*.id, count.index)}"
//  network_security_group_id = "${element(azurerm_network_security_group.nsg.*.id, count.index)}"
//}
//resource "azurerm_subnet_route_table_association" "nat" {
//  count          = "${var.module_enabled ? length(var.nat_subnets) : 0}"
//  subnet_id      = "${element(azurerm_subnet.subnet.*.id, count.index + 1)}"
//  route_table_id = "${azurerm_route_table.nattable.id}"
//}
# UDR
//resource "azurerm_route_table" "nattable" {
//  count               = "${var.module_enabled}"
//  name                 = "${var.deploy_name}-${var.region}"
//  location             = "${var.region}"
//  resource_group_name  = "${azurerm_resource_group.network.name}"
//
//  route {
//    name           = "all-traffic-via-nat"
//    address_prefix = "0.0.0.0/0"
//    next_hop_type  = "VirtualAppliance"
//    next_hop_in_ip_address = "${var.natgw_private_ip}"
//  }
//
//  tags = {
//    environment = "${var.environment}"
//  }
//}
