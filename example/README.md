setup your private tf provider registry
=======================================

1. Create the private tf provider registry
```sh
cd terraform
terraform init
read -p "project id:" TF_VAR_project_id
read -p "dns zone name:" TF_VAR_dns_managed_zone
terraform apply
```

1. Build your 3rd party provider

```sh
git clone https://github.com/jianyuan/terraform-provider-sentry.git
cd terraform-provider-sentry
git checkout v0.6.0
read -p  "pgp key fingerprint:" PGP_FINGERPRINT
TF_REGISTRY_BUCKET=${TF_VAR_project_id}-tf-registry
goreleaser release -f ../goreleaser.yaml --rm-dist
"
```

1. Generate TF Registry API documents

```sh
REGISTRY_URL=https://registry.$(gcloud dns managed-zones describe $TF_VAR_dns_managed_zone) --format 'value(dns_name)'

go get github.com/mollie/tf-provider-registry-api-generator
tf-provider-registry-api-generator \
  --bucket-name $TF_REGISTRY_BUCKET \
  --prefix binaries/terraform-provider-sentry/v0.6.0/ \
  --namespace jianyuan \
  --fingerprint $PGP_FINGERPRINT \
  --url $REGISTRY_URL
```
