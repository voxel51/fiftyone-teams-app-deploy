resource "google_compute_address" "dbnode" {
  name = "fiftyone-dbnode-ip"
}

resource "google_compute_instance" "fiftyone_dbnode" {
  name         = "fiftyone-dbnode"
  machine_type = var.fiftyone_dbnode_machine_type

  boot_disk {
    initialize_params {
      image = data.google_compute_image.base_image.self_link
      size  = var.fiftyone_dbnode_disk_gb
    }
  }

  metadata = {
    ssh-keys = "${file(var.google_ssh_key_file)}"
  }

  network_interface {
    network = var.fiftyone_network_name
    access_config {
      nat_ip = google_compute_address.dbnode.address
    }
  }

  tags = ["fiftyone-db"]
}

resource "google_compute_firewall" "fiftyone-db" {
  name        = "fiftyone-db"
  network     = var.fiftyone_network_name
  description = "Allow access to fiftyone dbnode for mongodb"
  allow {
    protocol = "tcp"
    ports    = ["27017"]
  }

  source_ranges = var.fiftyone_dbnode_allowed
  target_tags   = ["fiftyone-db"]

}
