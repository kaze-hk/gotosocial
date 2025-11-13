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

package bundb_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type DomainLimitTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *DomainLimitTestSuite) TestCreateMatchDomainLimit() {
	var (
		ctx   = suite.T().Context()
		limit = &gtsmodel.DomainLimit{
			ID:                 "01JCZN614XG85GCGAMSV9ZZAEJ",
			Domain:             "exämple.org",
			CreatedByAccountID: suite.testAccounts["admin_account"].ID,
			PrivateComment:     "this domain is poo",
			PublicComment:      "this domain is poo, but phrased in a more outward-facing way",
		}
	)

	// Whack the limit in.
	if err := suite.state.DB.PutDomainLimit(ctx, limit); err != nil {
		suite.FailNow(err.Error())
	}

	// Check that we get back the limit we just
	// created when trying to match on a subdomain.
	dbLimit, err := suite.state.DB.MatchDomainLimit(ctx, "test.exämple.org")
	if err != nil {
		suite.FailNow(err.Error())
	}

	if dbLimit == nil {
		suite.FailNow("domain was not limited")
	}

	// Domain is stored punified.
	suite.Equal("xn--exmple-cua.org", limit.Domain)

	// Everything else should be @ default values.
	suite.Equal(gtsmodel.MediaPolicyNoAction, limit.MediaPolicy)
	suite.Equal(gtsmodel.FollowsPolicyNoAction, limit.FollowsPolicy)
	suite.Equal(gtsmodel.StatusesPolicyNoAction, limit.StatusesPolicy)
	suite.Equal(gtsmodel.AccountsPolicyNoAction, limit.AccountsPolicy)
	suite.Equal("", limit.ContentWarning)
}

func TestDomainLimitTestSuite(t *testing.T) {
	suite.Run(t, new(DomainLimitTestSuite))
}
