package equinix

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/equinix/ecx-go/v2"
	"github.com/equinix/rest-go"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var ecxL2ConnectionSchemaNames = map[string]string{
	"UUID":                "uuid",
	"Name":                "name",
	"ProfileUUID":         "profile_uuid",
	"Speed":               "speed",
	"SpeedUnit":           "speed_unit",
	"Status":              "status",
	"ProviderStatus":      "provider_status",
	"Notifications":       "notifications",
	"PurchaseOrderNumber": "purchase_order_number",
	"PortUUID":            "port_uuid",
	"DeviceUUID":          "device_uuid",
	"DeviceInterfaceID":   "device_interface_id",
	"VlanSTag":            "vlan_stag",
	"VlanCTag":            "vlan_ctag",
	"NamedTag":            "named_tag",
	"AdditionalInfo":      "additional_info",
	"ZSidePortUUID":       "zside_port_uuid",
	"ZSideVlanSTag":       "zside_vlan_stag",
	"ZSideVlanCTag":       "zside_vlan_ctag",
	"SellerRegion":        "seller_region",
	"SellerMetroCode":     "seller_metro_code",
	"AuthorizationKey":    "authorization_key",
	"RedundantUUID":       "redundant_uuid",
	"RedundancyType":      "redundancy_type",
	"SecondaryConnection": "secondary_connection",
}

var ecxL2ConnectionDescriptions = map[string]string{
	"UUID":                "Unique identifier of the connection",
	"Name":                "Connection name. An alpha-numeric 24 characters string which can include only hyphens and underscores",
	"ProfileUUID":         "Unique identifier of the service provider's service profile",
	"Speed":               "Speed/Bandwidth to be allocated to the connection",
	"SpeedUnit":           "Unit of the speed/bandwidth to be allocated to the connection",
	"Status":              "Connection provisioning status on Equinix Fabric side",
	"ProviderStatus":      "Connection provisioning status on service provider's side",
	"Notifications":       "A list of email addresses used for sending connection update notifications",
	"PurchaseOrderNumber": "Connection's purchase order number to reflect on the invoice",
	"PortUUID":            "Unique identifier of the buyer's port from which the connection would originate",
	"DeviceUUID":          "Unique identifier of the Network Edge virtual device from which the connection would originate",
	"DeviceInterfaceID":   "Identifier of network interface on a given device, used for a connection. If not specified then first available interface will be selected",
	"VlanSTag":            "S-Tag/Outer-Tag of the connection, a numeric character ranging from 2 - 4094",
	"VlanCTag":            "C-Tag/Inner-Tag of the connection, a numeric character ranging from 2 - 4094",
	"NamedTag":            "The type of peering to set up in case when connecting to Azure Express Route. One of Public, Private, Microsoft, Manual",
	"AdditionalInfo":      "One or more additional information key-value objects",
	"ZSidePortUUID":       "Unique identifier of the port on the remote side (z-side)",
	"ZSideVlanSTag":       "S-Tag/Outer-Tag of the connection on the remote side (z-side)",
	"ZSideVlanCTag":       "C-Tag/Inner-Tag of the connection on the remote side (z-side)",
	"SellerRegion":        "The region in which the seller port resides",
	"SellerMetroCode":     "The metro code that denotes the connection’s remote side (z-side)",
	"AuthorizationKey":    "Text field used to authorize connection on the provider side. Value depends on a provider service profile used for connection",
	"RedundantUUID":       "Unique identifier of the redundant connection, applicable for HA connections",
	"RedundancyType":      "Connection redundancy type, applicable for HA connections. Either primary or secondary",
	"SecondaryConnection": "Definition of secondary connection for redundant, HA connectivity",
}

var ecxL2ConnectionAdditionalInfoSchemaNames = map[string]string{
	"Name":  "name",
	"Value": "value",
}

var ecxL2ConnectionAdditionalInfoDescriptions = map[string]string{
	"Name":  "Additional information key",
	"Value": "Additional information value",
}

func resourceECXL2Connection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceECXL2ConnectionCreate,
		ReadContext:   resourceECXL2ConnectionRead,
		UpdateContext: resourceECXL2ConnectionUpdate,
		DeleteContext: resourceECXL2ConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: createECXL2ConnectionResourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Description: "Resource allows creation and management of Equinix Fabric	layer 2 connections",
	}
}

func createECXL2ConnectionResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		ecxL2ConnectionSchemaNames["UUID"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: ecxL2ConnectionDescriptions["UUID"],
		},
		ecxL2ConnectionSchemaNames["Name"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 24),
			Description:  ecxL2ConnectionDescriptions["Name"],
		},
		ecxL2ConnectionSchemaNames["ProfileUUID"]: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			AtLeastOneOf: []string{ecxL2ConnectionSchemaNames["ProfileUUID"], ecxL2ConnectionSchemaNames["ZSidePortUUID"]},
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionDescriptions["ProfileUUID"],
		},
		ecxL2ConnectionSchemaNames["Speed"]: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(1),
			Description:  ecxL2ConnectionDescriptions["Speed"],
		},
		ecxL2ConnectionSchemaNames["SpeedUnit"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"MB", "GB"}, false),
			Description:  ecxL2ConnectionDescriptions["SpeedUnit"],
		},
		ecxL2ConnectionSchemaNames["Status"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: ecxL2ConnectionDescriptions["Status"],
		},
		ecxL2ConnectionSchemaNames["ProviderStatus"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: ecxL2ConnectionDescriptions["ProviderStatus"],
		},
		ecxL2ConnectionSchemaNames["Notifications"]: {
			Type:     schema.TypeSet,
			Required: true,
			ForceNew: true,
			MinItems: 1,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: stringIsEmailAddress(),
			},
			Description: ecxL2ConnectionDescriptions["Notifications"],
		},
		ecxL2ConnectionSchemaNames["PurchaseOrderNumber"]: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(1, 30),
			Description:  ecxL2ConnectionDescriptions["PurchaseOrderNumber"],
		},
		ecxL2ConnectionSchemaNames["PortUUID"]: {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ValidateFunc:  validation.StringIsNotEmpty,
			AtLeastOneOf:  []string{ecxL2ConnectionSchemaNames["PortUUID"], ecxL2ConnectionSchemaNames["DeviceUUID"]},
			ConflictsWith: []string{ecxL2ConnectionSchemaNames["DeviceUUID"]},
			Description:   ecxL2ConnectionDescriptions["PortUUID"],
		},
		ecxL2ConnectionSchemaNames["DeviceUUID"]: {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ValidateFunc:  validation.StringIsNotEmpty,
			ConflictsWith: []string{ecxL2ConnectionSchemaNames["PortUUID"]},
			Description:   ecxL2ConnectionDescriptions["DeviceUUID"],
		},
		ecxL2ConnectionSchemaNames["DeviceInterfaceID"]: {
			Type:          schema.TypeInt,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{ecxL2ConnectionSchemaNames["PortUUID"]},
			Description:   ecxL2ConnectionDescriptions["DeviceInterfaceID"],
		},
		ecxL2ConnectionSchemaNames["VlanSTag"]: {
			Type:          schema.TypeInt,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ValidateFunc:  validation.IntBetween(2, 4092),
			RequiredWith:  []string{ecxL2ConnectionSchemaNames["PortUUID"]},
			ConflictsWith: []string{ecxL2ConnectionSchemaNames["DeviceUUID"]},
			Description:   ecxL2ConnectionDescriptions["VlanSTag"],
		},
		ecxL2ConnectionSchemaNames["VlanCTag"]: {
			Type:          schema.TypeInt,
			Optional:      true,
			ForceNew:      true,
			ValidateFunc:  validation.IntBetween(2, 4092),
			ConflictsWith: []string{ecxL2ConnectionSchemaNames["DeviceUUID"]},
			Description:   ecxL2ConnectionDescriptions["VlanCTag"],
		},
		ecxL2ConnectionSchemaNames["NamedTag"]: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"Private", "Public", "Microsoft", "Manual"}, false),
			Description:  ecxL2ConnectionDescriptions["NamedTag"],
		},
		ecxL2ConnectionSchemaNames["AdditionalInfo"]: {
			Type:        schema.TypeSet,
			Optional:    true,
			ForceNew:    true,
			MinItems:    1,
			Description: ecxL2ConnectionDescriptions["AdditionalInfo"],
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					ecxL2ConnectionAdditionalInfoSchemaNames["Name"]: {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  ecxL2ConnectionAdditionalInfoDescriptions["Name"],
					},
					ecxL2ConnectionAdditionalInfoSchemaNames["Value"]: {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  ecxL2ConnectionAdditionalInfoDescriptions["Value"],
					},
				},
			},
		},
		ecxL2ConnectionSchemaNames["ZSidePortUUID"]: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionDescriptions["ZSidePortUUID"],
		},
		ecxL2ConnectionSchemaNames["ZSideVlanSTag"]: {
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(2, 4092),
			Description:  ecxL2ConnectionDescriptions["ZSideVlanSTag"],
		},
		ecxL2ConnectionSchemaNames["ZSideVlanCTag"]: {
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(2, 4092),
			Description:  ecxL2ConnectionDescriptions["ZSideVlanCTag"],
		},
		ecxL2ConnectionSchemaNames["SellerRegion"]: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionDescriptions["SellerRegion"],
		},
		ecxL2ConnectionSchemaNames["SellerMetroCode"]: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: stringIsMetroCode(),
			Description:  ecxL2ConnectionDescriptions["SellerMetroCode"],
		},
		ecxL2ConnectionSchemaNames["AuthorizationKey"]: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionDescriptions["AuthorizationKey"],
		},
		ecxL2ConnectionSchemaNames["RedundantUUID"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: ecxL2ConnectionDescriptions["RedundantUUID"],
		},
		ecxL2ConnectionSchemaNames["RedundancyType"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: ecxL2ConnectionDescriptions["RedundancyType"],
		},
		ecxL2ConnectionSchemaNames["SecondaryConnection"]: {
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Description: ecxL2ConnectionDescriptions["SecondaryConnection"],
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					ecxL2ConnectionSchemaNames["UUID"]: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["UUID"],
					},
					ecxL2ConnectionSchemaNames["Name"]: {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 24),
						Description:  ecxL2ConnectionDescriptions["Name"],
					},
					ecxL2ConnectionSchemaNames["ProfileUUID"]: {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  ecxL2ConnectionDescriptions["ProfileUUID"],
					},
					ecxL2ConnectionSchemaNames["Speed"]: {
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						ValidateFunc: validation.IntAtLeast(1),
						Description:  ecxL2ConnectionDescriptions["Speed"],
					},
					ecxL2ConnectionSchemaNames["SpeedUnit"]: {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringInSlice([]string{"MB", "GB"}, false),
						RequiredWith: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["Speed"]},
						Description:  ecxL2ConnectionDescriptions["SpeedUnit"],
					},
					ecxL2ConnectionSchemaNames["Status"]: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["Status"],
					},
					ecxL2ConnectionSchemaNames["ProviderStatus"]: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["ProviderStatus"],
					},
					ecxL2ConnectionSchemaNames["PortUUID"]: {
						Type:         schema.TypeString,
						ForceNew:     true,
						Optional:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						AtLeastOneOf: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["PortUUID"],
							ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["DeviceUUID"]},
						ConflictsWith: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["DeviceUUID"]},
						Description:   ecxL2ConnectionDescriptions["PortUUID"],
					},
					ecxL2ConnectionSchemaNames["DeviceUUID"]: {
						Type:          schema.TypeString,
						ForceNew:      true,
						Optional:      true,
						ValidateFunc:  validation.StringIsNotEmpty,
						ConflictsWith: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["PortUUID"]},
						Description:   ecxL2ConnectionDescriptions["DeviceUUID"],
					},
					ecxL2ConnectionSchemaNames["DeviceInterfaceID"]: {
						Type:          schema.TypeInt,
						Optional:      true,
						Computed:      true,
						ForceNew:      true,
						ConflictsWith: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["PortUUID"]},
						Description:   ecxL2ConnectionDescriptions["DeviceInterfaceID"],
					},
					ecxL2ConnectionSchemaNames["VlanSTag"]: {
						Type:          schema.TypeInt,
						ForceNew:      true,
						Optional:      true,
						Computed:      true,
						ValidateFunc:  validation.IntBetween(2, 4092),
						RequiredWith:  []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["PortUUID"]},
						ConflictsWith: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["DeviceUUID"]},
						Description:   ecxL2ConnectionDescriptions["VlanSTag"],
					},
					ecxL2ConnectionSchemaNames["VlanCTag"]: {
						Type:          schema.TypeInt,
						ForceNew:      true,
						Optional:      true,
						ValidateFunc:  validation.IntBetween(2, 4092),
						ConflictsWith: []string{ecxL2ConnectionSchemaNames["SecondaryConnection"] + ".0." + ecxL2ConnectionSchemaNames["DeviceUUID"]},
						Description:   ecxL2ConnectionDescriptions["VlanCTag"],
					},
					ecxL2ConnectionSchemaNames["ZSidePortUUID"]: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["ZSidePortUUID"],
					},
					ecxL2ConnectionSchemaNames["ZSideVlanSTag"]: {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["ZSideVlanSTag"],
					},
					ecxL2ConnectionSchemaNames["ZSideVlanCTag"]: {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["ZSideVlanCTag"],
					},
					ecxL2ConnectionSchemaNames["SellerRegion"]: {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  ecxL2ConnectionDescriptions["SellerRegion"],
					},
					ecxL2ConnectionSchemaNames["SellerMetroCode"]: {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						ValidateFunc: stringIsMetroCode(),
						Description:  ecxL2ConnectionDescriptions["SellerMetroCode"],
					},
					ecxL2ConnectionSchemaNames["AuthorizationKey"]: {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  ecxL2ConnectionDescriptions["AuthorizationKey"],
					},
					ecxL2ConnectionSchemaNames["RedundantUUID"]: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["RedundantUUID"],
					},
					ecxL2ConnectionSchemaNames["RedundancyType"]: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: ecxL2ConnectionDescriptions["RedundancyType"],
					},
				},
			},
		},
	}
}

func resourceECXL2ConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	primary, secondary := createECXL2Connections(d)
	var primaryID *string
	var err error
	if secondary != nil {
		primaryID, _, err = conf.ecx.CreateL2RedundantConnection(*primary, *secondary)
	} else {
		primaryID, err = conf.ecx.CreateL2Connection(*primary)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ecx.StringValue(primaryID))
	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			ecx.ConnectionStatusProvisioning,
			ecx.ConnectionStatusPendingAutoApproval,
		},
		Target: []string{
			ecx.ConnectionStatusProvisioned,
			ecx.ConnectionStatusPendingApproval,
			ecx.ConnectionStatusPendingBGPPeering,
			ecx.ConnectionStatusPendingProviderVlan,
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      2 * time.Second,
		MinTimeout: 2 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conf.ecx.GetL2Connection(d.Id())
			if err != nil {
				return nil, "", err
			}
			return resp, ecx.StringValue(resp.Status), nil
		},
	}
	if _, err := createStateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for connection (%s) to be created: %s", d.Id(), err)
	}
	diags = append(diags, resourceECXL2ConnectionRead(ctx, d, m)...)
	return diags
}

func resourceECXL2ConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	var err error
	var primary *ecx.L2Connection
	var secondary *ecx.L2Connection

	primary, err = conf.ecx.GetL2Connection(d.Id())
	if err != nil {
		return diag.Errorf("cannot fetch primary connection due to %v", err)
	}
	if isStringInSlice(ecx.StringValue(primary.Status), []string{
		ecx.ConnectionStatusPendingDelete,
		ecx.ConnectionStatusDeprovisioning,
		ecx.ConnectionStatusDeprovisioned,
		ecx.ConnectionStatusDeleted,
	}) {
		d.SetId("")
		return nil
	}
	if ecx.StringValue(primary.RedundantUUID) != "" {
		secondary, err = conf.ecx.GetL2Connection(ecx.StringValue(primary.RedundantUUID))
		if err != nil {
			return diag.Errorf("cannot fetch secondary connection due to %v", err)
		}
	}
	if err := updateECXL2ConnectionResource(primary, secondary, d); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceECXL2ConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	supportedChanges := []string{ecxL2ConnectionSchemaNames["Name"],
		ecxL2ConnectionSchemaNames["Speed"],
		ecxL2ConnectionSchemaNames["SpeedUnit"]}
	primaryChanges := getResourceDataChangedKeys(supportedChanges, d)
	primaryUpdateReq := conf.ecx.NewL2ConnectionUpdateRequest(d.Id())
	if err := fillFabricL2ConnectionUpdateRequest(primaryUpdateReq, primaryChanges).Execute(); err != nil {
		return diag.FromErr(err)
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["RedundantUUID"]); ok {
		secondaryChanges := getResourceDataListElementChanges(supportedChanges, ecxL2ConnectionSchemaNames["SecondaryConnection"], 0, d)
		secondaryUpdateReq := conf.ecx.NewL2ConnectionUpdateRequest(v.(string))
		if err := fillFabricL2ConnectionUpdateRequest(secondaryUpdateReq, secondaryChanges).Execute(); err != nil {
			return diag.FromErr(err)
		}
	}
	diags = append(diags, resourceECXL2ConnectionRead(ctx, d, m)...)
	return diags
}

func resourceECXL2ConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	if err := conf.ecx.DeleteL2Connection(d.Id()); err != nil {
		restErr, ok := err.(rest.Error)
		if ok {
			//IC-LAYER2-4021 = Connection already deleted
			if hasApplicationErrorCode(restErr.ApplicationErrors, "IC-LAYER2-4021") {
				return diags
			}
		}
		return diag.FromErr(err)
	}
	//remove secondary connection, don't fail on error as there is no partial state on delete
	if redID, ok := d.GetOk(ecxL2ConnectionSchemaNames["RedundantUUID"]); ok {
		if err := conf.ecx.DeleteL2Connection(redID.(string)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Failed to remove secondary connection with UUID %q", redID.(string)),
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath(ecxL2ConnectionSchemaNames["RedundantUUID"]),
			})
		}
	}
	deleteStateConf := &resource.StateChangeConf{
		Pending: []string{
			ecx.ConnectionStatusDeprovisioning,
		},
		Target: []string{
			ecx.ConnectionStatusPendingDelete,
			ecx.ConnectionStatusDeprovisioned,
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      2 * time.Second,
		MinTimeout: 2 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conf.ecx.GetL2Connection(d.Id())
			if err != nil {
				return nil, "", err
			}
			return resp, ecx.StringValue(resp.Status), nil
		},
	}
	if _, err := deleteStateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for connection (%s) to be removed: %s", d.Id(), err)
	}
	return diags
}

func createECXL2Connections(d *schema.ResourceData) (*ecx.L2Connection, *ecx.L2Connection) {
	var primary, secondary *ecx.L2Connection
	primary = &ecx.L2Connection{}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["Name"]); ok {
		primary.Name = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["ProfileUUID"]); ok {
		primary.ProfileUUID = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["Speed"]); ok {
		primary.Speed = ecx.Int(v.(int))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["SpeedUnit"]); ok {
		primary.SpeedUnit = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["Notifications"]); ok {
		primary.Notifications = expandSetToStringList(v.(*schema.Set))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["PurchaseOrderNumber"]); ok {
		primary.PurchaseOrderNumber = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["PortUUID"]); ok {
		primary.PortUUID = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["DeviceUUID"]); ok {
		primary.DeviceUUID = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["DeviceInterfaceID"]); ok {
		primary.DeviceInterfaceID = ecx.Int(v.(int))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["VlanSTag"]); ok {
		primary.VlanSTag = ecx.Int(v.(int))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["VlanCTag"]); ok {
		primary.VlanCTag = ecx.Int(v.(int))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["NamedTag"]); ok {
		primary.NamedTag = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["AdditionalInfo"]); ok {
		primary.AdditionalInfo = expandECXL2ConnectionAdditionalInfo(v.(*schema.Set))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["ZSidePortUUID"]); ok {
		primary.ZSidePortUUID = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["ZSideVlanSTag"]); ok {
		primary.ZSideVlanSTag = ecx.Int(v.(int))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["ZSideVlanCTag"]); ok {
		primary.ZSideVlanCTag = ecx.Int(v.(int))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["SellerRegion"]); ok {
		primary.SellerRegion = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["SellerMetroCode"]); ok {
		primary.SellerMetroCode = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["AuthorizationKey"]); ok {
		primary.AuthorizationKey = ecx.String(v.(string))
	}
	if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["SecondaryConnection"]); ok {
		secondary = expandECXL2ConnectionSecondary(v.([]interface{}))
	}
	return primary, secondary
}

func updateECXL2ConnectionResource(primary *ecx.L2Connection, secondary *ecx.L2Connection, d *schema.ResourceData) error {
	if err := d.Set(ecxL2ConnectionSchemaNames["UUID"], primary.UUID); err != nil {
		return fmt.Errorf("error reading UUID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["Name"], primary.Name); err != nil {
		return fmt.Errorf("error reading Name: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["ProfileUUID"], primary.ProfileUUID); err != nil {
		return fmt.Errorf("error reading ProfileUUID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["Speed"], primary.Speed); err != nil {
		return fmt.Errorf("error reading Speed: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["SpeedUnit"], primary.SpeedUnit); err != nil {
		return fmt.Errorf("error reading SpeedUnit: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["Status"], primary.Status); err != nil {
		return fmt.Errorf("error reading Status: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["ProviderStatus"], primary.ProviderStatus); err != nil {
		return fmt.Errorf("error reading ProviderStatus: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["Notifications"], primary.Notifications); err != nil {
		return fmt.Errorf("error reading Notifications: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["PurchaseOrderNumber"], primary.PurchaseOrderNumber); err != nil {
		return fmt.Errorf("error reading PurchaseOrderNumber: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["PortUUID"], primary.PortUUID); err != nil {
		return fmt.Errorf("error reading PortUUID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["DeviceUUID"], primary.DeviceUUID); err != nil {
		return fmt.Errorf("error reading DeviceUUID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["VlanSTag"], primary.VlanSTag); err != nil {
		return fmt.Errorf("error reading VlanSTag: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["VlanCTag"], primary.VlanCTag); err != nil {
		return fmt.Errorf("error reading VlanCTag: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["NamedTag"], primary.NamedTag); err != nil {
		return fmt.Errorf("error reading NamedTag: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["AdditionalInfo"], flattenECXL2ConnectionAdditionalInfo(primary.AdditionalInfo)); err != nil {
		return fmt.Errorf("error reading AdditionalInfo: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["ZSidePortUUID"], primary.ZSidePortUUID); err != nil {
		return fmt.Errorf("error reading ZSidePortUUID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["ZSideVlanSTag"], primary.ZSideVlanSTag); err != nil {
		return fmt.Errorf("error reading ZSideVlanSTag: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["ZSideVlanCTag"], primary.ZSideVlanCTag); err != nil {
		return fmt.Errorf("error reading ZSideVlanCTag: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["SellerRegion"], primary.SellerRegion); err != nil {
		return fmt.Errorf("error reading SellerRegion: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["SellerMetroCode"], primary.SellerMetroCode); err != nil {
		return fmt.Errorf("error reading SellerMetroCode: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["AuthorizationKey"], primary.AuthorizationKey); err != nil {
		return fmt.Errorf("error reading AuthorizationKey: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["RedundantUUID"], primary.RedundantUUID); err != nil {
		return fmt.Errorf("error reading RedundantUUID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionSchemaNames["RedundancyType"], primary.RedundancyType); err != nil {
		return fmt.Errorf("error reading RedundancyType: %s", err)
	}
	if secondary != nil {
		var prevSecondary *ecx.L2Connection
		if v, ok := d.GetOk(ecxL2ConnectionSchemaNames["SecondaryConnection"]); ok {
			prevSecondary = expandECXL2ConnectionSecondary(v.([]interface{}))
		}
		if err := d.Set(ecxL2ConnectionSchemaNames["SecondaryConnection"], flattenECXL2ConnectionSecondary(prevSecondary, secondary)); err != nil {
			return fmt.Errorf("error reading SecondaryConnection: %s", err)
		}
	}
	return nil
}

func flattenECXL2ConnectionSecondary(previous, conn *ecx.L2Connection) interface{} {
	transformed := make(map[string]interface{})
	transformed[ecxL2ConnectionSchemaNames["UUID"]] = conn.UUID
	transformed[ecxL2ConnectionSchemaNames["Name"]] = conn.Name
	transformed[ecxL2ConnectionSchemaNames["ProfileUUID"]] = conn.ProfileUUID
	transformed[ecxL2ConnectionSchemaNames["Speed"]] = conn.Speed
	transformed[ecxL2ConnectionSchemaNames["SpeedUnit"]] = conn.SpeedUnit
	transformed[ecxL2ConnectionSchemaNames["Status"]] = conn.Status
	transformed[ecxL2ConnectionSchemaNames["ProviderStatus"]] = conn.ProviderStatus
	transformed[ecxL2ConnectionSchemaNames["PortUUID"]] = conn.PortUUID
	transformed[ecxL2ConnectionSchemaNames["DeviceUUID"]] = conn.DeviceUUID
	transformed[ecxL2ConnectionSchemaNames["DeviceInterfaceID"]] = conn.DeviceInterfaceID
	if previous != nil && ecx.IntValue(previous.DeviceInterfaceID) != 0 {
		transformed[ecxL2ConnectionSchemaNames["DeviceInterfaceID"]] = previous.DeviceInterfaceID
	}
	transformed[ecxL2ConnectionSchemaNames["VlanSTag"]] = conn.VlanSTag
	transformed[ecxL2ConnectionSchemaNames["VlanCTag"]] = conn.VlanCTag
	transformed[ecxL2ConnectionSchemaNames["ZSidePortUUID"]] = conn.ZSidePortUUID
	transformed[ecxL2ConnectionSchemaNames["ZSideVlanSTag"]] = conn.ZSideVlanSTag
	transformed[ecxL2ConnectionSchemaNames["ZSideVlanCTag"]] = conn.ZSideVlanCTag
	transformed[ecxL2ConnectionSchemaNames["SellerRegion"]] = conn.SellerRegion
	transformed[ecxL2ConnectionSchemaNames["SellerMetroCode"]] = conn.SellerMetroCode
	transformed[ecxL2ConnectionSchemaNames["AuthorizationKey"]] = conn.AuthorizationKey
	transformed[ecxL2ConnectionSchemaNames["RedundantUUID"]] = conn.RedundantUUID
	transformed[ecxL2ConnectionSchemaNames["RedundancyType"]] = conn.RedundancyType
	return []interface{}{transformed}
}

func expandECXL2ConnectionSecondary(conns []interface{}) *ecx.L2Connection {
	if len(conns) < 1 {
		log.Printf("[WARN] resource_ecx_l2_connection expanding empty secondary connection collection")
		return nil
	}
	conn := conns[0].(map[string]interface{})
	transformed := ecx.L2Connection{}
	if v, ok := conn[ecxL2ConnectionSchemaNames["Name"]]; ok {
		transformed.Name = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["ProfileUUID"]]; ok && !isEmpty(v) {
		transformed.ProfileUUID = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["Speed"]]; ok && !isEmpty(v) {
		transformed.Speed = ecx.Int(v.(int))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["SpeedUnit"]]; ok && !isEmpty(v) {
		transformed.SpeedUnit = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["PortUUID"]]; ok && !isEmpty(v) {
		transformed.PortUUID = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["DeviceUUID"]]; ok && !isEmpty(v) {
		transformed.DeviceUUID = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["DeviceInterfaceID"]]; ok && !isEmpty(v) {
		transformed.DeviceInterfaceID = ecx.Int(v.(int))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["VlanSTag"]]; ok && !isEmpty(v) {
		transformed.VlanSTag = ecx.Int(v.(int))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["VlanCTag"]]; ok && !isEmpty(v) {
		transformed.VlanCTag = ecx.Int(v.(int))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["SellerRegion"]]; ok && !isEmpty(v) {
		transformed.SellerRegion = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["SellerMetroCode"]]; ok && !isEmpty(v) {
		transformed.SellerMetroCode = ecx.String(v.(string))
	}
	if v, ok := conn[ecxL2ConnectionSchemaNames["AuthorizationKey"]]; ok && !isEmpty(v) {
		transformed.AuthorizationKey = ecx.String(v.(string))
	}
	return &transformed
}

func flattenECXL2ConnectionAdditionalInfo(infos []ecx.L2ConnectionAdditionalInfo) interface{} {
	transformed := make([]interface{}, 0, len(infos))
	for _, info := range infos {
		transformed = append(transformed, map[string]interface{}{
			ecxL2ConnectionAdditionalInfoSchemaNames["Name"]:  info.Name,
			ecxL2ConnectionAdditionalInfoSchemaNames["Value"]: info.Value,
		})
	}
	return transformed
}

func expandECXL2ConnectionAdditionalInfo(infos *schema.Set) []ecx.L2ConnectionAdditionalInfo {
	transformed := make([]ecx.L2ConnectionAdditionalInfo, 0, infos.Len())
	for _, info := range infos.List() {
		infoMap := info.(map[string]interface{})
		transformed = append(transformed, ecx.L2ConnectionAdditionalInfo{
			Name:  ecx.String(infoMap[ecxL2ConnectionAdditionalInfoSchemaNames["Name"]].(string)),
			Value: ecx.String(infoMap[ecxL2ConnectionAdditionalInfoSchemaNames["Value"]].(string)),
		})
	}
	return transformed
}

func fillFabricL2ConnectionUpdateRequest(updateReq ecx.L2ConnectionUpdateRequest, changes map[string]interface{}) ecx.L2ConnectionUpdateRequest {
	for change, changeValue := range changes {
		switch change {
		case ecxL2ConnectionSchemaNames["Name"]:
			updateReq.WithName(changeValue.(string))
		case ecxL2ConnectionSchemaNames["Speed"]:
			updateReq.WithSpeed(changeValue.(int))
		case ecxL2ConnectionSchemaNames["SpeedUnit"]:
			updateReq.WithSpeedUnit(changeValue.(string))
		}
	}
	return updateReq
}
