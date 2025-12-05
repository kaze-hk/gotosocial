#!/bin/bash

# Determine latest git tag tarball URL from github
URL=$(wget -q -O - 'https://api.github.com/repos/mathiasbynens/emoji-test-regex-pattern/tags' | \
      grep -Eo '"https://api.github.com/repos/mathiasbynens/emoji-test-regex-pattern/tarball/refs/tags/v[0-9\.]+"' | \
      head -n 1 | cut -d'"' -f2)

# Download latest tarball to tmpfile
wget -q -O /tmp/latest-regex.tgz "$URL"
trap 'rm /tmp/latest-regex.tgz' exit

# Extract the C++ RE2 expression text from tarball into variable
REGEX=$(tar --wildcards -xf /tmp/latest-regex.tgz */dist/latest/cpp-re2.txt \
            --to-stdout)

cat << EOF > emoji.go
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

package regexes

/*
      Emoji regex courtesy of https://github.com/mathiasbynens/emoji-test-regex-pattern.
      Matches a single unicode 17 emoji.

      See:

            https://raw.githubusercontent.com/mathiasbynens/emoji-test-regex-pattern/refs/heads/main/dist/emoji-17.0/cpp-re2.txt

            Copyright Mathias Bynens <https://mathiasbynens.be/>

            Permission is hereby granted, free of charge, to any person obtaining
            a copy of this software and associated documentation files (the
            "Software"), to deal in the Software without restriction, including
            without limitation the rights to use, copy, modify, merge, publish,
            distribute, sublicense, and/or sell copies of the Software, and to
            permit persons to whom the Software is furnished to do so, subject to
            the following conditions:

            The above copyright notice and this permission notice shall be
            included in all copies or substantial portions of the Software.

            THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
            EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
            MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
            NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
            LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
            OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
            WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

const unicodeEmoji = \`${REGEX}\`
EOF

# And then gofmt, to
# appease the linter...
gofmt -w emoji.go
