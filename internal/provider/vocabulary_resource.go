// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weisshorn-cyd/gocti"
	"github.com/weisshorn-cyd/gocti/entity"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &vocabularyResource{}
	_ resource.ResourceWithConfigure   = &vocabularyResource{}
	_ resource.ResourceWithImportState = &vocabularyResource{}
)

// NewVocabularyResource is a helper function to simplify the provider implementation.
func NewVocabularyResource() resource.Resource {
	return &vocabularyResource{}
}

// vocabularyResource is the resource implementation.
type vocabularyResource struct {
	client *gocti.OpenCTIAPIClient
}

// vocabularyResourceModel maps the resource schema data.
type vocabularyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Category    types.String `tfsdk:"category"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *vocabularyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vocabulary"
}

// Schema defines the schema for the resource.
func (r *vocabularyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"category": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *vocabularyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan vocabularyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating vocabulary")

	// Create new vocabulary
	createdVoc, err := r.client.CreateVocabulary(ctx, "id name description category { key }", entity.VocabularyAddInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Category:    plan.Category.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating vocabulary",
			"Could not create vocabulary, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("Vocabulary created: %+v", createdVoc))

	plan.ID = types.StringValue(createdVoc.ID)
	plan.Name = types.StringValue(createdVoc.Name)
	plan.Description = types.StringValue(createdVoc.Description)
	plan.Category = types.StringValue(createdVoc.Category.Key)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *vocabularyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state vocabularyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read vocabulary from opencti
	voc, err := r.client.ReadVocabulary(ctx, "id name description category { key }", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti vocabulary", err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("Vocabulary read: %+v", voc))

	state.ID = types.StringValue(voc.ID)
	state.Name = types.StringValue(voc.Name)
	state.Description = types.StringValue(voc.Description)
	state.Category = types.StringValue(voc.Category.Key)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vocabularyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan vocabularyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating vocabulary")

	// Create vocabulary with new values
	createdVoc, err := r.client.CreateVocabulary(ctx, "", entity.VocabularyAddInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Category:    plan.Category.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating vocabulary",
			"Could not create vocabulary, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("Vocabulary created: %+v", createdVoc))

	plan.ID = types.StringValue(createdVoc.ID)
	plan.Name = types.StringValue(createdVoc.Name)
	plan.Description = types.StringValue(createdVoc.Description)
	plan.Category = types.StringValue(createdVoc.Category.Key)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *vocabularyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state vocabularyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteVocabulary(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI Vocabulary",
			"Could not delete vocabulary, unexpected error: "+err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *vocabularyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gocti.OpenCTIAPIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gocti.OpenCTIAPIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *vocabularyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
