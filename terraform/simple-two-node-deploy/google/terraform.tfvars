google_image_family  = "debian-11"
google_image_project = "debian-cloud"
google_project       = "your-project-here"
google_region        = "us-east1"
google_ssh_key_file  = "public-ssh-keys"
google_zone          = "us-east1-b"

# all hosts are required access for letsencrypt certificates
fiftyone_appnode_allowed      = ["0.0.0.0/0"]
fiftyone_appnode_disk_gb      = "20"
fiftyone_appnode_machine_type = "n2-standard-2"
fiftyone_dbnode_allowed       = ["98.31.1.253/32", "68.48.243.169/32"]
fiftyone_dbnode_disk_gb       = "200"
fiftyone_dbnode_machine_type  = "n2-standard-16"
