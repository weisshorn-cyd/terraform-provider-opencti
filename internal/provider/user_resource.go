package provider

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
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
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource is the resource implementation.
type userResource struct {
	client *gocti.OpenCTIAPIClient
}

// userResourceModel maps the resource schema data.
type userResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	UserEmail   types.String `tfsdk:"user_email"`
	APIToken    types.String `tfsdk:"api_token"`
	Groups      types.List   `tfsdk:"groups"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"user_email": schema.StringAttribute{
				Required: true,
			},
			"api_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Sensitive: true,
			},
			"groups": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating user")

	usersRaw, err := r.client.ListUsers(ctx, "", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti users",
			"Could not read opencti users, unexpected error: "+err.Error(),
		)

		return
	}

	createdUser := system.User{}

	// Check if the user exist, if so, retrieve it
	for _, user := range usersRaw {
		if plan.Name.ValueString() == user.Name {
			tflog.Info(ctx, fmt.Sprintf("User already exist: %+v", user))
			createdUser = user

			break
		}
	}

	// Create new user
	if createdUser.Name == "" {
		tflog.Info(ctx, "User does not exist, creating")

		createdUser, err = r.client.CreateUser(ctx, "id name user_email api_token", system.UserAddInput{
			UserEmail: plan.UserEmail.ValueString(),
			Name:      plan.Name.ValueString(),
			Password:  uuid.New().String(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating user",
				"Could not create user, unexpected error: "+err.Error(),
			)

			return
		}
	}

	existingGroups, err := r.client.ListGroups(ctx, "id name", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing groups",
			"Could not create user, unexpected error: "+err.Error(),
		)

		return
	}

	groupsAssigned := []string{}

	// Assign the groups
	for _, group := range plan.Groups.Elements() {
		tflog.Info(ctx, fmt.Sprintf("Assigning group %s to user %s", group, createdUser.Name))

		for _, remoteGroup := range existingGroups {
			if strings.Trim(group.String(), "\"") == remoteGroup.Name {
				if _, err := createdUser.AssignGroup(ctx, r.client, remoteGroup.ID); err != nil {
					resp.Diagnostics.AddError(
						"Error assigning group to user", err.Error(),
					)

					return
				}

				groupsAssigned = append(groupsAssigned, remoteGroup.Name)

				break
			}
		}
	}

	sort.Strings(groupsAssigned)

	tflog.Debug(ctx, fmt.Sprintf("User created: %+v", createdUser))

	groupsAssignedList, diags := types.ListValueFrom(ctx, types.StringType, groupsAssigned)
	resp.Diagnostics.Append(diags...)

	plan = userResourceModel{
		ID:        types.StringValue(createdUser.ID),
		Name:      types.StringValue(createdUser.Name),
		UserEmail: types.StringValue(createdUser.UserEmail),
		APIToken:  types.StringValue(createdUser.ApiToken),
		Groups:    groupsAssignedList,
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
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read user from opencti
	user, err := r.client.ReadUser(ctx, "id name user_email api_token groups { edges { node {id name} } }", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti user", err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("User read: %+v", user))

	groups := []string{}
	for _, group := range user.Groups.Edges {
		groups = append(groups, group.Node.Name)
	}

	sort.Strings(groups)
	groupsList, diags := types.ListValueFrom(ctx, types.StringType, groups)
	resp.Diagnostics.Append(diags...)

	state.ID = types.StringValue(user.ID)
	state.Name = types.StringValue(user.Name)
	state.UserEmail = types.StringValue(user.UserEmail)
	state.APIToken = types.StringValue(user.ApiToken)
	state.Groups = groupsList

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.ReadUser(ctx, "id name user_email api_token groups { edges { node {id name} } }", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti user", err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("User read: %+v", user))

	var groupsPlan []string
	diags = plan.Groups.ElementsAs(ctx, &groupsPlan, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	groupsOfUser := []string{}

	// Remove groups
	for _, group := range user.Groups.Edges {
		if !slices.Contains(groupsPlan, group.Node.Name) {
			tflog.Info(ctx, fmt.Sprintf("Removing group: %s", group.Node.Name))

			if _, err := user.UnassignGroup(ctx, r.client, group.Node.ID); err != nil {
				resp.Diagnostics.AddError(
					"Error Unassigning OpenCTI Group from User", err.Error(),
				)

				return
			}
		} else {
			groupsOfUser = append(groupsOfUser, group.Node.Name)
		}
	}

	// Add groups
	existingGroups, err := r.client.ListGroups(ctx, "id name", true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing groups", err.Error(),
		)

		return
	}

	for _, group := range groupsPlan {
		if !slices.Contains(groupsOfUser, group) {
			tflog.Info(ctx, fmt.Sprintf("Adding group: %s", group))

			for _, remoteGroup := range existingGroups {
				if group == remoteGroup.Name {
					if _, err := user.AssignGroup(ctx, r.client, remoteGroup.ID); err != nil {
						resp.Diagnostics.AddError(
							"Error assigning group to user", err.Error(),
						)

						return
					}

					groupsOfUser = append(groupsOfUser, remoteGroup.Name)

					break
				}
			}
		}
	}

	sort.Strings(groupsOfUser)

	tflog.Debug(ctx, fmt.Sprintf("Groups: %s", groupsOfUser))

	groupsList, diags := types.ListValueFrom(ctx, types.StringType, groupsOfUser)
	resp.Diagnostics.Append(diags...)

	plan.Groups = groupsList

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI User", err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
