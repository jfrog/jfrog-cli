# The following outputs allow authentication and connectivity to the GKE Cluster.
output "client_certificate" {
  value = google_container_cluster.primary.*.master_auth.0.client_certificate
}

output "client_key" {
  value = element(
    concat(
      google_container_cluster.primary.*.master_auth.0.client_key,
      [""],
    ),
    0,
  )
}

output "cluster_ca_certificate" {
  value = element(
    concat(
      google_container_cluster.primary.*.master_auth.0.cluster_ca_certificate,
      [""],
    ),
    0,
  )
}

output "cluster_name" {
  value = element(concat(google_container_cluster.primary.*.name, [""]), 0)
}

output "cluster_ip" {
  value = element(concat(google_container_cluster.primary.*.endpoint, [""]), 0)
}

output "cluster_username" {
  value = element(
    concat(
      google_container_cluster.primary.*.master_auth.0.username,
      [""],
    ),
    0,
  )
}

output "cluster_password" {
  value = element(
    concat(
      google_container_cluster.primary.*.master_auth.0.password,
      [""],
    ),
    0,
  )
  sensitive = true
}

