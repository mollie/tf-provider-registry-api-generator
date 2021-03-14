setup your private tf provider registry
=======================================

## Create the private tf provider registry
```sh
cd terraform
terraform init
read -p "project id:" TF_VAR_project_id
read -p "dns zone name:" TF_VAR_dns_managed_zone
terraform apply
```

## Build your 3rd party provider

```sh
git clone https://github.com/jianyuan/terraform-provider-sentry.git
cd terraform-provider-sentry
git checkout v0.6.0
read -p  "pgp key fingerprint:" PGP_FINGERPRINT
TF_REGISTRY_BUCKET=${TF_VAR_project_id}-tf-registry
goreleaser release -f ../goreleaser.yaml --rm-dist
"
```

## Generate TF Registry API documents

```sh
REGISTRY_URL=https://registry.$(gcloud dns managed-zones \
     describe $TF_VAR_dns_managed_zone \
    --format 'value(dns_name)' | \
  sed -e 's/\.$//')

go get github.com/mollie/tf-provider-registry-api-generator
tf-provider-registry-api-generator \
  --bucket-name $TF_REGISTRY_BUCKET \
  --prefix binaries/terraform-provider-sentry/v0.6.0/ \
  --namespace jianyuan \
  --fingerprint $PGP_FINGERPRINT \
  --protocols 5.0 \
  --url $REGISTRY_URL
```

## all steps in one go
You can also include all required steps in 1 single [goreleaser speification](./goreleaser.yaml).