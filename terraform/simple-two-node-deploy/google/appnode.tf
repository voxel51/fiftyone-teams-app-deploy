resource "google_compute_address" "appnode" {
  name = "fiftyone-appnode-ip"
}

resource "google_compute_instance" "fiftyone_appnode" {
  name         = "fiftyone-appnode"
  machine_type = var.fiftyone_appnode_machine_type

  boot_disk {
    initialize_params {
      image = data.google_compute_image.base_image.self_link
      size  = var.fiftyone_appnode_disk_gb
    }
  }

  metadata = {
    ssh-keys = "${file(var.google_ssh_key_file)}"
  }

  network_interface {
    network = var.fiftyone_network_name
    access_config {
      nat_ip = google_compute_address.appnode.address
    }
  }

  tags = ["fiftyone-app"]
}

resource "google_compute_firewall" "fiftyone-app" {
  name        = "fiftyone-app"
  network     = var.fiftyone_network_name
  description = "Allow access to fiftyone appnode for http/https"
  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = var.fiftyone_appnode_allowed
  target_tags   = ["fiftyone-app"]

}
