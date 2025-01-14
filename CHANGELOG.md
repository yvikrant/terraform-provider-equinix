## 1.3.0 (UNRELEASED)

IMPROVEMENTS:

- `equinix_ecx_l2_connection` resources can now be imported
- `equinix_ecx_l2_connection_accepter` resources can now be imported
- `equinix_ecx_l2_serviceprofile` resources can now be imported
- `equinix_network_acl_template` resources can now be imported
- `equinix_network_bgp` resources can now be imported
- `equinix_network_device_link` resources can now be imported
- `equinix_network_ssh_key` resources can now be imported
- `equinix_network_ssh_user` resources can now be imported

## 1.2.0 (April 27, 2021)

FEATURES:

- **New Resource**: `equinix_network_device_link` ([#43](https://github.com/equinix/terraform-provider-equinix/issues/43))

## 1.1.0 (April 09, 2021)

BUG FIXES:

- creation of Equinix Fabric layer2 redundant connection from a single device
is now possible by specifying same `deviceUUID` argument for both primary and
secondary connection. API logic of Fabric is reflected accordingly in client module

FEATURES:

- **New Data source**: `equinix_ecx_l2_sellerprofiles`: ([#40](https://github.com/equinix/terraform-provider-equinix/issues/40))
- **New Resource**: `equinix_network_ssh_key` ([#25](https://github.com/equinix/terraform-provider-equinix/issues/25))
- **New Resource**: `equinix_network_acl_template` ([#19](https://github.com/equinix/terraform-provider-equinix/issues/19))
- **New Resource**: `equinix_network_bgp` ([#16](https://github.com/equinix/terraform-provider-equinix/issues/16))
- **New Data source**: `equinix_network_account` ([#13](https://github.com/equinix/terraform-provider-equinix/issues/13))
- **New Data source**: `equinix_network_device_type` ([#13](https://github.com/equinix/terraform-provider-equinix/issues/13))
- **New Data source**: `equinix_network_device_software` ([#13](https://github.com/equinix/terraform-provider-equinix/issues/13))
- **New Data source**: `equinix_network_device_platform` ([#13](https://github.com/equinix/terraform-provider-equinix/issues/13))
- **New Resource**: `equinix_network_device` ([#4](https://github.com/equinix/terraform-provider-equinix/issues/4))
- **New Resource**: `equinix_network_ssh_user` ([#4](https://github.com/equinix/terraform-provider-equinix/issues/4))

ENHANCEMENTS:

- Equinix provider: setting `TF_LOG` to `TRACE` enables logging of Equinix REST
API requests and responses
- resource/equinix_ecx_l2_connection: internal representation of secondary connection
block has changed from Set to List. This enables plan to better communicate secondary
connection changes and allows using `Optional` + `Computed` schema options
([#39](https://github.com/equinix/terraform-provider-equinix/issues/39))
- resource/equinix_ecx_l2_connection: added additional arguments for `secondary_connection`
([#18](https://github.com/equinix/terraform-provider-equinix/issues/18)):
  - `speed`
  - `speed_unit`
  - `profile_uuid`
  - `authorization_key`
  - `seller_metro_code`
  - `seller_region`

## 1.0.3 (January 07, 2021)

ENHANCEMENTS:

- resource/equinix_ecx_l2_connection_accepter: AWS credentials can be provided
using additional ways: environmental variables and shared configuration files
- resource/equinix_ecx_l2_service_profile: introduced schema validations,
updated acceptance tests and resource documentation

BUG FIXES:

- resource/equinix_ecx_l2_connection_accepter: creation waits for PROVISIONED provider
status of the connection before succeeding
([#37](https://github.com/equinix/terraform-provider-equinix/issues/37))

## 1.0.2 (November 17, 2020)

ENHANCEMENTS:

- resource/equinix_ecx_l2_connection_accepter: creation awaits for desired
connection provider state before succeeding ([#26](https://github.com/equinix/terraform-provider-equinix/issues/26))

BUG FIXES:

- resource/equinix_ecx_l2_connection: z-side port identifier, vlan C-tag and vlan
S-tag for secondary connection are properly populated with values from the Fabric
([#24](https://github.com/equinix/terraform-provider-equinix/issues/24))

## 1.0.1 (November 09, 2020)

NOTES:

- this version of module started to use `equinix/rest-go` client
for any REST interactions with Equinix APIs

ENHANCEMENTS:

- resource/equinix_ecx_l2_connection_accepter: added `aws_connection_id` attribute
([#22](https://github.com/equinix/terraform-provider-equinix/issues/22))
- resource/equinix_ecx_l2_connection: removal awaits for desired
connection state before succeeding ([#21](https://github.com/equinix/terraform-provider-equinix/issues/21))
- resource/equinix_ecx_l2_connection: added `device_interface_id` argument ([#18](https://github.com/equinix/terraform-provider-equinix/issues/18))
- resource/equinix_ecx_l2_connection: added `provider_status` and
 `redundancy_type` attributes ([#14](https://github.com/equinix/terraform-provider-equinix/issues/14))
- resource/equinix_ecx_l2_connection: creation awaits for desired
connection state before succeeding ([#15](https://github.com/equinix/terraform-provider-equinix/issues/15))

## 1.0.0 (September 02, 2020)

NOTES:

- first version of official Equinix Terraform provider

FEATURES:

- **New Resource**: `equinix_ecx_l2_connection`
- **New Resource**: `equinix_ecx_l2_connection_accepter`
- **New Resource**: `equinix_ecx_l2_serviceprofile`
- **New Data Source**: `equinix_ecx_port`
- **New Data Source**: `equinix_ecx_l2_sellerprofile`
