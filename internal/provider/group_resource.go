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
	"github.com/weisshorn-cyd/gocti/graphql"
	"github.com/weisshorn-cyd/gocti/system"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

// NewGroupResource is a helper function to simplify the provider implementation.
func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// groupResource is the resource implementation.
type groupResource struct {
	client *gocti.OpenCTIAPIClient
}

// groupResourceModel maps the resource schema data.
type groupResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Roles              types.List   `tfsdk:"roles"`
	AllowedMarking     types.List   `tfsdk:"allowed_marking"`
	MaxConfidenceLevel types.Int32  `tfsdk:"max_confidence_level"`
	AutoNewMarking     types.Bool   `tfsdk:"auto_new_marking"`
	DefaultAssignation types.Bool   `tfsdk:"default_assignation"`
	LastUpdated        types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"roles": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"allowed_marking": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"max_confidence_level": schema.Int32Attribute{
				Required: true,
			},
			"auto_new_marking": schema.BoolAttribute{
				Required: true,
			},
			"default_assignation": schema.BoolAttribute{
				Required: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating group")

	// Create new group
	createdGroup, err := r.client.CreateGroup(ctx, "id name description default_assignation auto_new_marking group_confidence_level { max_confidence }", system.GroupAddInput{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		DefaultAssignation: plan.DefaultAssignation.ValueBool(),
		AutoNewMarking:     plan.AutoNewMarking.ValueBool(),
		GroupConfidenceLevel: graphql.ConfidenceLevelInput{
			MaxConfidence: int(plan.MaxConfidenceLevel.ValueInt32()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			"Could not create group, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Group created: %+v", createdGroup))

	existingRoles, err := r.client.ListRoles(ctx, "", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing roles",
			"Could not create group, unexpected error: "+err.Error(),
		)

		return
	}

	rolesAssigned := []string{}

	// Assign the roles
	for _, role := range plan.Roles.Elements() {
		tflog.Info(ctx, fmt.Sprintf("Assigning role %s to group %s", role.String(), createdGroup.Name))

		for _, remoteRole := range existingRoles {
			if strings.Trim(role.String(), "\"") == remoteRole.Name {
				if _, err := createdGroup.AssignRole(ctx, r.client, remoteRole.ID); err != nil {
					resp.Diagnostics.AddError(
						"Error assigning role to group",
						"Could not create group, unexpected error: "+err.Error(),
					)

					return
				}

				rolesAssigned = append(rolesAssigned, remoteRole.Name)

				break
			}
		}
	}

	sort.Strings(rolesAssigned)

	// Assign the marking definitions
	existingMarkings, err := r.client.ListMarkingDefinitions(ctx, "id definition_type definition x_opencti_order x_opencti_color", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing markings",
			"Could not create group, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Existing markings: %+v", existingMarkings))

	markingsAssigned := []string{}

	for _, marking := range plan.AllowedMarking.Elements() {
		tflog.Info(ctx, fmt.Sprintf("Assigning marking %s to group %s", marking, createdGroup.Name))

		for _, remoteMarking := range existingMarkings {
			if strings.Trim(marking.String(), "\"") == remoteMarking.Definition {
				if _, err := createdGroup.AssignMarkingDefinition(ctx, r.client, remoteMarking.ID); err != nil {
					resp.Diagnostics.AddError(
						"Error assigning marking to group",
						"Could not create group, unexpected error: "+err.Error(),
					)

					return
				}

				tflog.Debug(ctx, fmt.Sprintf("Assigning markings: %+v", remoteMarking))
				markingsAssigned = append(markingsAssigned, remoteMarking.Definition)

				break
			}
		}
	}

	sort.Strings(markingsAssigned)

	rolesAssignedList, diags := types.ListValueFrom(ctx, types.StringType, rolesAssigned)
	resp.Diagnostics.Append(diags...)

	markingsAllowedList, diags := types.ListValueFrom(ctx, types.StringType, markingsAssigned)
	resp.Diagnostics.Append(diags...)

	tflog.Info(ctx, fmt.Sprintf("Group: %+v", createdGroup))
	tflog.Debug(ctx, fmt.Sprintf("Roles assigned : %+v", rolesAssigned))
	tflog.Debug(ctx, fmt.Sprintf("Markings assigned : %+v", markingsAssigned))

	plan = groupResourceModel{
		ID:                 types.StringValue(createdGroup.ID),
		Name:               types.StringValue(createdGroup.Name),
		Description:        types.StringValue(createdGroup.Description),
		Roles:              rolesAssignedList,
		AllowedMarking:     markingsAllowedList,
		MaxConfidenceLevel: types.Int32Value(int32(createdGroup.GroupConfidenceLevel.MaxConfidence)),
		AutoNewMarking:     types.BoolValue(createdGroup.AutoNewMarking),
		DefaultAssignation: types.BoolValue(createdGroup.DefaultAssignation),
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
func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read group from opencti
	group, err := r.client.ReadGroup(ctx, "id name description roles { edges { node {id name} } } allowed_marking {id definition_type definition} default_assignation auto_new_marking group_confidence_level { max_confidence }", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti group", err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Group read: %+v", group))

	// Parse the roles
	roles := []string{}
	for _, role := range group.Roles.Edges {
		roles = append(roles, role.Node.Name)
	}

	sort.Strings(roles)

	rolesList, diags := types.ListValueFrom(ctx, types.StringType, roles)
	resp.Diagnostics.Append(diags...)

	// Parse the markings
	markings := []string{}
	for _, marking := range group.AllowedMarking {
		markings = append(markings, marking.Definition)
	}

	sort.Strings(markings)

	markingsList, diags := types.ListValueFrom(ctx, types.StringType, markings)
	resp.Diagnostics.Append(diags...)

	state.ID = types.StringValue(group.ID)
	state.Name = types.StringValue(group.Name)
	state.Description = types.StringValue(group.Description)
	state.Roles = rolesList
	state.AllowedMarking = markingsList
	state.MaxConfidenceLevel = types.Int32Value(int32(group.GroupConfidenceLevel.MaxConfidence))
	state.AutoNewMarking = types.BoolValue(group.AutoNewMarking)
	state.DefaultAssignation = types.BoolValue(group.DefaultAssignation)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.ReadGroup(ctx, "id name description roles { edges { node {id name} } } allowed_marking {id definition_type definition} default_assignation auto_new_marking group_confidence_level { max_confidence }", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti group", err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("Group read: %+v", group))

	var rolesPlan []string
	diags = plan.Roles.ElementsAs(ctx, &rolesPlan, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	rolesOfGroup := []string{}

	// Remove roles
	for _, role := range group.Roles.Edges {
		if !slices.Contains(rolesPlan, role.Node.Name) {
			tflog.Info(ctx, fmt.Sprintf("Removing role: %s", role.Node.Name))

			if _, err := group.UnassignRole(ctx, r.client, role.Node.ID); err != nil {
				resp.Diagnostics.AddError(
					"Error Unassigning OpenCTI Role from Group", err.Error(),
				)

				return
			}
		} else {
			rolesOfGroup = append(rolesOfGroup, role.Node.Name)
		}
	}

	// Add roles
	existingRoles, err := r.client.ListRoles(ctx, "id name", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing roles", err.Error(),
		)

		return
	}

	for _, role := range rolesPlan {
		if !slices.Contains(rolesOfGroup, role) {
			tflog.Info(ctx, fmt.Sprintf("Adding role: %s", role))

			for _, remoteRole := range existingRoles {
				if role == remoteRole.Name {
					if _, err := group.AssignRole(ctx, r.client, remoteRole.ID); err != nil {
						resp.Diagnostics.AddError(
							"Error assigning role to group", err.Error(),
						)

						return
					}

					rolesOfGroup = append(rolesOfGroup, remoteRole.Name)

					break
				}
			}
		}
	}

	sort.Strings(rolesOfGroup)

	tflog.Debug(ctx, fmt.Sprintf("Roles: %s", rolesOfGroup))

	rolesList, diags := types.ListValueFrom(ctx, types.StringType, rolesOfGroup)
	resp.Diagnostics.Append(diags...)

	plan.Roles = rolesList

	var markingsPlan []string
	diags = plan.AllowedMarking.ElementsAs(ctx, &markingsPlan, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	markingsOfGroup := []string{}

	// Remove markings
	for _, marking := range group.AllowedMarking {
		if !slices.Contains(markingsPlan, marking.Definition) {
			tflog.Info(ctx, fmt.Sprintf("Removing marking definition: %s", marking.Definition))

			if _, err := group.UnassignMarkingDefinition(ctx, r.client, marking.ID); err != nil {
				resp.Diagnostics.AddError(
					"Error Unassigning OpenCTI Marking Definition from Group", err.Error(),
				)

				return
			}
		} else {
			markingsOfGroup = append(markingsOfGroup, marking.Definition)
		}
	}

	// Add markings
	existingMarkings, err := r.client.ListMarkingDefinitions(ctx, "id definition", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing marking definitions", err.Error(),
		)

		return
	}

	for _, marking := range markingsPlan {
		if !slices.Contains(markingsOfGroup, marking) {
			tflog.Info(ctx, fmt.Sprintf("Adding marking definition: %s", marking))

			for _, remoteMarking := range existingMarkings {
				if marking == remoteMarking.Definition {
					if _, err := group.AssignMarkingDefinition(ctx, r.client, remoteMarking.ID); err != nil {
						resp.Diagnostics.AddError(
							"Error assigning marking definition to group", err.Error(),
						)

						return
					}

					markingsOfGroup = append(markingsOfGroup, remoteMarking.Definition)

					break
				}
			}
		}
	}

	sort.Strings(markingsOfGroup)

	tflog.Debug(ctx, fmt.Sprintf("Marking definitions: %s", markingsOfGroup))

	markingsList, diags := types.ListValueFrom(ctx, types.StringType, markingsOfGroup)
	resp.Diagnostics.Append(diags...)

	plan.AllowedMarking = markingsList

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteGroup(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI Group",
			"Could not delete group, unexpected error: "+err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
