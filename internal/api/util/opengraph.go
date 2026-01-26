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

package util

import (
	"html"
	"math"
	"strconv"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// OGMeta represents supported OpenGraph Meta tags
//
// see eg https://ogp.me/
type OGMeta struct {
	/* Vanilla og tags */

	Title       string // og:title
	Type        string // og:type
	Locale      string // og:locale
	URL         string // og:url
	SiteName    string // og:site_name
	Description string // og:description

	// Zero or more media entries of type image,
	// video, or audio (https://ogp.me/#array).
	Media []OGMedia

	/* Article tags. */

	ArticlePublisher     string // article:publisher
	ArticleAuthor        string // article:author
	ArticleModifiedTime  string // article:modified_time
	ArticlePublishedTime string // article:published_time

	/* Profile tags. */

	ProfileUsername string // profile:username

	/*
		Twitter card stuff
		https://developer.twitter.com/en/docs/twitter-for-websites/cards/overview/abouts-cards
	*/

	// Set to media URL for media posts.
	TwitterSummaryLargeImage string
	TwitterImageAlt          string
}

// OGMedia represents one OpenGraph media
// entry of type image, video, or audio.
type OGMedia struct {
	OGType   string // image/video/audio
	URL      string // og:${type}
	MIMEType string // og:${type}:type
	Width    string // og:${type}:width
	Height   string // og:${type}:height
	Alt      string // og:${type}:alt
}

// OGBase returns an *ogMeta suitable for serving at
// the base root of an instance. It also serves as a
// foundation for building account / status ogMeta.
func OGBase(instance *apimodel.InstanceV1) *OGMeta {
	var locale string
	if len(instance.Languages) > 0 {
		locale = instance.Languages[0]
	}

	og := &OGMeta{
		Title:       text.StripHTMLFromText(instance.Title) + " - GoToSocial",
		Type:        "website",
		Locale:      locale,
		URL:         instance.URI,
		SiteName:    instance.AccountDomain,
		Description: ParseDescription(instance.ShortDescription),
		Media: []OGMedia{
			{
				OGType:   "image",
				URL:      instance.Thumbnail,
				Alt:      instance.ThumbnailDescription,
				MIMEType: instance.ThumbnailType,
			},
		},
	}

	return og
}

// WithAccount uses the given account to build
// an ogMeta struct specific to that account.
// It's suitable for serving at account profile pages.
func (o *OGMeta) WithAccount(acct *apimodel.WebAccount) *OGMeta {
	o.Title = AccountTitle(acct, o.SiteName)
	o.ProfileUsername = acct.Username + "@" + o.SiteName
	o.Type = "profile"
	o.URL = acct.URL
	if acct.Note != "" {
		o.Description = ParseDescription(acct.Note)
	} else {
		const desc = "This GoToSocial user hasn't written a bio yet!"
		o.Description = desc
	}

	// Set avatar image.
	o.Media = []OGMedia{ogImgForAcct(acct)}
	return o
}

// util funct to return OGImage using account.
func ogImgForAcct(account *apimodel.WebAccount) OGMedia {
	ogMedia := OGMedia{
		OGType: "image",
		URL:    account.Avatar,
		Alt:    "Avatar for " + account.Username,
	}

	if desc := account.AvatarDescription; desc != "" {
		ogMedia.Alt += ": " + desc
	}

	// Check if the account
	// has an avatar set
	// that's not the default.
	a := account.AvatarAttachment
	if a == nil {
		// Nothing
		// left to do.
		return ogMedia
	}

	// Set the MIME type.
	ogMedia.MIMEType = a.MIMEType

	// Check width + height.
	width := a.Meta.Original.Width
	height := a.Meta.Original.Height

	// Find longest side.
	longest := max(
		width,
		height,
	)

	// Max width or length should
	// be 400 or this will look
	// very silly in previews.
	const max = 400
	if longest > max {
		multiplier := float32(max) / float32(longest)
		width = int(math.Round(float64(multiplier * float32(width))))
		height = int(math.Round(float64(multiplier * float32(height))))
	}

	ogMedia.Width = strconv.Itoa(width)
	ogMedia.Height = strconv.Itoa(height)
	return ogMedia
}

// WithStatus uses the given status to build
// and ogMeta struct specific to that status.
// It's suitable for serving at status pages.
func (o *OGMeta) WithStatus(status *apimodel.WebStatus) *OGMeta {
	// Set title to something like
	// "Display Name (@username@account.domain)"
	o.Title = AccountTitle(status.Account, o.SiteName)

	// It's a post not an article
	// but this is all we have.
	o.Type = "article"
	if status.Language != nil {
		o.Locale = *status.Language
	}

	// Self-explanatory.
	o.URL = status.URL

	// Derive description based on
	// sensitivity + media attachments.
	attachLen := len(status.MediaAttachments)
	attachSet := attachLen != 0
	cwSet := status.SpoilerText != ""
	contentSet := status.Text != ""

	switch {

	// If content warning is set then this
	// is a sensitive post by default and
	// we should not use the post content
	// at all in the description.
	case cwSet:
		if attachSet {
			o.Description = ParseDescription("Sensitive content [" + mediaCount(attachLen) + "]" + ": " + status.SpoilerText)
		} else {
			o.Description = ParseDescription("Sensitive content: " + status.SpoilerText)
		}

	// There's no content warning set but
	// the status is marked sensitive and
	// it has text content. We can use the
	// status content in the description
	// but warn that it's sensitive.
	case status.Sensitive && contentSet:
		if attachSet {
			o.Description = ParseDescription("Sensitive content [" + mediaCount(attachLen) + "]" + ": " + status.Text)
		} else {
			o.Description = ParseDescription("Sensitive content: " + status.Text)
		}

	// There's no content warning set
	// and no text content set, but
	// there are sensitive attachments.
	case status.Sensitive && attachSet:
		o.Description = "Sensitive media: " + mediaCount(attachLen)

	// Status isn't sensitive and there's
	// no content warning set, use the
	// post content in the description.
	case !status.Sensitive && contentSet:
		if attachSet {
			o.Description = ParseDescription("[" + mediaCount(attachLen) + "] " + status.Text)
		} else {
			o.Description = ParseDescription(status.Text)
		}

	// Status isn't sensitive and there's
	// no content warning or content set.
	case !status.Sensitive && !contentSet:
		if attachSet {
			o.Description = mediaCount(attachLen)
		} else {
			o.Description = ParseDescription("Post by " + o.Title)
		}

	// Fall back to
	// account title.
	default:
		o.Description = o.Title
	}

	o.ArticlePublisher = status.Account.URL
	o.ArticleAuthor = status.Account.URL
	o.ArticlePublishedTime = status.CreatedAt
	o.ArticleModifiedTime = util.PtrOrValue(status.EditedAt, status.CreatedAt)

	// Clear any existing medias.
	o.Media = []OGMedia{}

	// If media is sensitive then
	// don't append it to preview.
	if status.Sensitive {
		return o
	}

	// Add image / media previews.
	for _, a := range status.MediaAttachments {
		if a.Type == "unknown" {
			// Skip unknown.
			continue
		}

		// Start building entry
		// with common media tags.
		desc := util.PtrOrZero(a.Description)
		ogMedia := OGMedia{
			URL:      *a.URL,
			MIMEType: a.MIMEType,
			Alt:      desc,
		}

		// Add further tags
		// depending on type.
		switch a.Type {

		case "image":
			ogMedia.OGType = "image"
			ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
			ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)

			// If this image is the only piece of media,
			// set TwitterSummaryLargeImage to indicate
			// that a large image summary is preferred.
			if attachLen == 1 {
				o.TwitterSummaryLargeImage = *a.URL
				o.TwitterImageAlt = desc
			}

		case "audio":
			ogMedia.OGType = "audio"

		case "video", "gifv":
			ogMedia.OGType = "video"
			ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
			ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)
		}

		// Add this to our gathered entries.
		o.Media = append(o.Media, ogMedia)

		// Include static/thumb for non-visual files
		// (eg., audios) if they have a preview url set.
		if a.Type != "image" && a.PreviewURL != nil {
			o.Media = append(
				o.Media,
				OGMedia{
					OGType:   "image",
					URL:      *a.PreviewURL,
					MIMEType: a.PreviewMIMEType,
					Width:    strconv.Itoa(a.Meta.Small.Width),
					Height:   strconv.Itoa(a.Meta.Small.Height),
					Alt:      util.PtrOrZero(a.Description),
				},
			)
		}
	}

	return o
}

// AccountTitle parses a page title
// from account and accountDomain.
func AccountTitle(
	account *apimodel.WebAccount,
	accountDomain string,
) string {
	var displayName string
	if account.DisplayName != "" {
		displayName = account.DisplayName
	} else {
		displayName = account.Username
	}
	nameString := "@" + account.Username + "@" + accountDomain
	return displayName + " (" + nameString + ")"
}

// ParseDescription returns a string description which is
// safe to use as the content of a `content="..."` attribute.
func ParseDescription(in string) string {
	i := text.StripHTMLFromText(in)
	i = strings.ReplaceAll(i, "\n", " ")
	i = strings.Join(strings.Fields(i), " ")
	i = html.EscapeString(i)
	i = strings.ReplaceAll(i, `\`, "&bsol;")
	return truncate(i)
}

// truncate trims string
// to maximum 300 runes.
func truncate(s string) string {
	const truncateLen = 300

	r := []rune(s)
	if len(r) < truncateLen {
		// No need
		// to trim.
		return s
	}

	return string(r[:truncateLen-3]) + "…"
}

// mediaCount returns a
// neat media count string.
func mediaCount(attachLen int) string {
	switch attachLen {
	case 0:
		return ""
	case 1:
		return "1 media attachment"
	default:
		return strconv.FormatInt(int64(attachLen), 10) + " media attachments"
	}
}
