// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/lakeformation/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	// "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	// "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	// "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// @FrameworkResource("aws_lakeformation_lake_formation_opt_in", name="Lake Formation Opt In")
func newResourceLakeFormationOptIn(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceLakeFormationOptIn{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameLakeFormationOptIn = "Lake Formation Opt In"
	resource_name             = "plan.Name.String() - add name for this resource"
	state_id                  = "state.ID.String() - add parameters to uniquely identify this resource"
)

type resourceLakeFormationOptIn struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *resourceLakeFormationOptIn) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lakeformation_lake_formation_opt_in"
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceLakeFormationOptIn) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrPrincipal: schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			names.AttrID:        framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrDatabase: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[LFOptInDatabase](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptional_duplicate(),
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"table": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[LFOptInTable](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptional_duplicate(),
						names.AttrDatabaseName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrName: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrName),
									path.MatchRelative().AtParent().AtName("wildcard"),
								),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"wildcard": schema.BoolAttribute{
							Optional: true,
							Validators: []validator.Bool{
								boolvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrName),
									path.MatchRelative().AtParent().AtName("wildcard"),
								),
							},
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

// TODO:  share function in resource_lf_tag.go? what is the etiquette?
func catalogIDSchemaOptional_duplicate() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func (r *resourceLakeFormationOptIn) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan
	// 3. Populate a create input structure
	// 4. Call the AWS create/put function
	// 5. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work, as well as any computed
	//    only attributes.
	// 6. Use a waiter to wait for create to complete
	// 7. Save the request plan to response state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().LakeFormationClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan ResourceLakeFormationOptInData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input lakeformation.CreateLakeFormationOptInInput
	input.Principal = &awstypes.DataLakePrincipal{
		DataLakePrincipalIdentifier: fwflex.StringFromFramework(ctx, plan.Principal),
	}

	lfoptin := newLFOptIn(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := lfoptin.expandResource(ctx, &resp.Diagnostics)
	input.Resource = res
	if resp.Diagnostics.HasError() {
		return
	}

	var output *lakeformation.CreateLakeFormationOptInOutput
	err := retry.RetryContext(ctx, IAMPropagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateLakeFormationOptIn(ctx, &input)
		if err != nil {
			if errs.IsA[*awstypes.ConcurrentModificationException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameLakeFormationOptIn, prettify(input), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := strconv.Itoa(create.StringHashcode(prettify(input)))
	plan.ID = fwflex.StringValueToFramework(ctx, id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func findLFOptInByID(ctx context.Context, conn *lakeformation.Client, principal *awstypes.DataLakePrincipal, resource *awstypes.Resource) (*lakeformation.ListLakeFormationOptInsOutput, error) {
	in := &lakeformation.ListLakeFormationOptInsInput{
		Principal: principal,
		Resource:  resource,
	}

	out, err := conn.ListLakeFormationOptIns(ctx, in)

	if len(out.LakeFormationOptInsInfoList) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *resourceLakeFormationOptIn) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().LakeFormationClient(ctx)

	// TIP: -- 2. Fetch the state
	var state ResourceLakeFormationOptInData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lfoptin := newLFOptIn(&state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := lfoptin.expandResource(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	principal := &awstypes.DataLakePrincipal{DataLakePrincipalIdentifier: fwflex.StringFromFramework(ctx, state.Principal)}
	out, err := findLFOptInByID(ctx, conn, principal, res)
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameLakeFormationOptIn, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceLakeFormationOptIn) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	// TIP: -- 2. Fetch the state
	var state ResourceLakeFormationOptInData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &lakeformation.DeleteLakeFormationOptInInput{
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(state.Principal.ValueString()),
		},
	}

	lfoptin := newLFOptIn(&state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := lfoptin.expandResource(ctx, &resp.Diagnostics)
	input.Resource = res
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		var err error
		_, err = conn.DeleteLakeFormationOptIn(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(fmt.Errorf("removing Lake Formation Opt In: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteLakeFormationOptIn(ctx, input)
	}

	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionDeleting, ResNameLakeFormationOptIn, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceLakeFormationOptIn) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrDatabase),
			path.MatchRoot("table"),
		),
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceLakeFormationOptIn) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.

/*
TODO: all waiters are commented out
func waitLakeFormationOptInCreated(ctx context.Context, conn *lakeformation.Client, id string, timeout time.Duration) (*awstypes.LakeFormationOptIn, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusLakeFormationOptIn(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lakeformation.LakeFormationOptIn); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitLakeFormationOptInUpdated(ctx context.Context, conn *lakeformation.Client, id string, timeout time.Duration) (*awstypes.LakeFormationOptIn, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusLakeFormationOptIn(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lakeformation.LakeFormationOptIn); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitLakeFormationOptInDeleted(ctx context.Context, conn *lakeformation.Client, id string, timeout time.Duration) (*awstypes.LakeFormationOptIn, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusLakeFormationOptIn(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lakeformation.LakeFormationOptIn); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusLakeFormationOptIn(ctx context.Context, conn *lakeformation.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findLakeFormationOptIn(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

*/

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.

/*
	TODO think about cases like:

- there is opt-in for a database and this function executes for db and table
- opposite of above
- other combinations like opt-in for a role (is this allowed?)
TODO - the below shouldn't be a public function. test.go notes to add to exports.go
*/
func FindLakeFormationOptIn(ctx context.Context, conn *lakeformation.Client, principal string) (*lakeformation.ListLakeFormationOptInsOutput, error) {
	in := &lakeformation.ListLakeFormationOptInsInput{
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: &principal,
		},
	}

	out, err := conn.ListLakeFormationOptIns(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func (d *dbOptIn) expandResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	dbptr, err := d.data.Database.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var db awstypes.DatabaseResource
	diags.Append(fwflex.Expand(ctx, dbptr, &db)...)
	if diags.HasError() {
		return nil
	}

	r.Database = &db

	return &r
}

func (d *tbOptIn) expandResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource

	tbptr, err := d.data.Table.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var tb awstypes.TableResource
	diags.Append(fwflex.Expand(ctx, tbptr, &tb)...)
	if diags.HasError() {
		return nil
	}

	r.Table = &tb

	return &r
}

type lfOptIn interface {
	expandResource(context.Context, *diag.Diagnostics) *awstypes.Resource
	// findTag(context.Context, *lakeformation.GetResourceLFTagsOutput, *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[LFTag]
}

type dbOptIn struct {
	data *ResourceLakeFormationOptInData
}

type tbOptIn struct {
	data *ResourceLakeFormationOptInData
}

func newLFOptIn(r *ResourceLakeFormationOptInData, diags *diag.Diagnostics) lfOptIn {
	switch {
	case !r.Database.IsNull():
		return &dbOptIn{data: r}
	case !r.Table.IsNull():
		return &tbOptIn{data: r}
	default:
		diags.AddError("unexpected resource type",
			"unexpected resource type")
		return nil
	}
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values

// TODO: reuse ones in resource_lf_tag.go?
type ResourceLakeFormationOptInData struct {
	Database  fwtypes.ListNestedObjectValueOf[LFOptInDatabase] `tfsdk:"database"`
	ID        types.String                                     `tfsdk:"id"`
	Principal types.String                                     `tfsdk:"principal"`
	Table     fwtypes.ListNestedObjectValueOf[LFOptInTable]    `tfsdk:"table"`
	Timeouts  timeouts.Value                                   `tfsdk:"timeouts"`
}

type LFOptInDatabase struct {
	CatalogID types.String `tfsdk:"catalog_id"`
	Name      types.String `tfsdk:"name"`
}

type LFOptInTable struct {
	CatalogID    types.String `tfsdk:"catalog_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	Name         types.String `tfsdk:"name"`
	Wildcard     types.Bool   `tfsdk:"wildcard"`
}
