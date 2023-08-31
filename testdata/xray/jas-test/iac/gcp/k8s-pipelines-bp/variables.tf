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

variable "autoscaling_parameters"{ 
}

variable "gke_map"{

}
variable "network_policy" {
  default = false
}
variable "maintenance_window" {
  default = {
    recurrence = "FREQ=WEEKLY;BYDAY=SU"
    start_time = "2021-11-21T01:00:00Z"
    end_time   = "2021-11-21T18:00:00Z"
  }
}

variable "gke_auth" {
}

variable "oauth_scopes" {
  default = [
     "https://www.googleapis.com/auth/logging.write",
     "https://www.googleapis.com/auth/monitoring",
    ]
}

variable "node_config"{
}
variable "enable_intranode_visibility" {}
variable "rbac_admin_roles"{
default = []
}

variable "rbac_readonly_roles"{
default = []
}

variable "resource_usage_export_config_parameters" {
  default = null
}