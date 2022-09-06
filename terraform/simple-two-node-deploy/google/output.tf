output "appnode_public_ip" {
  value = google_compute_address.appnode.address
}

output "dbnode_public_ip" {
  value = google_compute_address.dbnode.address
}
