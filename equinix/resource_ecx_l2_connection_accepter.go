package equinix

import (
	"context"
	"fmt"
	"log"
	"time"

	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/equinix/ecx-go/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var ecxL2ConnectionAccepterSchemaNames = map[string]string{
	"ConnectionId":    "connection_id",
	"AccessKey":       "access_key",
	"SecretKey":       "secret_key",
	"Profile":         "aws_profile",
	"AWSConnectionID": "aws_connection_id",
}

var ecxL2ConnectionAccepterDescriptions = map[string]string{
	"ConnectionId":    "Identifier of layer 2 connection that will be accepted",
	"AccessKey":       "Access Key used to accept connection on provider side",
	"SecretKey":       "Secret Key used to accept connection on provider side",
	"Profile":         "AWS Profile Name for retrieving credentials from shared credentials file",
	"AWSConnectionID": "Identifier of a hosted Direct Connect connection on AWS side, applicable for accepter resource with connections to AWS only",
}

func resourceECXL2ConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceECXL2ConnectionAccepterCreate,
		ReadContext:   resourceECXL2ConnectionAccepterRead,
		DeleteContext: resourceECXL2ConnectionAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema:      createECXL2ConnectionAccepterResourceSchema(),
		Description: "Resource is used to accept Equinix Fabric layer 2 connection on provider side",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func createECXL2ConnectionAccepterResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		ecxL2ConnectionAccepterSchemaNames["ConnectionId"]: {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionAccepterDescriptions["ConnectionId"],
		},
		ecxL2ConnectionAccepterSchemaNames["AccessKey"]: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			Sensitive:    true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionAccepterDescriptions["AccessKey"],
		},
		ecxL2ConnectionAccepterSchemaNames["SecretKey"]: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			Sensitive:    true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionAccepterDescriptions["SecretKey"],
		},
		ecxL2ConnectionAccepterSchemaNames["Profile"]: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  ecxL2ConnectionAccepterDescriptions["Profile"],
		},
		ecxL2ConnectionAccepterSchemaNames["AWSConnectionID"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: ecxL2ConnectionAccepterDescriptions["AWSConnectionID"],
		},
	}
}

func resourceECXL2ConnectionAccepterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	req := ecx.L2ConnectionToConfirm{}
	creds, err := retrieveAWSCredentials(d)
	if err != nil {
		return diag.Errorf("error retrieving AWS credentials: %s", err)
	}
	req.AccessKey = ecx.String(creds.AccessKeyID)
	req.SecretKey = ecx.String(creds.SecretAccessKey)
	connID := d.Get(ecxL2ConnectionAccepterSchemaNames["ConnectionId"]).(string)
	if _, err := conf.ecx.ConfirmL2Connection(connID, req); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(connID)

	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			ecx.ConnectionStatusProvisioning,
			ecx.ConnectionStatusPendingApproval,
		},
		Target: []string{
			ecx.ConnectionStatusProvisioned,
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conf.ecx.GetL2Connection(connID)
			if err != nil {
				return nil, "", err
			}
			return resp, ecx.StringValue(resp.ProviderStatus), nil
		},
	}
	if _, err := createStateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for connection %q to be provisioned on provider side: %s", connID, err)
	}
	diags = append(diags, resourceECXL2ConnectionAccepterRead(ctx, d, m)...)
	return diags
}

func resourceECXL2ConnectionAccepterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	conn, err := conf.ecx.GetL2Connection(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if conn == nil || isStringInSlice(ecx.StringValue(conn.Status), []string{
		ecx.ConnectionStatusPendingDelete,
		ecx.ConnectionStatusDeprovisioning,
		ecx.ConnectionStatusDeprovisioned,
		ecx.ConnectionStatusDeleted,
	}) {
		d.SetId("")
		return diags
	}
	if err := updateECXL2ConnectionAccepterResource(conn, d); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceECXL2ConnectionAccepterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[WARN] [equinix_ecx_l2_connection_accepter] Will not delete ECX L2 connection (%s)"+
		"Terraform will remove this resource from the state file, however resources may remain.", d.Id())
	return nil
}

func updateECXL2ConnectionAccepterResource(conn *ecx.L2Connection, d *schema.ResourceData) error {
	if err := d.Set(ecxL2ConnectionAccepterSchemaNames["ConnectionId"], conn.UUID); err != nil {
		return fmt.Errorf("error reading connection UUID: %s", err)
	}
	creds, err := retrieveAWSCredentials(d)
	if err != nil {
		return fmt.Errorf("error retrieving AWS credentials: %s", err)
	}
	if err := d.Set(ecxL2ConnectionAccepterSchemaNames["AccessKey"], creds.AccessKeyID); err != nil {
		return fmt.Errorf("error reading AWS accessKeyID: %s", err)
	}
	if err := d.Set(ecxL2ConnectionAccepterSchemaNames["SecretKey"], creds.SecretAccessKey); err != nil {
		return fmt.Errorf("error reading AWS secretAccessKey: %s", err)
	}
	var awsConnectionID *string
	for _, action := range conn.Actions {
		if ecx.StringValue(action.OperationID) != "CONFIRM_CONNECTION" {
			continue
		}
		for _, actionData := range action.RequiredData {
			if ecx.StringValue(actionData.Key) != "awsConnectionId" {
				continue
			}
			awsConnectionID = actionData.Value
		}
	}
	if err := d.Set(ecxL2ConnectionAccepterSchemaNames["AWSConnectionID"], awsConnectionID); err != nil {
		return fmt.Errorf("error reading connection AWSConnectionID: %s", err)
	}
	return nil
}

func retrieveAWSCredentials(d *schema.ResourceData) (awsCredentials.Value, error) {
	credsProviders := []awsCredentials.Provider{
		&awsCredentials.StaticProvider{
			Value: awsCredentials.Value{
				AccessKeyID:     d.Get(ecxL2ConnectionAccepterSchemaNames["AccessKey"]).(string),
				SecretAccessKey: d.Get(ecxL2ConnectionAccepterSchemaNames["SecretKey"]).(string),
				SessionToken:    "",
			},
		},
		&awsCredentials.EnvProvider{},
		&awsCredentials.SharedCredentialsProvider{
			Filename: "",
			Profile:  d.Get(ecxL2ConnectionAccepterSchemaNames["Profile"]).(string),
		},
	}
	creds := awsCredentials.NewChainCredentials(credsProviders)
	return creds.Get()
}
