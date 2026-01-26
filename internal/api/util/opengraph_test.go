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
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type OpenGraphTestSuite struct {
	suite.Suite
}

func (suite *OpenGraphTestSuite) TestParseDescription() {
	tests := []struct {
		name, in, exp string
	}{
		{name: "shellcmd", in: `echo '\e]8;;http://example.com\e\This is a link\e]8;;\e'`, exp: `echo &#39;&bsol;e]8;;http://example.com&bsol;e&bsol;This is a link&bsol;e]8;;&bsol;e&#39;`},
		{name: "newlines", in: "test\n\ntest\ntest", exp: "test test test"},
	}

	for _, tt := range tests {
		tt := tt
		suite.Run(tt.name, func() {
			suite.Equal(tt.exp, ParseDescription(tt.in))
		})
	}
}

func (suite *OpenGraphTestSuite) TestWithAccountWithNote() {
	baseMeta := OGBase(&apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	})

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "<p>This is my profile, read it and weep! Weep then!</p>",
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	accountMeta := baseMeta.WithAccount(&apimodel.WebAccount{Account: acct})

	suite.EqualValues(OGMeta{
		Title:       "example person!!, @example_account@example.org",
		Type:        "profile",
		Locale:      "en",
		URL:         "https://example.org/@example_account",
		SiteName:    "example.org",
		Description: "This is my profile, read it and weep! Weep then!",
		Media: []OGMedia{
			{
				OGType: "image",
				Alt:    "Avatar for example_account",
				URL:    "https://example.org/avatar.jpg",
			},
			{
				// Instance avatar.
				OGType:   "image",
				URL:      "https://example.org/instance-avatar.webp",
				MIMEType: "image/webp",
			},
		},
		ArticlePublisher:     "",
		ArticleAuthor:        "",
		ArticleModifiedTime:  "",
		ArticlePublishedTime: "",
		ProfileUsername:      "example_account",
	}, *accountMeta)
}

func (suite *OpenGraphTestSuite) TestWithAccountNoNote() {
	baseMeta := OGBase(&apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	})

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "", // <- empty
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	accountMeta := baseMeta.WithAccount(&apimodel.WebAccount{Account: acct})

	suite.EqualValues(OGMeta{
		Title:       "example person!!, @example_account@example.org",
		Type:        "profile",
		Locale:      "en",
		URL:         "https://example.org/@example_account",
		SiteName:    "example.org",
		Description: "This GoToSocial user hasn't written a bio yet!",
		Media: []OGMedia{
			{
				OGType: "image",
				Alt:    "Avatar for example_account",
				URL:    "https://example.org/avatar.jpg",
			},
			{
				// Instance avatar.
				OGType:   "image",
				URL:      "https://example.org/instance-avatar.webp",
				MIMEType: "image/webp",
			},
		},
		ArticlePublisher:     "",
		ArticleAuthor:        "",
		ArticleModifiedTime:  "",
		ArticlePublishedTime: "",
		ProfileUsername:      "example_account",
	}, *accountMeta)
}

func (suite *OpenGraphTestSuite) TestWithStatus() {
	baseMeta := OGBase(&apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	})

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "", // <- empty
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	apiStatus := &apimodel.Status{
		ID:               "12345",
		CreatedAt:        "2025-01-18T00:00:00+00:00",
		EditedAt:         util.Ptr("2025-01-18T11:00:00+00:00"),
		Sensitive:        false,
		SpoilerText:      "",
		Visibility:       typeutils.VisToAPIVis(gtsmodel.VisibilityPublic),
		LocalOnly:        false,
		Language:         util.Ptr("en"),
		URI:              "https://example.org/statuses/12345",
		URL:              "https://example.org/@example_account/12345",
		Content:          "<b>test status</b>",
		Account:          acct,
		MediaAttachments: []*apimodel.Attachment{},
		Text:             "**test status**",
		ContentType:      apimodel.StatusContentTypeMarkdown,
	}

	status := &apimodel.WebStatus{
		Status:         apiStatus,
		SpoilerContent: "", // <- empty
		Account: &apimodel.WebAccount{
			Account:          acct,
			AvatarAttachment: nil,
			HeaderAttachment: nil,
			WebLayout:        gtsmodel.WebLayoutMicroblog.String(),
		},
	}

	statusMeta := baseMeta.WithStatus(status)

	suite.EqualValues(OGMeta{
		Title:       "Post by example person!!, @example_account@example.org",
		Type:        "article",
		Locale:      "en",
		URL:         "https://example.org/@example_account/12345",
		SiteName:    "example.org",
		Description: "**test status**",
		Media: []OGMedia{
			{
				OGType: "image",
				Alt:    "Avatar for example_account",
				URL:    "https://example.org/avatar.jpg",
			},
			{
				// Instance avatar.
				OGType:   "image",
				URL:      "https://example.org/instance-avatar.webp",
				MIMEType: "image/webp",
			},
		},
		ArticlePublisher:     "https://example.org/@example_account",
		ArticleAuthor:        "https://example.org/@example_account",
		ArticleModifiedTime:  "2025-01-18T11:00:00+00:00",
		ArticlePublishedTime: "2025-01-18T00:00:00+00:00",
		ProfileUsername:      "",
	}, *statusMeta)
}

func (suite *OpenGraphTestSuite) TestWithStatusWithImage() {
	baseMeta := OGBase(&apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	})

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "", // <- empty
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	imageAttachment := &apimodel.Attachment{
		ID:               "00IMAGE00",
		Type:             "image",
		URL:              util.Ptr("https://example.org/@example_account/12345/example.png"),
		TextURL:          util.Ptr("https://example.org/@example_account/12345/example.png"),
		PreviewURL:       util.Ptr("https://example.org/@example_account/12345/small/example.png"),
		RemoteURL:        nil,
		PreviewRemoteURL: nil,
		Meta: &apimodel.MediaMeta{
			Original: apimodel.MediaDimensions{
				Width:  1920,
				Height: 1080,
				Size:   "1920x1080",
				Aspect: 1920.0 / 1080,
			},
			Small: apimodel.MediaDimensions{
				Width:  320,
				Height: 240,
				Size:   "320x240",
				Aspect: 320.0 / 240,
			},
			Focus: nil,
		},
		Description: util.Ptr("an example image"),
		Blurhash:    util.Ptr("LKE3VIw}0KD%a2o{M|t7NFWps:t7"), // <- from testmodels
	}

	anotherImageAttachment := &apimodel.Attachment{
		ID:               "00IMAGE11",
		Type:             "image",
		URL:              util.Ptr("https://example.org/@example_account/12345/example2.png"),
		TextURL:          util.Ptr("https://example.org/@example_account/12345/example2.png"),
		PreviewURL:       util.Ptr("https://example.org/@example_account/12345/small/example2.png"),
		RemoteURL:        nil,
		PreviewRemoteURL: nil,
		Meta: &apimodel.MediaMeta{
			Original: apimodel.MediaDimensions{
				Width:  1000,
				Height: 1000,
				Size:   "1000x1000",
				Aspect: 1,
			},
			Small: apimodel.MediaDimensions{
				Width:  200,
				Height: 200,
				Size:   "200x200",
				Aspect: 1,
			},
			Focus: nil,
		},
		Description: util.Ptr("another example image"),
		Blurhash:    util.Ptr("LNABP8o#Dge,S6M}axxVEQjYxWbH"), // <- from testmodels
	}

	apiStatus := &apimodel.Status{
		ID:               "12345",
		CreatedAt:        "2025-01-18T00:00:00+00:00",
		EditedAt:         util.Ptr("2025-01-18T11:00:00+00:00"),
		Sensitive:        false,
		SpoilerText:      "",
		Visibility:       typeutils.VisToAPIVis(gtsmodel.VisibilityPublic),
		LocalOnly:        false,
		Language:         util.Ptr("en"),
		URI:              "https://example.org/statuses/12345",
		URL:              "https://example.org/@example_account/12345",
		Content:          "<b>test status</b>",
		Account:          acct,
		MediaAttachments: []*apimodel.Attachment{imageAttachment, anotherImageAttachment},
		Text:             "**test status**",
		ContentType:      apimodel.StatusContentTypeMarkdown,
	}

	webAttachment := &apimodel.WebAttachment{
		Attachment:       imageAttachment,
		Sensitive:        false,
		MIMEType:         "image/png",
		PreviewMIMEType:  "image/png",
		ParentStatusLink: "https://example.org/@example_account/12345",
	}

	anotherWebAttachment := &apimodel.WebAttachment{
		Attachment:       anotherImageAttachment,
		Sensitive:        false,
		MIMEType:         "image/png",
		PreviewMIMEType:  "image/png",
		ParentStatusLink: "https://example.org/@example_account/12345",
	}

	status := &apimodel.WebStatus{
		Status:           apiStatus,
		SpoilerContent:   "", // <- empty
		MediaAttachments: []*apimodel.WebAttachment{webAttachment, anotherWebAttachment},
		Account: &apimodel.WebAccount{
			Account:          acct,
			AvatarAttachment: nil,
			HeaderAttachment: nil,
			WebLayout:        gtsmodel.WebLayoutMicroblog.String(),
		},
	}

	statusMeta := baseMeta.WithStatus(status)

	suite.EqualValues(OGMeta{
		Title:       "Post by example person!!, @example_account@example.org",
		Type:        "article",
		Locale:      "en",
		URL:         "https://example.org/@example_account/12345",
		SiteName:    "example.org",
		Description: "**test status**",
		Media: []OGMedia{
			{
				OGType:   "image",
				Alt:      "an example image",
				URL:      "https://example.org/@example_account/12345/example.png",
				MIMEType: "image/png",
				Width:    "1920",
				Height:   "1080",
			},
		},
		ArticlePublisher:     "https://example.org/@example_account",
		ArticleAuthor:        "https://example.org/@example_account",
		ArticleModifiedTime:  "2025-01-18T11:00:00+00:00",
		ArticlePublishedTime: "2025-01-18T00:00:00+00:00",
		ProfileUsername:      "",
	}, *statusMeta)
}

func TestOpenGraphTestSuite(t *testing.T) {
	suite.Run(t, &OpenGraphTestSuite{})
}
