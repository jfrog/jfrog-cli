variable "module_enabled" {
  default = true
}

variable "project_name" {
}

variable "region" {
}

variable "region_zone" {
}

variable "deploy_name" {
}

variable "network" {
}

variable "subnetwork" {
}

variable "instance_tags" {
  type = list(string)
}

variable "subnet_cidr" {
  type = map(string)
}

variable "min_node_count" {
}

variable "max_node_count" {
}

variable "logging_service" {
}

variable "monitoring_service" {
}

variable "enable_legacy_abac" {
}

variable "worker_machine_type" {
}

variable "ft_machine_type" {
}

variable "image_type" {
}

variable "ng_disk_size_gb" {
}
variable "ft_disk_size_gb" {
}

variable "label" {
}

variable "natgw_ip" {
}

variable "gcp_azs" {
  type = map(string)
  default = {
    us-east1     = "us-east1-c,us-east1-d"
    us-west1     = "us-west1-c,us-west1-a"
    us-central1  = "us-central1-c,us-central1-f"
    europe-west2 = "europe-west2-a,europe-west2-c"
    europe-west1 = "europe-west1-c,europe-west1-d"
  }
}

variable "ssh_key" {
}

variable "k8s_master_version" {
}

variable "k8s_node_version" {
}

variable "client_certificate" {
}

variable "k8s_zonal" {
}

variable "override_ft_name" {
}

variable "override_ng_name" {
}

variable "gke_auth" {
}

variable "node_config" {
}