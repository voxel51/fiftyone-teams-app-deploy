variable "google_image_family" {}
variable "google_image_project" {}
variable "google_project" {}
variable "google_region" {}
variable "google_ssh_key_file" {}
variable "google_zone" {}

variable "fiftyone_appnode_allowed" {}
variable "fiftyone_appnode_disk_gb" { type = number }
variable "fiftyone_appnode_machine_type" {}
variable "fiftyone_dbnode_allowed" {}
variable "fiftyone_dbnode_disk_gb" { type = number }
variable "fiftyone_dbnode_machine_type" {}
variable "fiftyone_network_name" { default = "default" }
