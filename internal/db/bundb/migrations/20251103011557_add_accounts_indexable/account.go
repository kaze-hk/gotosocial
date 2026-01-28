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
	"crypto/rsa"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Account represents either a local or a remote ActivityPub actor.
// https://www.w3.org/TR/activitypub/#actor-objects
type Account struct {
	// Database ID of the account.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Datetime when the account was created.
	// Corresponds to ActivityStreams `published` prop.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Datetime when was the account was last updated,
	// ie., when the actor last sent out an Update
	// activity, or if never, when it was `published`.
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Datetime when the account was last fetched /
	// dereferenced by this GoToSocial instance.
	FetchedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Username of the account.
	//
	// Corresponds to AS `preferredUsername` prop, which gives
	// no uniqueness guarantee. However, we do enforce uniqueness
	// for it as, in practice, it always is and we rely on this.
	Username string `bun:",nullzero,notnull,unique:accounts_username_domain_uniq"`

	// Domain of the account, discovered via webfinger.
	//
	// Null if this is a local account, otherwise
	// something like `example.org`.
	Domain string `bun:",nullzero,unique:accounts_username_domain_uniq"`

	// Database ID of the account's avatar MediaAttachment, if set.
	AvatarMediaAttachmentID string `bun:"type:CHAR(26),nullzero"`

	// MediaAttachment corresponding to AvatarMediaAttachmentID.
	AvatarMediaAttachment *gtsmodel.MediaAttachment `bun:"-"`

	// URL of the avatar media.
	//
	// Null for local accounts.
	AvatarRemoteURL string `bun:",nullzero"`

	// Database ID of the account's header MediaAttachment, if set.
	HeaderMediaAttachmentID string `bun:"type:CHAR(26),nullzero"`

	// MediaAttachment corresponding to HeaderMediaAttachmentID.
	HeaderMediaAttachment *gtsmodel.MediaAttachment `bun:"-"`

	// URL of the header media.
	//
	// Null for local accounts.
	HeaderRemoteURL string `bun:",nullzero"`

	// Display name for this account, if set.
	//
	// Corresponds to the ActivityStreams `name` property.
	//
	// If null, fall back to username for display purposes.
	DisplayName string `bun:",nullzero"`

	// Database IDs of any emojis used in
	// this account's bio, display name, etc
	EmojiIDs []string `bun:"emojis,array"`

	// Emojis corresponding to EmojiIDs.
	Emojis []*gtsmodel.Emoji `bun:"-"`

	// A slice of of key/value fields that
	// this account has added to their profile.
	//
	// Corresponds to schema.org PropertyValue types in `attachments`.
	Fields []*gtsmodel.Field `bun:",nullzero"`

	// The raw (unparsed) content of fields that this
	// account has added to their profile, before
	// conversion to HTML.
	//
	// Only set for local accounts.
	FieldsRaw []*gtsmodel.Field `bun:",nullzero"`

	// A note that this account has on their profile
	// (ie., the account's bio/description of themselves).
	//
	// Corresponds to the ActivityStreams `summary` property.
	Note string `bun:",nullzero"`

	// The raw (unparsed) version of Note, before conversion to HTML.
	//
	// Only set for local accounts.
	NoteRaw string `bun:",nullzero"`

	// ActivityPub URI/IDs by which this account is also known.
	//
	// Corresponds to the ActivityStreams `alsoKnownAs` property.
	AlsoKnownAsURIs []string `bun:"also_known_as_uris,array"`

	// Accounts matching AlsoKnownAsURIs.
	AlsoKnownAs []*Account `bun:"-"`

	// URI/ID to which the account has (or claims to have) moved.
	//
	// Corresponds to the ActivityStreams `movedTo` property.
	//
	// Even if this field is set the move may not yet have been
	// processed. Check `move` for this.
	MovedToURI string `bun:",nullzero"`

	// Account matching MovedToURI.
	MovedTo *Account `bun:"-"`

	// ID of a Move in the database for this account.
	// Only set if we received or created a Move activity
	// for which this account URI was the origin.
	MoveID string `bun:"type:CHAR(26),nullzero"`

	// Move corresponding to MoveID, if set.
	Move *gtsmodel.Move `bun:"-"`

	// True if account requires manual approval of Follows.
	//
	// Corresponds to AS `manuallyApprovesFollowers` prop.
	Locked *bool `bun:",nullzero,notnull,default:true"`

	// True if account has opted in to being shown in
	// directories and exposed to search engines.
	//
	// Corresponds to the toot `discoverable` property.
	Discoverable *bool `bun:",nullzero,notnull,default:false"`

	// True if account has opted into its posts being indexed for full-text search.
	//
	// Corresponds to the toot `indexable` property.
	Indexable *bool `bun:",nullzero,notnull,default:false"`

	// ActivityPub URI/ID for this account.
	//
	// Must be set, must be unique.
	URI string `bun:",nullzero,notnull,unique"`

	// URL at which a web representation of this
	// account should be available, if set.
	//
	// Corresponds to ActivityStreams `url` prop.
	URL string `bun:",nullzero"`

	// URI of the actor's inbox.
	//
	// Corresponds to ActivityPub `inbox` property.
	//
	// According to AP this MUST be set, but some
	// implementations don't set it for service actors.
	InboxURI string `bun:",nullzero"`

	// URI/ID of this account's sharedInbox, if set.
	//
	// Corresponds to ActivityPub `endpoints.sharedInbox`.
	//
	// Gotcha warning: this is a string pointer because
	// it has three possible states:
	//
	//   1. null: We don't know (yet) if actor has a shared inbox.
	//   2. empty: We know it doesn't have a shared inbox.
	//   3. not empty: We know it does have a shared inbox.
	SharedInboxURI *string `bun:""`

	// URI/ID of the actor's outbox.
	//
	// Corresponds to ActivityPub `outbox` property.
	//
	// According to AP this MUST be set, but some
	// implementations don't set it for service actors.
	OutboxURI string `bun:",nullzero"`

	// URI/ID of the actor's following collection.
	//
	// Corresponds to ActivityPub `following` property.
	//
	// According to AP this SHOULD be set.
	FollowingURI string `bun:",nullzero"`

	// URI/ID of the actor's followers collection.
	//
	// Corresponds to ActivityPub `followers` property.
	//
	// According to AP this SHOULD be set.
	FollowersURI string `bun:",nullzero"`

	// URI/ID of the actor's featured collection.
	//
	// Corresponds to the Toot `featured` property.
	FeaturedCollectionURI string `bun:",nullzero"`

	// ActivityStreams type of the actor.
	//
	// Application, Group, Organization, Person, or Service.
	ActorType gtsmodel.AccountActorType `bun:",nullzero,notnull"`

	// Private key for signing http requests.
	//
	// Only defined for local accounts
	PrivateKey *rsa.PrivateKey `bun:""`

	// Public key for authorizing signed http requests.
	//
	// Defined for both local and remote accounts
	PublicKey *rsa.PublicKey `bun:",notnull"`

	// Dereferenceable location of this actor's public key.
	//
	// Corresponds to https://w3id.org/security/v1 `publicKey.id`.
	PublicKeyURI string `bun:",nullzero,notnull,unique"`

	// Datetime at which public key will expire/has expired,
	// and should be fetched again as appropriate.
	//
	// Only ever set for remote accounts.
	PublicKeyExpiresAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was marked as a "memorial",
	// ie., user owning the account has passed away.
	MemorializedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was set to
	// have all its media shown as sensitive.
	SensitizedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was silenced.
	SilencedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was suspended.
	SuspendedAt time.Time `bun:"type:timestamptz,nullzero"`

	// ID of the database entry that caused this account to
	// be suspended. Can be an account ID or a domain block ID.
	SuspensionOrigin string `bun:"type:CHAR(26),nullzero"`

	// gtsmodel.AccountSettings for this account.
	//
	// Local, non-instance-actor accounts only.
	Settings *gtsmodel.AccountSettings `bun:"-"`

	// gtsmodel.AccountStats for this account.
	Stats *gtsmodel.AccountStats `bun:"-"`

	// True if the actor hides to-public statusables
	// from unauthenticated public access via the web.
	// Default "false" if not set on the actor model.
	HidesToPublicFromUnauthedWeb *bool `bun:",nullzero,notnull,default:false"`

	// True if the actor hides cc-public statusables
	// from unauthenticated public access via the web.
	// Default "true" if not set on the actor model.
	HidesCcPublicFromUnauthedWeb *bool `bun:",nullzero,notnull,default:true"`
}
