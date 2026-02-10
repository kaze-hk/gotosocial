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

package debug

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/gin-gonic/gin"
)

const (
	BasePath             = "/v1/debug"
	APUrlPath            = BasePath + "/apurl"
	ClearCachesPath      = BasePath + "/caches/clear"
	StatusVisibilityPath = BasePath + "/status/visibility"

	// endpoint clones to maintain
	// backwards compatibility with
	// previous gotosocial versions
	_CompatAPUrlPath       = "/v1/admin/debug/apurl"
	_CompatClearCachesPath = "/v1/admin/debug/caches/clear"
)

type Module struct {
	state     *state.State
	processor *processing.Processor
}

func New(state *state.State, processor *processing.Processor) *Module {
	return &Module{
		state:     state,
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	// activitypub debug endpoints.
	attachHandler(http.MethodGet, APUrlPath, m.APUrlGETHandler)

	// cache debug endpoints.
	attachHandler(http.MethodPost, ClearCachesPath, m.ClearCachesPOSTHandler)

	// status debug endpoints.
	attachHandler(http.MethodGet, StatusVisibilityPath, m.StatusVisibilityGETHandler)

	// backwards compatibility endpoints
	attachHandler(http.MethodGet, _CompatAPUrlPath, m.APUrlGETHandler)
	attachHandler(http.MethodPost, _CompatClearCachesPath, m.ClearCachesPOSTHandler)
}
