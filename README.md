Terraform provider registry API generator for Google Cloud Storage
=====================================================================
Since version 0.13 of Hashicorp Terraform, terraform supports the notion
of a provider registry. When you create your own provider, you
can make it accessible via [https://registry.terraform.io](https://registry.terraform.io).

There is no support for private registries, except for the definition
of the [wire protocol of the registry](https://www.terraform.io/docs/internals/provider-registry-protocol.html).

This utility creates a static terraform provider registry on Google Cloud Storage. It does this by 
generating the required API response documents and storing them in the correct location.
