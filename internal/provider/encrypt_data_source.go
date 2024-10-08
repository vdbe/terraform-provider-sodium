// Copyright (c) vdbe
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vdbe/terraform-provider-sodium/internal/encryption"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &EncryptDataSource{}

func NewEncryptedDataSource() datasource.DataSource {
	return &EncryptDataSource{}
}

type EncryptDataSource struct{}

type EncryptDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	PublicKeyBase64 types.String `tfsdk:"public_key_base64"`
	Value           types.String `tfsdk:"value"`
	ValueBase64     types.String `tfsdk:"value_base64"`
	EncryptedBase64 types.String `tfsdk:"encrypted_base64"`
	Encrypted       types.String `tfsdk:"encrypted"`
	Base64Encode    types.Bool   `tfsdk:"base64_encode"`
}

func (d *EncryptDataSourceModel) SetDefaults(ctx context.Context) (err error) {
	// Set Value/ValueBase64 based on eachother
	tflog.Warn(ctx, fmt.Sprintf("value is null: %v", d.Value.IsNull()))
	tflog.Warn(ctx, fmt.Sprintf("value is unknown: %v", d.Value.IsUnknown()))
	if d.Value.IsNull() || d.Value.IsUnknown() {
		value_base64 := d.ValueBase64.ValueString()
		value_bytes, e := base64.StdEncoding.DecodeString(value_base64)
		if e != nil {
			err = fmt.Errorf("unable to decode `value_base64`, got error: %s", e)
			return
		}

		value := string(value_bytes[:])
		d.Value = types.StringValue(value)
	} else {
		value := d.Value.ValueString()
		value_base64 := base64.StdEncoding.EncodeToString([]byte(value))

		d.ValueBase64 = types.StringValue(value_base64)
	}

	// base64_encode (defaults to `false`)
	if d.Base64Encode.IsNull() || d.Base64Encode.IsUnknown() {
		d.Base64Encode = types.BoolValue(false)
	}

	return
}

func (d *EncryptDataSourceModel) Encrypt() (err error) {
	base64_encode := d.Base64Encode.ValueBool()

	// Get secret in the correct format
	var secret string
	if base64_encode {
		secret = d.ValueBase64.ValueString()
	} else {
		secret = d.Value.ValueString()
	}

	// Encrypt secret
	pub_key_base64 := d.PublicKeyBase64.ValueString()
	res := encryption.Encrypt(&pub_key_base64, []byte(secret))
	if res.Err != nil {
		err = fmt.Errorf("failed to encrypt the provided value, got error: %s", res.Err)
		return
	}
	d.EncryptedBase64 = types.StringValue(res.Encoded)
	d.Encrypted = types.StringValue(string(res.Raw))

	// Generate sha256 for id
	h := sha256.New()
	h.Write([]byte(pub_key_base64))
	h.Write([]byte("."))
	h.Write([]byte(secret))
	id := hex.EncodeToString(h.Sum(nil))
	d.Id = types.StringValue(id)

	return
}

func EncryptDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Encrypt the value using public key encryption",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier",
				Computed:            true,
			},
			"public_key_base64": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded public key to encrypt the value",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value to be encrypted",
				Sensitive:           true,
				Optional:            true,
				Computed:            true,
			},
			"value_base64": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded value to be encrypted",
				Sensitive:           true,
				Optional:            true,
				Computed:            true,
			},
			"base64_encode": schema.BoolAttribute{
				MarkdownDescription: "Base64 encoded value to be encrypted. Default `false`",
				Optional:            true,
			},
			"encrypted": schema.StringAttribute{
				MarkdownDescription: "Encrypted value",
				Computed:            true,
				Sensitive:           true,
			},
			"encrypted_base64": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded encrypted value",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *EncryptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encrypt"
}

func (d *EncryptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EncryptDataSourceSchema()
}

func (d EncryptDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("value"),
			path.MatchRoot("value_base64"),
		),
	}
}

func (d *EncryptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *EncryptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EncryptDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := data.SetDefaults(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Attribute values format", fmt.Sprintf("Failed to set defaults, got error: %s", err))
		return
	}

	err = data.Encrypt()
	if err != nil {
		resp.Diagnostics.AddError("Encrypt", fmt.Sprintf("Failed to encrypt, got error: %s", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
