/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package pools

import "github.com/hashicorp/terraform-plugin-framework/types"

type PoolModel struct {
	ID      types.String `tfsdk:"id"`
	Comment types.String `tfsdk:"comment"`
}
