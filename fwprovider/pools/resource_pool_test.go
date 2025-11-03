//go:build acceptance || all

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package pools_test

import (
	"context"
	fwpools "github.com/bpg/terraform-provider-proxmox/fwprovider/pools"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/test"
	"github.com/bpg/terraform-provider-proxmox/proxmox/pools"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAccPoolResource(t *testing.T) {
	model := fwpools.PoolModel{}
	te := test.InitEnvironment(t)
	accTestPoolName := gofakeit.Word()
	accTestPoolName2 := gofakeit.Word()
	te.AddTemplateVars(map[string]interface{}{
		"TestPoolName":  accTestPoolName,
		"TestPoolName2": accTestPoolName2,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: te.AccProviders,
		Steps: []resource.TestStep{
			{
				Config: te.RenderConfig(`
				resource "proxmox_virtual_environment_pool" "test_pool" {
					id = "{{ .TestPoolName }}"
					comment = "Hello world"
				}`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(
							"proxmox_virtual_environment_pool.test_pool",
							plancheck.ResourceActionCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccGetPool(t, te, accTestPoolName, &model),
					func(state *terraform.State) error {
						assert.Equal(t, model.ID.ValueString(), accTestPoolName)
						assert.Equal(t, model.Comment.ValueString(), "Hello world")
						return nil
					},
				),
			},
			{
				ResourceName:      "proxmox_virtual_environment_pool.test_pool",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: te.RenderConfig(`
				resource "proxmox_virtual_environment_pool" "test_pool" {
					id = "{{ .TestPoolName }}"
					comment = "Hello world"
				}`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: te.RenderConfig(`
				resource "proxmox_virtual_environment_pool" "test_pool" {
					id = "{{ .TestPoolName }}"
					comment = "Hello world2"
				}`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(
							"proxmox_virtual_environment_pool.test_pool",
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccGetPool(t, te, accTestPoolName, &model),
					func(state *terraform.State) error {
						assert.Equal(t, model.ID.ValueString(), accTestPoolName)
						assert.Equal(t, model.Comment.ValueString(), "Hello world2")
						return nil
					},
				),
			},
			{
				Config: te.RenderConfig(`
				resource "proxmox_virtual_environment_pool" "test_pool" {
					id = "{{ .TestPoolName2 }}"
					comment = "Hello world2"
				}`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(
							"proxmox_virtual_environment_pool.test_pool",
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolNotExists(t, te, accTestPoolName),
					testAccGetPool(t, te, accTestPoolName2, &model),
					func(state *terraform.State) error {
						assert.Equal(t, model.ID.ValueString(), accTestPoolName2)
						assert.Equal(t, model.Comment.ValueString(), "Hello world2")
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckPoolNotExists(t *testing.T, te *test.Environment, id string) resource.TestCheckFunc {
	t.Helper()
	return func(state *terraform.State) error {
		client := pools.Client{Client: te.Client()}
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		_, err := client.GetPool(ctx, id)
		require.Error(t, err)
		return nil
	}
}

func testAccGetPool(t *testing.T, te *test.Environment, id string, model *fwpools.PoolModel) resource.TestCheckFunc {
	t.Helper()
	return func(state *terraform.State) error {
		client := pools.Client{Client: te.Client()}
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		pool, err := client.GetPool(ctx, id)
		require.NoError(t, err)

		model.ID = types.StringValue(id)
		model.Comment = types.StringPointerValue(pool.Comment)
		return nil
	}
}
