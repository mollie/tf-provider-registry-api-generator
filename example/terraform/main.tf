variable project_id {
  description = "to create the terraform private registry in"
  type = string
}

variable dns_managed_zone {
  description = "name of the dns managed zone to register the dns name in"
  type = string
}

provider google {
  project = var.project_id
}

resource "google_compute_global_forwarding_rule" "tf_registry" {
  project    = var.project_id
  name       = "tf-registry"
  target     = google_compute_target_https_proxy.tf_registry.self_link
  ip_address = google_compute_global_address.tf_registry.address
  port_range = "443"
}

resource "google_compute_global_address" "tf_registry" {
  project = var.project_id
  name    = "tf-registry"
}

resource "google_dns_record_set" "registry" {
  project      = var.project_id
  name         = format("registry.%s", data.google_dns_managed_zone.tf_registry.dns_name)
  type         = "A"
  ttl          = 300
  managed_zone = data.google_dns_managed_zone.tf_registry.name
  rrdatas      = [google_compute_global_address.tf_registry.address]
}

data "google_dns_managed_zone" "tf_registry" {
  name = var.dns_managed_zone
}

resource "google_compute_managed_ssl_certificate" "tf_registry" {
  project  = var.project_id
  name     = "tf-registry"
  managed {
    domains = [
      google_dns_record_set.registry.name
    ]
  }
}

resource "google_storage_bucket" "tf_registry" {
  project       = var.project_id
  name          = format("%s-tf-registry", var.project_id)
  storage_class = "MULTI_REGIONAL"
  location      = "EU"

}

resource "google_compute_backend_bucket" "tf_registry" {
  project     = var.project_id
  name        = google_storage_bucket.tf_registry.name
  description = "bucket for terraform private registry"
  bucket_name = google_storage_bucket.tf_registry.name
  enable_cdn  = false
}

resource "google_compute_url_map" "http_to_https" {
  project = var.project_id
  name    = "http-to-https"
  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

resource "google_compute_target_http_proxy" "http_to_https" {
  project = var.project_id
  name    = "http-to-https"
  url_map = google_compute_url_map.http_to_https.self_link
}

resource "google_compute_global_forwarding_rule" "http_to_https" {
  project    = var.project_id
  name       = "http-to-https"
  target     = google_compute_target_http_proxy.http_to_https.self_link
  ip_address = google_compute_global_address.tf_registry.address
  port_range = "80"
}

resource "google_compute_url_map" "tf_registry" {
  project         = var.project_id
  name            = "tf-registry"
  default_service = google_compute_backend_bucket.tf_registry.id
}

resource "google_compute_target_https_proxy" "tf_registry" {
  project = var.project_id
  name    = "tf-registry"
  url_map = google_compute_url_map.tf_registry.self_link
  ssl_certificates = [
    google_compute_managed_ssl_certificate.tf_registry.id
  ]
}
