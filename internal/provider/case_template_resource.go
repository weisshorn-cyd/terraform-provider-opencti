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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weisshorn-cyd/gocti"
	"github.com/weisshorn-cyd/gocti/system"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &caseTemplateResource{}
	_ resource.ResourceWithConfigure   = &caseTemplateResource{}
	_ resource.ResourceWithImportState = &caseTemplateResource{}
)

// NewCaseTemplateResource is a helper function to simplify the provider implementation.
func NewCaseTemplateResource() resource.Resource {
	return &caseTemplateResource{}
}

// caseTemplateResource is the resource implementation.
type caseTemplateResource struct {
	client *gocti.OpenCTIAPIClient
}

// caseTemplateResourceModel maps the resource schema data.
type caseTemplateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tasks       types.Set    `tfsdk:"tasks"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *caseTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_template"
}

// Schema defines the schema for the resource.
func (r *caseTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tasks": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *caseTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan caseTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating case templates")

	// Convert the tasks ListValue to a []string
	var taskList []string

	diags = plan.Tasks.ElementsAs(ctx, &taskList, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create new case template
	createdCase, err := r.client.CreateCaseTemplate(ctx, "id name description tasks { edges { node { id name } } }", system.CaseTemplateAddInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tasks:       taskList,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating case template",
			"Could not create case template, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Case template created: %+v", createdCase))

	tasks := []string{}
	for _, task := range createdCase.Tasks.Edges {
		tasks = append(tasks, task.Node.ID)
	}

	tasksList, diags := types.SetValueFrom(ctx, types.StringType, tasks)
	resp.Diagnostics.Append(diags...)

	plan.ID = types.StringValue(createdCase.ID)
	plan.Name = types.StringValue(createdCase.Name)
	plan.Description = types.StringValue(createdCase.Description)
	plan.Tasks = tasksList
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *caseTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state caseTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read case template from opencti
	caseTemplate, err := r.client.ReadCaseTemplate(ctx, "id name description tasks { edges { node { id name } } }", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti case template", err.Error(),
		)

		return
	}

	tflog.Info(ctx, fmt.Sprintf("Case template read: %+v", caseTemplate))

	tasks := []string{}
	for _, task := range caseTemplate.Tasks.Edges {
		tasks = append(tasks, task.Node.ID)
	}

	tasksList, diags := types.SetValueFrom(ctx, types.StringType, tasks)
	resp.Diagnostics.Append(diags...)

	state.ID = types.StringValue(caseTemplate.ID)
	state.Name = types.StringValue(caseTemplate.Name)
	state.Description = types.StringValue(caseTemplate.Description)
	state.Tasks = tasksList

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *caseTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *caseTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state caseTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteCaseTemplate(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI Case Template",
			"Could not delete case template, unexpected error: "+err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *caseTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *caseTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
