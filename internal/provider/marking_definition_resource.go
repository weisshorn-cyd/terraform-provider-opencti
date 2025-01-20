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
	_ resource.Resource                = &markingDefinitionResource{}
	_ resource.ResourceWithConfigure   = &markingDefinitionResource{}
	_ resource.ResourceWithImportState = &markingDefinitionResource{}
)

// NewmarkingDefinitionDefinitionResource is a helper function to simplify the provider implementation.
func NewMarkingDefinitionResource() resource.Resource {
	return &markingDefinitionResource{}
}

// markingDefinitionResource is the resource implementation.
type markingDefinitionResource struct {
	client *gocti.OpenCTIAPIClient
}

// markingDefinitionResourceModel maps the resource schema data.
type markingDefinitionResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DefinitionType types.String `tfsdk:"definition_type"`
	Definition     types.String `tfsdk:"definition"`
	XOpenctiOrder  types.Int32  `tfsdk:"x_opencti_order"`
	XOpenctiColor  types.String `tfsdk:"x_opencti_color"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *markingDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marking_definition"
}

// Schema defines the schema for the resource.
func (r *markingDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"definition_type": schema.StringAttribute{
				Required: true,
			},
			"definition": schema.StringAttribute{
				Required: true,
			},
			"x_opencti_order": schema.Int32Attribute{
				Required: true,
			},
			"x_opencti_color": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *markingDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan markingDefinitionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating marking definition")

	// Create new markingDefinition
	createdMarking, err := r.client.CreateMarkingDefinition(ctx, "", entity.MarkingDefinitionAddInput{
		DefinitionType: plan.DefinitionType.ValueString(),
		Definition:     plan.Definition.ValueString(),
		XOpenctiOrder:  int(plan.XOpenctiOrder.ValueInt32()),
		XOpenctiColor:  plan.XOpenctiColor.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating marking definition",
			"Could not create marking definition, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Marking definition created: %+v", createdMarking))

	plan.ID = types.StringValue(createdMarking.ID)
	plan.DefinitionType = types.StringValue(createdMarking.DefinitionType)
	plan.Definition = types.StringValue(createdMarking.Definition)
	plan.XOpenctiOrder = types.Int32Value(int32(createdMarking.XOpenctiOrder))
	plan.XOpenctiColor = types.StringValue(createdMarking.XOpenctiColor)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *markingDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state markingDefinitionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get marking definitions from opencti
	marking, err := r.client.ReadMarkingDefinition(ctx, "id definition_type definition x_opencti_order x_opencti_color", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti marking definition",
			"Could not read opencti marking definition, unexpected error: "+err.Error(),
		)

		return
	}

	state.ID = types.StringValue(marking.ID)
	state.DefinitionType = types.StringValue(marking.DefinitionType)
	state.Definition = types.StringValue(marking.Definition)
	state.XOpenctiOrder = types.Int32Value(int32(marking.XOpenctiOrder))
	state.XOpenctiColor = types.StringValue(marking.XOpenctiColor)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *markingDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan markingDefinitionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating marking definition")

	// Create new markingDefinition
	createdMarking, err := r.client.CreateMarkingDefinition(ctx, "", entity.MarkingDefinitionAddInput{
		DefinitionType: plan.DefinitionType.ValueString(),
		Definition:     plan.Definition.ValueString(),
		XOpenctiOrder:  int(plan.XOpenctiOrder.ValueInt32()),
		XOpenctiColor:  plan.XOpenctiColor.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating marking definition",
			"Could not create marking definition, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Marking definition created: %+v", createdMarking))

	plan.ID = types.StringValue(createdMarking.ID)
	plan.DefinitionType = types.StringValue(createdMarking.DefinitionType)
	plan.Definition = types.StringValue(createdMarking.Definition)
	plan.XOpenctiOrder = types.Int32Value(int32(createdMarking.XOpenctiOrder))
	plan.XOpenctiColor = types.StringValue(createdMarking.XOpenctiColor)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *markingDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state markingDefinitionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteMarkingDefinition(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI Marking Definition",
			"Could not delete marking definition, unexpected error: "+err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *markingDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *markingDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
