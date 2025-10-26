/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package pool

import (
	"testing"

	"github.com/bpg/terraform-provider-proxmox/proxmoxtf/test"
)

// TestMembershipInstantiation tests whether the Membership instance can be instantiated.
func TestMembershipInstantiation(t *testing.T) {
	t.Parallel()

	s := Membership()
	if s == nil {
		t.Fatalf("Cannot instantiate Pool")
	}
}

// TestMembershipSchema tests the Membership schema.
func TestMembershipSchema(t *testing.T) {
	t.Parallel()

	s := Membership().Schema

	test.AssertRequiredArguments(t, s, []string{
		mkResourceVirtualEnvironmentPoolMembershipPoolID,
	})

	test.AssertRequiredArguments(t, s, []string{
		mkResourceVirtualEnvironmentPoolMembershipVmID,
	})
}
