---
layout: "equinix"
page_title: "Equinix: equinix_ecx_l2_connection_accepter"
subcategory: ""
description: |-
  Provides Equinix Fabric Layer 2 connection accepter resource
---

# Resource: equinix_ecx_l2_connection_accepter

Resource `equinix_ecx_l2_connection_accepter` is used to accept Equinix Fabric 
layer 2 connection on provider side.

Resource leverages Equinix Fabric integration with service providers.
Currently supported providers are:

* `AWS` (AWS Direct Connect)

## Example Usage

```hcl
resource "equinix_ecx_l2_connection_accepter" "accepter" {
  connection_id = equinix_ecx_l2_connection.awsConn.id
}
```

## AWS Authentication

The `equinix_ecx_l2_connection_accepter` resource offers flexible means of providing
AWS credentials. The following methods are supported and evaluated in a given order:

* static credentials - uses `access_key` and `secret_key` resource arguments
* environmental variables - uses `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
 environmental variables
* shared credentials/configuration file - uses [AWS credentials or configuration
file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).
This method also supports profile configuration by setting `aws_profile`
argument or `AWS_PROFILE` environmental variable

**Please note** that it is not
recommended to keep credentials in any Terraform configuration.

## Argument Reference

* `connection_id` - (Required) Identifier of Layer 2 connection that will be accepted
* `access_key` - (Optional) Access Key used to accept connection on provider side
* `secret_key` - (Optional) Secret Key used to accept connection on provider side
* `aws_profile` - (Optional) AWS Profile Name for retrieving credentials from
 shared credentials file

## Attribute Reference

* `aws_connection_id` - Identifier of a hosted Direct Connect connection on AWS side,
applicable for accepter resource with connections to AWS only

## Import

This resource can be imported using an existing ID:

```sh
terraform import equinix_ecx_l2_connection_accepter.example {existing_id}
```
