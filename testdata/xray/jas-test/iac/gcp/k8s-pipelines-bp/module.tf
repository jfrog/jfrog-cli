# Create new K8S cluster with autoscaling

data "google_container_engine_versions" "region" {
  location = var.region
}

resource "random_string" "admin-password" {
  count  = var.module_enabled ? 1 : 0
  length = 16

//  lifecycle {
//    ignore_changes = [
//      initial_node_count, master_authorized_networks_config
//    ]
//  }
}

# New K8s Cluster, if creation failed you'll need to cleanup manually before running again.
resource "google_container_cluster" "primary" {
  count                    = var.module_enabled ? 1 : 0
  provider                 = google-beta
  name                     = "${var.deploy_name}-${var.region}"
  location                 = var.k8s_zonal == "" ? var.region : var.region_zone
  min_master_version       = var.k8s_master_version == "" ? data.google_container_engine_versions.region.latest_master_version :lookup(var.gke_map.override,"k8s_master_version")
  network                  = var.network
  subnetwork               = var.subnetwork
  logging_service          = var.logging_service
  monitoring_service       = var.monitoring_service
  enable_legacy_abac       = var.enable_legacy_abac
  remove_default_node_pool = "true"
  initial_node_count       = 1
  enable_shielded_nodes       = var.gke_auth.shielded_nodes
  enable_intranode_visibility = var.enable_intranode_visibility

  master_auth {
    username = var.gke_auth.basic_auth ? "basic-admin" : ""
    password = var.gke_auth.basic_auth ? random_string.admin-password[0].result : ""

    client_certificate_config {
      issue_client_certificate = var.client_certificate
    }
  }

    private_cluster_config {
      enable_private_endpoint = false
      enable_private_nodes = true
      master_ipv4_cidr_block = var.subnet_cidr["k8s-private"]
    }

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods-private-range"
    services_secondary_range_name = "services-private-range"
  }
    maintenance_policy {
    recurring_window {
      recurrence = var.maintenance_window.recurrence
      start_time = var.maintenance_window.start_time
      end_time   = var.maintenance_window.end_time
    }
  }

  # Authoroized networks allowed to access the Master

  # master_authorized_networks_config {
  
  #   cidr_blocks {
  #     cidr_block = "${var.natgw_ip[0]}/32"
  #     display_name = "natgw"
  #   }
  #   cidr_blocks {
  #     cidr_block = "${var.natgw_ip[1]}/32"
  #     display_name = "natgw"
  #   }
  # }
  master_authorized_networks_config {
    dynamic "cidr_blocks" {
      for_each = var.gke_map.override["public_access_cidrs"]
      iterator = authorized_network
      content {
        cidr_block = authorized_network.value.cidr_block
        display_name = authorized_network.value.display_name
      }
    }
  }

    dynamic "resource_usage_export_config" {
    for_each = toset(var.resource_usage_export_config_parameters != null ? ["exec"] : [])
    content {
      enable_network_egress_metering = lookup(var.resource_usage_export_config_parameters, "enable_network_egress_metering")
      enable_resource_consumption_metering = lookup(var.resource_usage_export_config_parameters, "enable_resource_consumption_metering")
      bigquery_destination {
        dataset_id = lookup(var.resource_usage_export_config_parameters, "bigquery_destination.dataset_id")
      }
    }
  }

  lifecycle {
    ignore_changes = [
      initial_node_count, master_auth
    ]
  }
}

# K8s cluster node pool creation
resource "google_container_node_pool" "worker" {
  count      = contains(keys(var.gke_map),"ng") ? 1 : 0
  name       = lookup(var.gke_map.ng, "name", "${var.deploy_name}-${var.region}-ng-1" )
  location   = var.k8s_zonal == "" ? var.region : var.region_zone
  cluster    = google_container_cluster.primary[0].name
  node_count = 1
  version    = lookup(var.gke_map.override, "k8s_node_version", data.google_container_engine_versions.region.latest_node_version)

dynamic "autoscaling" {
    for_each = toset(var.autoscaling_parameters != null ? ["exec"] : [])
    content {
      min_node_count = lookup(var.autoscaling_parameters, "min_node_count")
      max_node_count = lookup(var.autoscaling_parameters, "max_node_count")
    }
  }
  management {
    auto_repair  = lookup(var.node_config, "node_auto_repair")
    auto_upgrade = lookup(var.node_config, "node_auto_upgrade" )
  }
  node_config {
    machine_type = lookup(var.gke_map.ng, "instance_type", "n2-highmem-2")
    image_type   = lookup(var.gke_map.ng, "image_type","COS")
    disk_size_gb = lookup(var.gke_map.ng, "disk_size", "2000")
    disk_type    = "pd-ssd"
    oauth_scopes = var.oauth_scopes

    shielded_instance_config {
      enable_secure_boot = lookup(var.node_config, "enable_secure_boot" )
    }
//    workload_metadata_config {
//      node_metadata = "GKE_METADATA_SERVER"
//    }


//    labels = {
//      cluster = var.label
//    }
     metadata = {
      ssh-keys                 = "${var.ssh_key}"
       disable-legacy-endpoints = "true"
     }
    tags = var.instance_tags
  }
  lifecycle {
    ignore_changes = [
      autoscaling.0.max_node_count, node_count
    ]
  }
}
###node group for devops###
resource "google_container_node_pool" "devops_nodegroup" {
  count      = contains(keys(var.gke_map),"devops") ? 1 : 0
  name       = lookup(var.gke_map.devops, "name", "${var.deploy_name}-${var.region}-ng-1" )
  location   = var.k8s_zonal == "" ? var.region : var.region_zone
  cluster    = google_container_cluster.primary[0].name
  node_count = 1
  version    = lookup(var.gke_map.override, "k8s_node_version", data.google_container_engine_versions.region.latest_node_version)
  management {
    auto_repair  = lookup(var.node_config, "node_auto_repair")
    auto_upgrade = lookup(var.node_config, "node_auto_upgrade" )
  }

dynamic "autoscaling" {
    for_each = toset(var.gke_map.devops["autoscaling_parameters"] != {} ? ["exec"] : [])
    content {
      min_node_count = lookup(var.gke_map.devops["autoscaling_parameters"], "min_node_count")
      max_node_count = lookup(var.gke_map.devops["autoscaling_parameters"], "max_node_count")
    }
  }
  
  node_config {
    machine_type = lookup(var.gke_map.devops, "instance_type", "n2-standard-2")
    labels =  {
      "k8s.jfrog.com/pool_type" = "devops"
      }
    image_type   = var.image_type
    disk_size_gb = lookup(var.gke_map.devops, "disk_size", "2000")
    disk_type    = "pd-ssd"
    oauth_scopes= var.oauth_scopes
      taint {
      effect = "NO_SCHEDULE"
      key    = "pool_type"
      value  = "devops"
    }
    shielded_instance_config {
      enable_secure_boot = lookup(var.node_config, "enable_secure_boot" )
    }
//    workload_metadata_config {
//      node_metadata = "GKE_METADATA_SERVER"
//    }

    metadata = {
      ssh-keys                 = "ubuntu:${var.ssh_key} ubuntu"
      disable-legacy-endpoints = "true"
    }
    tags = var.instance_tags
  }
  lifecycle {
    ignore_changes = [
      autoscaling.0.max_node_count, node_count
    ]
  }
}