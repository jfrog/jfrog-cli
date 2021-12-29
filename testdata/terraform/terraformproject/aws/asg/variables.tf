variable "module_enabled" {
  default = true
}

variable "region" {
}

variable "deploy_name" {
}

variable "service_name" {
}

variable "disk_size_gb" {
}

variable "machine_type" {
}

variable "instance_count" {
}

variable "subnets" {
  type = list(string)
}

variable "vpc_id" {
}

variable "public_ip" {
}

variable "mmsGroupId" {
}

variable "mmsApiKey" {
}

variable "key_name" {
}

variable "security_groups" {
  type = list(string)
}

variable "asg_specified_ami" {
}

variable "port" {
}

