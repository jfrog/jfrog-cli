variable "dbs_count" {}

variable "byok_keyvault_exists" {
  default = true
}

variable "postgres_dbs" { type = any }

variable "region" {}

variable "environment" {}

variable "deploy_name" {}

variable "resource_group_name" {}

variable "byok_storage_enable" {
  default = 0
}

variable "user_name" {}

variable "user_password" {}

variable "allow_blob_public_access" {}

variable "account_replication_type" {}

variable "account_tier" {}

variable "account_kind" {}

variable "backup_retention_days" { default = 7 }

variable "customer_name" {
  type = string
}

variable "key_vault_id" {}

//variable "private_subnet" {}
//
variable "data_subnet" {}

//variable "private_dns_id" {}

variable "private_dns_name" {}