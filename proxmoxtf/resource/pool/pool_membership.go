/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package pool

import (
	"context"
	"fmt"
	"github.com/bpg/terraform-provider-proxmox/proxmox/pools"
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
	"github.com/bpg/terraform-provider-proxmox/proxmoxtf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"strings"
)

const (
	mkResourceVirtualEnvironmentPoolMembershipVmID   = "vm_id"
	mkResourceVirtualEnvironmentPoolMembershipPoolID = "pool_id"
)

// Membership returns a resource that manages pool memberships.
func Membership() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			mkResourceVirtualEnvironmentPoolMembersVMID: {
				Type:        schema.TypeInt,
				Description: "VM or CT ID",
				Required:    true,
				ForceNew:    true,
			},
			mkResourceVirtualEnvironmentPoolMembershipPoolID: {
				Type:        schema.TypeString,
				Description: "Pool ID",
				Required:    true,
				ForceNew:    true,
			},
		},
		CreateContext: poolMembershipCreate,
		ReadContext:   poolMembershipRead,
		DeleteContext: poolMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: poolMembershipImport,
		},
	}
}

func poolMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()
	if err != nil {
		return diag.FromErr(err)
	}

	vmId := d.Get(mkResourceVirtualEnvironmentPoolMembershipVmID).(int)
	poolId := d.Get(mkResourceVirtualEnvironmentPoolMembershipPoolID).(string)

	poolApi := client.Pool()

	vmList := (types.CustomCommaSeparatedList)([]string{strconv.Itoa(vmId)})

	trueValue := types.CustomBool(true)
	body := &pools.PoolUpdateRequestBody{
		VMs:       &vmList,
		AllowMove: &trueValue,
	}

	err = poolApi.UpdatePool(ctx, poolId, body)
	if err != nil {
		return diag.FromErr(err)
	}

	id := fmt.Sprintf("%s/%d", poolId, vmId)
	d.SetId(id)
	return poolMembershipRead(ctx, d, m)
}

func poolMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()

	if err != nil {
		return diag.FromErr(err)
	}

	poolId := d.Get(mkResourceVirtualEnvironmentPoolMembershipPoolID).(string)
	vmId := d.Get(mkResourceVirtualEnvironmentPoolMembershipVmID).(int)
	pool, err := client.Pool().GetPool(ctx, poolId)

	if err != nil {
		return diag.FromErr(err)
	}

	for _, member := range pool.Members {
		if member.VMID != nil && *member.VMID == vmId {
			return nil
		}
	}
	d.SetId("")
	return nil
}

func poolMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()

	if err != nil {
		return diag.FromErr(err)
	}

	poolId := d.Get(mkResourceVirtualEnvironmentPoolMembershipPoolID).(string)
	vmId := d.Get(mkResourceVirtualEnvironmentPoolMembershipVmID).(int)

	vmList := (types.CustomCommaSeparatedList)([]string{strconv.Itoa(vmId)})

	trueValue := types.CustomBool(true)
	body := &pools.PoolUpdateRequestBody{
		VMs:    &vmList,
		Delete: &trueValue,
	}

	err = client.Pool().UpdatePool(ctx, poolId, body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func poolMembershipImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format, expected: pool-id/vm-id, got: %s", d.Id())
	}

	poolID := parts[0]
	vmIDStr := parts[1]

	vmID, err := strconv.Atoi(vmIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid VM ID %q: must be an integer", vmIDStr)
	}

	if err = d.Set(mkResourceVirtualEnvironmentPoolMembershipPoolID, poolID); err != nil {
		return nil, fmt.Errorf("failed to set pool_id: %w", err)
	}

	if err = d.Set(mkResourceVirtualEnvironmentPoolMembershipVmID, vmID); err != nil {
		return nil, fmt.Errorf("failed to set vm_id: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}
