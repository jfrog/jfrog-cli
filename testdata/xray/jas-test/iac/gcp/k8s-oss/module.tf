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
  min_master_version       = var.k8s_master_version == "" ? data.google_container_engine_versions.region.latest_master_version : var.k8s_master_version
  network                  = var.network
  subnetwork               = var.subnetwork
  logging_service          = var.logging_service
  monitoring_service       = var.monitoring_service
  enable_legacy_abac       = var.enable_legacy_abac
  remove_default_node_pool = "true"
  initial_node_count       = 1
  enable_shielded_nodes     = var.gke_auth.shielded_nodes
  enable_intranode_visibility = var.gke_auth.enable_intranode_visibility

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

  # Authoroized networks allowed to access the Master

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block = "82.81.195.5/32"
      display_name = "jfrog-office"
    }
    cidr_blocks {
      cidr_block = "52.8.67.255/32"
      display_name = "GlobalVpn"
    }
    cidr_blocks {
      cidr_block = "12.252.18.78/32"
      display_name = "US Office HA Public"
    }
    cidr_blocks {
      cidr_block = "52.9.243.19/32"
      display_name = "US IT AWS-NATGW"
    }
    cidr_blocks {
      cidr_block = "52.215.237.185/32"
      display_name = "EU IT AWS-NATGW"
    }
    cidr_blocks {
      cidr_block = "52.16.203.109/32"
      display_name = "GlobalVpn"
    }
    cidr_blocks {
      cidr_block = "146.148.8.199/32"
      display_name = "GCP jfrog-dev NAT"
    }
    cidr_blocks {
      cidr_block = "192.168.20.0/24" //should be 192.168.21.0/24
      display_name = "all_local"
    }
    cidr_blocks {
      cidr_block = "${var.natgw_ip[0]}/32"
      display_name = "natgw"
    }
    cidr_blocks {
      cidr_block = "${var.natgw_ip[1]}/32"
      display_name = "natgw"
    }
  }
  lifecycle {
    ignore_changes = [
      initial_node_count, master_authorized_networks_config, master_auth
    ]
  }
}

# K8s cluster node pool creation
resource "google_container_node_pool" "worker" {
  count      = var.module_enabled ? 1 : 0
  name       = var.override_ng_name == "" ? "${var.deploy_name}-${var.region}-ng-1" : var.override_ng_name
  location   = var.k8s_zonal == "" ? var.region : var.region_zone
  cluster    = google_container_cluster.primary[0].name
  node_count = 1
  version    = var.k8s_node_version == "" ? data.google_container_engine_versions.region.latest_node_version : var.k8s_node_version

  autoscaling {
    min_node_count = var.min_node_count
    max_node_count = var.max_node_count
  }

  management {
    auto_repair  = lookup(var.node_config, "node_auto_repair")
    auto_upgrade = lookup(var.node_config, "node_auto_upgrade" )
  }

  node_config {
    machine_type = var.worker_machine_type
    image_type   = var.image_type
    disk_size_gb = var.ng_disk_size_gb
    disk_type    = "pd-ssd"

    shielded_instance_config {
      enable_secure_boot = lookup(var.node_config, "enable_secure_boot" )
    }
//    workload_metadata_config {
//      node_metadata = "GKE_METADATA_SERVER"
//    }
//    oauth_scopes = [
//      "https://www.googleapis.com/auth/compute",
//      "https://www.googleapis.com/auth/devstorage.read_only",
//      "https://www.googleapis.com/auth/logging.write",
//      "https://www.googleapis.com/auth/monitoring",
//    ]

//    labels = {
//      cluster = var.label
//    }
//    metadata = {
//      ssh-keys                 = "ubuntu:${var.ssh_key} ubuntu"
//      disable-legacy-endpoints = "true"
//    }
    tags = var.instance_tags
  }
  lifecycle {
    ignore_changes = [
      autoscaling.0.max_node_count, node_count
    ]
  }
}