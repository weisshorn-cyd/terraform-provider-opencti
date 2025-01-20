// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weisshorn-cyd/gocti"
	"github.com/weisshorn-cyd/gocti/system"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

// NewRoleResource is a helper function to simplify the provider implementation.
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

// roleResource is the resource implementation.
type roleResource struct {
	client *gocti.OpenCTIAPIClient
}

// roleResourceModel maps the resource schema data.
type roleResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Capabilities types.List   `tfsdk:"capabilities"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"capabilities": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating roles")

	// Create new role
	createdRole, err := r.client.CreateRole(ctx, "id name", system.RoleAddInput{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role",
			"Could not create role, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Role created: %+v", createdRole))

	capabilitesAssigned := []string{}

	// Get capabilities to assign
	capabilities := []system.Capabilities{}
	for _, capability := range plan.Capabilities.Elements() {
		capabilities = append(capabilities, system.Capabilities(strings.Trim(capability.String(), "\"")))
		capabilitesAssigned = append(capabilitesAssigned, strings.Trim(capability.String(), "\""))
	}

	capabilitiesIDs, err := system.Capability{}.IDsByNames(ctx, r.client, capabilities)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving capabilities IDs",
			"Could not create role, unexpected error: "+err.Error(),
		)

		return
	}

	for _, capability := range capabilitiesIDs {
		if _, err := createdRole.AssignCapability(ctx, r.client, capability); err != nil {
			resp.Diagnostics.AddError(
				"Error assigning capabilities",
				"Could not create role, unexpected error: "+err.Error(),
			)

			return
		}
	}

	sort.Strings(capabilitesAssigned)

	capabilitiesAssignedList, diags := types.ListValueFrom(ctx, types.StringType, capabilitesAssigned)
	resp.Diagnostics.Append(diags...)

	plan = roleResourceModel{
		ID:           types.StringValue(createdRole.ID),
		Name:         types.StringValue(createdRole.Name),
		Capabilities: capabilitiesAssignedList,
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read role from opencti
	role, err := r.client.ReadRole(ctx, "id name capabilities {id name}", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti role", err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Role read: %+v", role))

	state.ID = types.StringValue(role.ID)
	state.Name = types.StringValue(role.Name)

	capabilities := []string{}
	for _, capability := range role.Capabilities {
		capabilities = append(capabilities, capability.Name)
	}

	sort.Strings(capabilities)

	capabilitiesList, diags := types.ListValueFrom(ctx, types.StringType, capabilities)
	resp.Diagnostics.Append(diags...)

	state.Capabilities = capabilitiesList

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.ReadRole(ctx, "id name capabilities {id name}", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti role", err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("Role to update: %+v", role))

	var capabilitiesPlan []string
	diags = plan.Capabilities.ElementsAs(ctx, &capabilitiesPlan, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	capabilitiesOfRole := []string{}

	// Get all capabilities IDs to assign
	capabilities := []system.Capabilities{}
	for _, capability := range capabilitiesPlan {
		capabilities = append(capabilities, system.Capabilities(capability))
	}

	capabilitiesIDs, err := system.Capability{}.IDsByNames(ctx, r.client, capabilities)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving capabilities IDs", err.Error(),
		)

		return
	}

	// Remove capabilities
	for _, capability := range role.Capabilities {
		if !slices.Contains(capabilitiesPlan, capability.Name) {
			tflog.Info(ctx, fmt.Sprintf("Removing capability: %s", capability.Name))

			if _, err := role.UnassignCapability(ctx, r.client, capability.ID); err != nil {
				resp.Diagnostics.AddError(
					"Error Unassigning OpenCTI Capability from Role", err.Error(),
				)

				return
			}
		} else {
			capabilitiesOfRole = append(capabilitiesOfRole, capability.Name)
		}
	}

	// Add capabilities
	for i, capability := range capabilitiesPlan {
		if !slices.Contains(capabilitiesOfRole, capability) {
			tflog.Info(ctx, fmt.Sprintf("Adding capability: %s", capability))

			if _, err := role.AssignCapability(ctx, r.client, capabilitiesIDs[i]); err != nil {
				resp.Diagnostics.AddError(
					"Error assigning capability to role", err.Error(),
				)

				return
			}

			capabilitiesOfRole = append(capabilitiesOfRole, capability)
		}
	}

	sort.Strings(capabilitiesOfRole)

	tflog.Debug(ctx, fmt.Sprintf("Capabilities: %v", capabilitiesOfRole))

	capabilitiesList, diags := types.ListValueFrom(ctx, types.StringType, capabilitiesOfRole)
	resp.Diagnostics.Append(diags...)

	plan.Capabilities = capabilitiesList

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteRole(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI Role",
			"Could not delete role, unexpected error: "+err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
