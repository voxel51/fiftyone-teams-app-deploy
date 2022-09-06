data "google_compute_image" "base_image" {
  family  = var.google_image_family
  project = var.google_image_project
}
