##################################################################################
# OUTPUT
##################################################################################

output "resource_group_id" {
  value = azurerm_resource_group.network[0].id
}

output "resource_group_name" {
  value = azurerm_resource_group.network[0].name
}

output "vnet_id" {
  value = element(concat(azurerm_virtual_network.vnet.*.id, [""]), 0)
}

output "vnet_location" {
  value = element(concat(azurerm_virtual_network.vnet.*.location, [""]), 0)
}

output "vnet_name" {
  value = element(concat(azurerm_virtual_network.vnet.*.name, [""]), 0)
}

output "private_dns_id" {
  value = element(
    concat(azurerm_private_dns_zone.postgres_private_dns.*.id, [""]),
    0,
  )
}

output "private_dns_name" {
  value = element(
    concat(azurerm_private_dns_zone.postgres_private_dns.*.name, [""]),
    0,
  )
}

//output "vnet_subnets" {
//  value       = "${azurerm_subnet.subnet.*.id}"
//}

### subnets ids ###
output "public_subnet" {
  value = element(concat(azurerm_subnet.subnet.*.id, [""]), 0)
}

output "private_subnet" {
  value = element(concat(azurerm_subnet.subnet.*.id, [""]), 1)
}
output "flexible_subnet" {
 value = element(concat(azurerm_subnet.subnet.*.id, [""]), 4)
}
output "data_subnet" {
  value = element(concat(azurerm_subnet.subnet.*.id, [""]), 2)
}

output "mgmt_subnet" {
  value = element(concat(azurerm_subnet.subnet.*.id, [""]), 3)
}

### subnets names ###
output "public_subnet_name" {
  value = element(concat(azurerm_subnet.subnet.*.name, [""]), 0)
}

output "private_subnet_name" {
  value = element(concat(azurerm_subnet.subnet.*.name, [""]), 1)
}

output "data_subnet_name" {
  value = element(concat(azurerm_subnet.subnet.*.name, [""]), 2)
}

output "mgmt_subnet_name" {
  value = element(concat(azurerm_subnet.subnet.*.name, [""]), 3)
}


