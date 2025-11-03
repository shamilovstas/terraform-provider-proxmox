/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package pools

import (
	"context"
	"errors"
	"fmt"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/pools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*poolResource)(nil)
	_ resource.ResourceWithConfigure   = (*poolResource)(nil)
	_ resource.ResourceWithImportState = (*poolResource)(nil)
)

func ErrPoolExists(poolID string) api.Error {
	return api.Error(fmt.Sprintf("create pool failed: pool '%s' already exists\\n", poolID))
}

func ErrPoolDoesNotExist(poolID string) api.Error {
	return api.Error(fmt.Sprintf("pool '%s' does not exist\\n", poolID))
}

func ErrDeletePoolDoesNotExist(poolID string) api.Error {
	return api.Error(fmt.Sprintf("delete pool failed: pool '%s' does not exist\\n", poolID))
}

type poolResource struct {
	client proxmox.Client
}

func NewPoolResource() resource.Resource {
	return &poolResource{}
}

func (r *poolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *poolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool2"
}

func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The pool ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "The pool comment",
			},
		},
	}
}

func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PoolModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	poolID := plan.ID.ValueString()
	comment := plan.Comment.ValueString()

	body := &pools.PoolCreateRequestBody{
		ID:      poolID,
		Comment: &comment,
	}

	if err := r.client.Pool().CreatePool(ctx, body); err != nil {
		var detailedErrMessage string
		if errors.Is(err, ErrPoolExists(poolID)) {
			detailedErrMessage = fmt.Sprintf("Pool with ID '%s' already exists", poolID)
		} else {
			detailedErrMessage = fmt.Sprintf("Could not create pool, error: %s", err.Error())
		}

		resp.Diagnostics.AddError("Pool creation failed", detailedErrMessage)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	poolID := state.ID.ValueString()

	pool, err := r.client.Pool().GetPool(ctx, poolID)
	if err != nil {
		if errors.Is(err, ErrPoolDoesNotExist(poolID)) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to get pool", fmt.Sprintf("Could not get pool with id '%s'", poolID))
		}

		return
	}

	state.Comment = types.StringPointerValue(pool.Comment)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *poolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state PoolModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	poolID := state.ID.ValueString()
	comment := state.Comment.ValueString()

	body := pools.PoolUpdateRequestBody{
		Comment: &comment,
	}

	if err := r.client.Pool().UpdatePool(ctx, poolID, &body); err != nil {
		var detailedErrMessage string
		if errors.Is(err, ErrPoolDoesNotExist(poolID)) {
			detailedErrMessage = fmt.Sprintf("Pool with ID '%s' already exists", poolID)
		} else {
			detailedErrMessage = fmt.Sprintf("Could not update pool with ID '%s'\n\nError: %s", poolID, err.Error())
		}

		resp.Diagnostics.AddError("Pool update failed", detailedErrMessage)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *poolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	poolID := state.ID.ValueString()
	if err := r.client.Pool().DeletePool(ctx, poolID); err != nil {
		var detailedErrMessage string
		if errors.Is(err, ErrDeletePoolDoesNotExist(poolID)) {
			detailedErrMessage = fmt.Sprintf("Pool with ID '%s' does not exist", poolID)
		} else {
			detailedErrMessage = fmt.Sprintf("Could not delete pool with ID '%s'.\n\nError: %s", poolID, err.Error())
		}

		resp.Diagnostics.AddError("Pool delete failed", detailedErrMessage)

		return
	}
}

func (r *poolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(config.Resource)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected config.Resource, got: %T", req.ProviderData),
		)

		return
	}

	r.client = cfg.Client
}
