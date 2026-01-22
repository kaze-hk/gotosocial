package pub

import (
	"code.superseriousbusiness.org/activity/streams/vocab"
)

// Activity represents any ActivityStreams Activity type.
//
// The Activity types provided in the streams package implement this.
type Activity interface {
	// Activity is also
	// a vocab.Type.
	vocab.Type

	// GetActivityStreamsActor returns "actor"
	// if it exists, or nil if it doesn't.
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
	// SetActivityStreamsActor sets "actor".
	SetActivityStreamsActor(i vocab.ActivityStreamsActorProperty)

	// GetActivityStreamsAttributedTo returns "attributedTo"
	// if it exists, or nil if it doesn't.
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
	// SetActivityStreamsAttributedTo sets "attributedTo".
	SetActivityStreamsAttributedTo(i vocab.ActivityStreamsAttributedToProperty)

	// GetActivityStreamsAudience returns "audience"
	// property if it exists, or nil if it doesn't.
	GetActivityStreamsAudience() vocab.ActivityStreamsAudienceProperty
	// SetActivityStreamsAudience sets "audience".
	SetActivityStreamsAudience(t vocab.ActivityStreamsAudienceProperty)

	// GetActivityStreamsObject returns the "object"
	// property if it exists, or nil if it doesn't.
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
	// SetActivityStreamsObject sets "object".
	SetActivityStreamsObject(i vocab.ActivityStreamsObjectProperty)

	// GetActivityStreamsInstrument returns the "instrument"
	// property if it exists, or nil if it doesn't.
	GetActivityStreamsInstrument() vocab.ActivityStreamsInstrumentProperty
	// SetActivityStreamsInstrument sets "instrument".
	SetActivityStreamsInstrument(i vocab.ActivityStreamsInstrumentProperty)

	// GetActivityStreamsCc returns "cc"
	// if it exists, or nil if it doesn't.
	GetActivityStreamsCc() vocab.ActivityStreamsCcProperty
	// SetActivityStreamsCc sets "cc".
	SetActivityStreamsCc(i vocab.ActivityStreamsCcProperty)

	// GetActivityStreamsTo returns "to"
	// if it exists, or nil if it doesn't.
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
	// SetActivityStreamsTo sets "to".
	SetActivityStreamsTo(i vocab.ActivityStreamsToProperty)

	// GetActivityStreamsBto returns "bto"
	// if it exists, or nil if it doesn't.
	GetActivityStreamsBto() vocab.ActivityStreamsBtoProperty
	// SetActivityStreamsBto sets "bto".
	SetActivityStreamsBto(i vocab.ActivityStreamsBtoProperty)

	// GetActivityStreamsBcc returns "bcc"
	// if it exists, or nil if it doesn't.
	GetActivityStreamsBcc() vocab.ActivityStreamsBccProperty
	// SetActivityStreamsBcc sets "bcc".
	SetActivityStreamsBcc(i vocab.ActivityStreamsBccProperty)
}
