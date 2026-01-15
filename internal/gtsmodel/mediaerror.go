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

package gtsmodel

import (
	"net/http"
	"strconv"
)

// MediaErrorDetails stores basic error details about
// why a piece of media may not have been downloaded.
// It contains a 16bit MediaErrorType, and the remaining
// 16bits may contain optional extra error details.
type MediaErrorDetails uint32

// MediaErrorType describes a broad error type for why
// one or more media files may not have been downloaded.
type MediaErrorType uint16

const (
	// MediaErrorTypeNone: no error returned.
	MediaErrorTypeNone MediaErrorType = 0

	// MediaErrorTypePolicy: file(s) not downloaded due to configured policy.
	MediaErrorTypePolicy        MediaErrorType = 1
	MediaErrorTypePolicy_Size   uint16         = 1 // nolint:revive
	MediaErrorTypePolicy_Domain uint16         = 2 // nolint:revive

	// MediaErrorTypeInterrupt: file(s) not downloaded due to interrupt (i.e. context errors).
	MediaErrorTypeInterrupt MediaErrorType = 2

	// MediaErrorTypeHTTP: file(s) not downloaded due to HTTP response error.
	// (the remaining 16bits of MediaErrorDetails store status code response)
	MediaErrorTypeHTTP MediaErrorType = 3

	// MediaErrorTypeNetwork: file(s) not downloaded due to network issue.
	MediaErrorTypeNetwork         MediaErrorType = 4
	MediaErrorTypeNetwork_Timeout uint16         = 1 // nolint:revive
	MediaErrorTypeNetwork_DNS     uint16         = 2 // nolint:revive

	// MediaErrorTypeCodec: file(s) not downloaded due to a codec issue.
	MediaErrorTypeCodec             MediaErrorType = 5
	MediaErrorTypeCodec_Unsupported uint16         = 1 // nolint:revive

	// MediaErrorTypeUnknown: file(s) not downloaded due to unclassified error.
	MediaErrorTypeUnknown MediaErrorType = 6
)

// NewMediaErrorDetails returns a new MediaErrorDetails encapsulating MediaErrorType and details (if any).
func NewMediaErrorDetails(errType MediaErrorType, details uint16) MediaErrorDetails {
	var d MediaErrorDetails
	d.Set(errType, details)
	return d
}

// Set will set the receiving MediaErrorDetails with MediaErrorType and extra details (if any).
func (d *MediaErrorDetails) Set(errType MediaErrorType, details uint16) {
	(*d) = MediaErrorDetails(packu16s(uint16(errType), details))
}

// Type returns embedded MediaErrorType within details.
func (d MediaErrorDetails) Type() MediaErrorType {
	const bits = 16
	return MediaErrorType(uint16(d >> bits)) // nolint:gosec
}

// Details returns extra details related to Type(), if any are set.
func (d MediaErrorDetails) Details() uint16 {
	const bits = 16
	const mask = (1 << bits) - 1
	return uint16(d & mask) // nolint:gosec
}

// SupportsRetry returns whether error supports a re-attempt
// to cache the media, i.e. due to it likely being transient.
func (d MediaErrorDetails) SupportsRetry() bool {
	switch d.Type() {

	// Either no error was encountered and it was
	// later uncached, or the original fetch was
	// interrupted by cancelled request etc.
	case MediaErrorTypeNone,
		MediaErrorTypeInterrupt:
		return true

	// All policy and media codec /
	// processing errors are permanent.
	case MediaErrorTypePolicy,
		MediaErrorTypeCodec:
		return false

	// On timeout errors we can retry, others
	// are more likely to be permanent.
	case MediaErrorTypeNetwork:
		return d.Details() == MediaErrorTypeNetwork_Timeout

	// HTTP response code errors
	// can be handled granularly
	// depending on situation.
	case MediaErrorTypeHTTP:
		switch code := d.Details(); {

		// 400-403 type errors (e.g. auth, forbidden, bad request)
		// *can* be transient e.g. due to bugs. Others in the 4xx
		// range are generally more permanent (e.g. not found).
		case code >= 404:
			return false

		// More likely to be
		// a temporary error.
		case code >= 500:
			return true

		// All else
		// we deny.
		default:
			return false
		}
	}

	// Default to yes.
	return true
}

// String returns a frontend API (and log) string describing error details.
func (d MediaErrorDetails) String() string {
	switch errType := d.Type(); errType {
	case MediaErrorTypeNone:
		return "none"
	case MediaErrorTypePolicy:
		switch d.Details() {
		case MediaErrorTypePolicy_Size:
			return "file size limit reached"
		case MediaErrorTypePolicy_Domain:
			return "domain media policy"
		default:
			return "configuration policy"
		}
	case MediaErrorTypeInterrupt:
		return "connection interrupted"
	case MediaErrorTypeHTTP:
		status := int(d.Details())
		return "http response (status code: " + strconv.Itoa(status) +
			" " + http.StatusText(status) + ")"
	case MediaErrorTypeNetwork:
		switch d.Details() {
		case MediaErrorTypeNetwork_Timeout:
			return "network timeout"
		default:
			return "network error"
		}
	case MediaErrorTypeCodec:
		switch d.Details() {
		case MediaErrorTypeCodec_Unsupported:
			return "unsupported media type"
		default:
			return "media processing error"
		}
	default:
		return "unclassified"
	}
}

func packu16s(u1, u2 uint16) uint32 {
	const bits = 16
	const mask = (1 << bits) - 1
	return uint32(u1)<<bits | uint32(u2)&mask
}
