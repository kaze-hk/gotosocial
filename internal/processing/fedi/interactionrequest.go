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

package fedi

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

func (p *Processor) LikeRequestGet(
	ctx context.Context,
	requestedUser string,
	id string,
) (any, gtserror.WithCode) {
	intReq, errWithCode := p.interactionRequestGet(ctx, requestedUser, id)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if intReq.InteractionType != gtsmodel.InteractionLike {
		const text = "interaction request was not LikeRequest"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	return p.intReqData(ctx, intReq)
}

func (p *Processor) ReplyRequestGet(
	ctx context.Context,
	requestedUser string,
	id string,
) (any, gtserror.WithCode) {
	intReq, errWithCode := p.interactionRequestGet(ctx, requestedUser, id)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if intReq.InteractionType != gtsmodel.InteractionReply {
		const text = "interaction request was not ReplyRequest"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	return p.intReqData(ctx, intReq)
}

func (p *Processor) AnnounceRequestGet(
	ctx context.Context,
	requestedUser string,
	id string,
) (any, gtserror.WithCode) {
	intReq, errWithCode := p.interactionRequestGet(ctx, requestedUser, id)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if intReq.InteractionType != gtsmodel.InteractionAnnounce {
		const text = "interaction request was not AnnounceRequest"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	return p.intReqData(ctx, intReq)
}

func (p *Processor) interactionRequestGet(
	ctx context.Context,
	requestedUser string,
	id string,
) (*gtsmodel.InteractionRequest, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if auth.handshakingURI != nil {
		// We're currently handshaking, which means
		// we don't know this account yet. This should
		// be a very rare race condition.
		err := gtserror.Newf("network race handshaking %s", auth.handshakingURI)
		return nil, gtserror.NewErrorInternalError(err)
	}
	receiver := auth.receiver
	requester := auth.requester

	intReq, err := p.state.DB.GetInteractionRequestByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if intReq == nil {
		err := gtserror.Newf("interaction request %s not found in the db", id)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Interaction request must be owned
	// by receiving account / requestedUser.
	if intReq.InteractingAccountID != receiver.ID {
		const text = "interaction request does not belong to receiving account"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// Requester must be either the owner of
	// the interaction request or the target.
	if requester.ID != intReq.TargetAccountID &&
		requester.ID != intReq.InteractingAccountID {
		const text = "interaction request not visible to requesting account"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// Only polite interaction requests can
	// be converted to InteractionRequestable.
	if !intReq.IsPolite() {
		const text = "interaction request not polite"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	return intReq, nil
}

func (p *Processor) intReqData(ctx context.Context, intReq *gtsmodel.InteractionRequest) (any, gtserror.WithCode) {
	intRequestable, err := p.converter.InteractionReqToASInteractionRequestable(ctx, intReq)
	if err != nil {
		err := gtserror.Newf("error converting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(intRequestable)
	if err != nil {
		err := gtserror.Newf("error serializing interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
