Terraform provider registry API generator for Google Cloud Storage
=====================================================================
Since version 0.13 of Hashicorp Terraform, terraform supports the notion of a provider registry. When you
create your own provider, you can make it accessible via [https://registry.terraform.io](https://registry.terraform.io).

At the moment there is no support for private registry implementation available, except for the definition
of the [wire protocol of the registry](https://www.terraform.io/docs/internals/provider-registry-protocol.html).

This utility creates a static terraform provider registry on Google Cloud Storage. It does this by
generating the required API response documents and storing them in the correct location.

## How to create your own terraform provider registry?
To create your own private provider registry on Google Cloud. you need to:

1. create a bucket to host a website
2. build your own terraform provider
3. generate the api documents
4. reference the provider in your private registry

## Create your private terraform provider registry
To create your own terraform provider registry, you can use this [terraform template](./example/terraform/main.tf)
It will:

- reserve a public ip address
- create a gcs bucket named `<project-name>-tf-registry`
- add a DNS record for `registry.<your-domain-name>`
- create a certificate for `registry.<your-domain-name>`
- forward all https requests to the bucket
- redirect http requests to https

To only thing you need is a GCP project, and a DNS manage zone. To use this template to create your template, type:

```shell
cd example/terraform
terraform init
read -p "project id:" TF_VAR_project_id
read -p "dns zone name:" TF_VAR_dns_managed_zone
terraform apply
```
After a few minutes, you should be able to connect to the registry:

```shell
DOMAIN_NAME=$(cloud dns managed-zones \
    describe $TF_VAR_dns_managed_zone \
    --format 'value(dns_name)' | \
    sed -e 's/\.$//' )

REGISTRY_URL=https://registry.$DOMAIN_NAME

curl $REGISTRY_URL
````

## Build your custom terraform provider
Now you can build your custom terraform provider. In this example, we will use the [sentry](https://github.com/jianyuan/terraform-provider-sentry.git)
provider build by [Jian Yuan](https://jianyuan.io).

```shell
git clone https://github.com/jianyuan/terraform-provider-sentry.git
cd terraform-provider-sentry
git checkout v0.6.0
read -p  "pgp key fingerprint:" PGP_FINGERPRINT
TF_REGISTRY_BUCKET=${TF_VAR_project_id}-tf-registry
goreleaser release --rm-dist
"
```

## Upload the binaries to the Google Storage Bucket
To upload the binaries into your storage bucket, type:

```shell
cd dist
gsutil cp -r . gs://{$TF_VAR_project}-tf-registry/binaries/jianyuan/
```

Alternatively, you can add the following code to your `goreleaser.yaml`:
```yaml
blobs:
  - provider: gs
    bucket: '{{ .Env.TF_REGISTRY_BUCKET }}'
    folder: 'binaries/{{ .Env.PROVIDER_NAMESPACE }}/{{ .ProjectName }}/{{ .Tag }}'
```

This will automatically upload the binaries into the bucket.

## Generate terraform provider registry API documents
Finally, to generate the required terraform provider registry API documents, type:

```sh
go get github.com/mollie/tf-provider-registry-api-generator
$GOPATH/bin/tf-provider-registry-api-generator \
  --bucket-name $TF_REGISTRY_BUCKET \
  --prefix binaries/jianyuan/terraform-provider-sentry/v0.6.0/ \
  --protocols 5.0 \
  --namespace jianyuan \
  --fingerprint $PGP_FINGERPRINT \
  --url $REGISTRY_URL
```

Alternatively, you can add the following code to your `goreleaser.yaml`:

```yaml
publishers:
  - name: generate-tf-api-documents
    cmd: >-
      tf-provider-registry-api-generator
      --use-default-credentials
      --bucket-name '{{ .Env.TF_REGISTRY_BUCKET }}'
      --prefix 'binaries/{{ .Env.PROVIDER_NAMESPACE }}/{{ .ProjectName }}/{{ .Tag }}'
      --namespace '{{ .Env.PROVIDER_NAMESPACE }}'
      --fingerprint '{{ .Env.GPG_FINGERPRINT }}'
      --url '{{ .Env.REGISTRY_URL }}'
      --protocols '{{ .Env.PROVIDER_PROTOCOLS }}'
    env:
      - "PATH={{ .Env.PATH }}"
      - "GOOGLE_APPLICATION_CREDENTIALS={{ .Env.GOOGLE_APPLICATION_CREDENTIALS }}"
```


## Access the generated terraform provider registry API documents
The generator generates three document types:
1. the discovery document
2. the provider versions document
3. the provider download metadata document

### discovery document
to view the discovery document, type:
```sh
$ curl $REGISTRY_URL/.well-known/terraform.json
{
  "providers.v1": "/v1/providers/"
}

### provider versions document
to view the available versions of the provider, type:

```sh
$ curl $REGISTRY_URL/v1/providers/jianyuan/sentry/versions
{
  "versions": [
    {
      "version": "0.6.0",
      "protocols": [
        "5.0"
      ],
      "platforms": [
        {
          "os": "darwin",
          "arch": "amd64"
        },
        {
          "os": "linux",
          "arch": "amd64"
        }
      ]
    }
  ]
}
```

### provider download metadata document
To retrieve the download metadata for a particular platform, type:\
```shell
$ curl $REGISTRY_URL/v1/providers/jianyuan/sentry/0.6.0/download/darwin/amd64
{
  "protocols": [
    "5.0"
  ],
  "os": "darwin",
  "arch": "amd64",
  "filename": "terraform-provider-sentry_0.6.0_darwin_amd64.zip",
  "download_url": "$REGISTRY_URL/binaries/jianyuan/terraform-provider-sentry/v0.6.0/terraform-provider-sentry_0.6.0_darwin_amd64.zip",
  "shasums_url": "$REGISTRY_URL/binaries/jianyuan/terraform-provider-sentry/v0.6.0/terraform-provider-sentry_0.6.0_SHA256SUMS",
  "shasums_signature_url": "$REGISTRY_URL/binaries/jianyuan/terraform-provider-sentry/v0.6.0/terraform-provider-sentry_0.6.0_SHA256SUMS.sig",
  "shasum": "a2c5881ea67e1c397cb26c6162d81829e058d5a993801bcb69df9982412d27e9",
  "signing_keys": {
    "gpg_public_keys": [
      {
        "key_id": "8B15B898C0AA84DC7A7B0E46B851229EAFE0F521",
        "ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----...-----END PGP PUBLIC KEY BLOCK-----",
        "trust_signature": "",
        "source": "",
        "source_url": null
      }
    ]
  }
```

### Use the terraform provider in your private registry
To use the provider from your private registry, create the following file:

```shell
terraform {
    required_providers {
        sentry = {
            source = "$DOMAIN_NAME/jianyuan/sentry"
            version = "0.6.0"
        }
    }
}
```

and type:

```shell
$ terraform init

Initializing provider plugins...
- Finding latest version of $DOMAIN_NAME/jianyuan/sentry...
- Installing $DOMAIN_NAME/jianyuan/sentry v0.6.0...
- Installed $DOMAIN_NAMEv/jianyuan/sentry v0.6.0 (self-signed, key ID B64689ABE6ED9C52)

Partner and community providers are signed by their developers.
If you'd like to know more about provider signing, you can read about it here:
https://www.terraform.io/docs/plugins/signing.html

The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, we recommend adding version constraints in a required_providers block
in your configuration, with the constraint strings suggested below.

* $DOMAIN_NAME/jianyuan/sentry: version = "~> 0.6.0"

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.

```

## All-in-one-go release specification
All the steps of releasing your provider into a private registry can be combined in a single
goreleaser specification. Checkout the example [goreleaser.yaml](./example/goreleaser.yaml).

## conclusion
Although no private provider registry implementation is available at the moment, it is quite easy
to generate the documents to implement the required protocol to create a provider registry from a
static website.
