// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gtsmodel_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/assert"
)

var mediaErrorDetailsTests = []struct {
	p  gtsmodel.MediaErrorDetails
	u1 gtsmodel.MediaErrorType
	u2 uint16
}{
	{
		p:  0,
		u1: 0,
		u2: 0,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeNone, 0),
		u1: gtsmodel.MediaErrorTypeNone,
		u2: 0,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeInterrupt, 0),
		u1: gtsmodel.MediaErrorTypeInterrupt,
		u2: 0,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypePolicy, gtsmodel.MediaErrorTypePolicy_Size),
		u1: gtsmodel.MediaErrorTypePolicy,
		u2: gtsmodel.MediaErrorTypePolicy_Size,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypePolicy, gtsmodel.MediaErrorTypePolicy_Domain),
		u1: gtsmodel.MediaErrorTypePolicy,
		u2: gtsmodel.MediaErrorTypePolicy_Domain,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeNetwork, gtsmodel.MediaErrorTypeNetwork_DNS),
		u1: gtsmodel.MediaErrorTypeNetwork,
		u2: gtsmodel.MediaErrorTypeNetwork_DNS,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeNetwork, gtsmodel.MediaErrorTypeNetwork_Timeout),
		u1: gtsmodel.MediaErrorTypeNetwork,
		u2: gtsmodel.MediaErrorTypeNetwork_Timeout,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeHTTP, 400),
		u1: gtsmodel.MediaErrorTypeHTTP,
		u2: 400,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeHTTP, 404),
		u1: gtsmodel.MediaErrorTypeHTTP,
		u2: 404,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeHTTP, 500),
		u1: gtsmodel.MediaErrorTypeHTTP,
		u2: 500,
	},
	{
		p:  gtsmodel.NewMediaErrorDetails(gtsmodel.MediaErrorTypeCodec, gtsmodel.MediaErrorTypeCodec_Unsupported),
		u1: gtsmodel.MediaErrorTypeCodec,
		u2: gtsmodel.MediaErrorTypeCodec_Unsupported,
	},
}

func TestMediaErrorDetailsPack(t *testing.T) {
	for _, test := range mediaErrorDetailsTests {
		d := gtsmodel.NewMediaErrorDetails(test.u1, test.u2)
		u1, u2 := unpacku16s(uint32(d))
		assert.Equal(t, u1, uint16(test.u1))
		assert.Equal(t, u2, uint16(test.u2))
	}
}

func TestMediaErrorDetailsUnpack(t *testing.T) {
	for _, test := range mediaErrorDetailsTests {
		assert.Equal(t, test.u1, test.p.Type())
		assert.Equal(t, test.u2, test.p.Details())
	}
}

func unpacku16s(u uint32) (u1, u2 uint16) {
	const bits = 16
	const mask = (1 << bits) - 1
	u1 = uint16(u >> bits)
	u2 = uint16(u & mask)
	return
}
