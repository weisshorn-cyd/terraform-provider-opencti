package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weisshorn-cyd/gocti"
	"github.com/weisshorn-cyd/gocti/system"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &statusTemplateResource{}
	_ resource.ResourceWithConfigure   = &statusTemplateResource{}
	_ resource.ResourceWithImportState = &statusTemplateResource{}
)

// NewStatusTemplateResource is a helper function to simplify the provider implementation.
func NewStatusTemplateResource() resource.Resource {
	return &statusTemplateResource{}
}

// statusTemplateResource is the resource implementation.
type statusTemplateResource struct {
	client *gocti.OpenCTIAPIClient
}

// statusTemplateResourceModel maps the resource schema data.
type statusTemplateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Workflows   types.List   `tfsdk:"workflows"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

type workflowModel struct {
	Entity types.String `tfsdk:"entity"`
	Order  types.Int64  `tfsdk:"order"`
}

// Metadata returns the resource type name.
func (r *statusTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_template"
}

// Schema defines the schema for the resource.
func (r *statusTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"color": schema.StringAttribute{
				Required: true,
			},
			"workflows": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"entity": schema.StringAttribute{
							Required: true,
						},
						"order": schema.Int64Attribute{
							Required: true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *statusTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan statusTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract the workflows (if provided)
	var workflows []workflowModel
	// Check if `workflows` is not null and is known
	if !plan.Workflows.IsNull() && !plan.Workflows.IsUnknown() {
		diags := plan.Workflows.ElementsAs(ctx, &workflows, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	tflog.Info(ctx, "Creating status templates")

	// Create new status template
	createdStatus, err := r.client.CreateStatusTemplate(ctx, "id name color", system.StatusTemplateAddInput{
		Name:  plan.Name.ValueString(),
		Color: plan.Color.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating status template",
			"Could not create status template, unexpected error: "+err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Status template created: %v", createdStatus))

	plan.ID = types.StringValue(createdStatus.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Create the workflows if any
	subType := system.SubType{}

	tflog.Info(ctx, fmt.Sprintf("Creating workflows %s", workflows))

	workflowsValues := []attr.Value{}

	for _, workflow := range workflows {
		tflog.Info(ctx, fmt.Sprintf("Assigning status %s to entity %s with order %d", plan.Name.ValueString(), workflow.Entity.ValueString(), workflow.Order.ValueInt64()))

		_, err = subType.SetStatusInWorkFlow(ctx, r.client, workflow.Entity.ValueString(), plan.ID.ValueString(), int(workflow.Order.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting status in workflow",
				"Could not create status in workflow, unexpected error: "+err.Error(),
			)

			return
		}

		tflog.Debug(ctx, "Status assigned in workflow")

		// Convert the workflowModel for the plan
		workflowValue, diags := types.ObjectValue(
			map[string]attr.Type{
				"entity": types.StringType,
				"order":  types.Int64Type,
			},
			map[string]attr.Value{
				"entity": types.StringValue(workflow.Entity.ValueString()),
				"order":  types.Int64Value(workflow.Order.ValueInt64()),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		workflowsValues = append(workflowsValues, workflowValue)
	}

	workflowsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"entity": types.StringType,
				"order":  types.Int64Type,
			},
		},
		workflowsValues,
	)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.Workflows = workflowsList

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *statusTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state statusTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read status template from opencti
	statusTemplate, err := r.client.ReadStatusTemplate(ctx, "id name color usages", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading opencti status template", err.Error(),
		)

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Status template read: %+v", statusTemplate))

	state.ID = types.StringValue(statusTemplate.ID)
	state.Name = types.StringValue(statusTemplate.Name)
	state.Color = types.StringValue(statusTemplate.Color)

	// It is not simple to retrieve the workflow of all the entities, compare only the number of usages.
	if len(state.Workflows.Elements()) != statusTemplate.Usages {
		tflog.Debug(ctx, "Number of workflows is different from the number of usages")

		workflowsListEmpty, diags := types.ListValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"entity": types.StringType,
					"order":  types.Int64Type,
				},
			},
			[]attr.Value{},
		)
		resp.Diagnostics.Append(diags...)

		state.Workflows = workflowsListEmpty
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *statusTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *statusTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state statusTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DeleteStatusTemplate(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenCTI Status Template",
			"Could not delete status template, unexpected error: "+err.Error(),
		)

		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *statusTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
