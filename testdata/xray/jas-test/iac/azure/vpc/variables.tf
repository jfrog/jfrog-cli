variable "module_enabled" {
  default = true
}

variable "region" {
}

variable "deploy_name" {
}

variable "vpc_cidr" {
}

variable "short_region" {
  default = " "
}

variable "subnet_prefixes" {
  type = list(string)
}

variable "ssh_source_ranges" {
  type = list(string)
}

variable "environment" {
}

variable "subnet_names" {
  type = list(string)
}

variable "enforce_pl_svc_net_private" {
  default = false
}
//variable "natgw_private_ip" {}
//variable "nat_subnets" {
//  type = "list"
//}
