// Copyright (c) vdbe
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EncryptResource{}
	_ resource.ResourceWithImportState = &EncryptResource{}
)

func NewEncryptResource() resource.Resource {
	return &EncryptResource{}
}

// EncryptResource defines the resource implementation.
type EncryptResource EncryptDataSourceModel

// EncryptResourceModel describes the resource data model.
type EncryptResourceModel = EncryptDataSourceModel

func (r *EncryptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encrypt"
}

func (r *EncryptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				MarkdownDescription: "Base64 encoded value to be encrypted.",
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
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

func (d EncryptResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("value"),
			path.MatchRoot("value_base64"),
		),
	}
}

func (r *EncryptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EncryptResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

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
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EncryptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EncryptResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := data.SetDefaults(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Attribute values format", fmt.Sprintf("Failed to set defaults, got error: %s", err))
		return
	}

	tflog.Warn(ctx, "read a resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EncryptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EncryptResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// var old_data EncryptDataSourceModel
	// req.State.Get(ctx, &old_data)
	//
	// tflog.Trace(ctx, fmt.Sprintf("old value: %s", old_data.Value.ValueString()))
	// tflog.Trace(ctx, fmt.Sprintf("value: %s", data.Value.ValueString()))

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

	tflog.Trace(ctx, "updated a resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EncryptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EncryptResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "deleted a resource")
}

func (r *EncryptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
