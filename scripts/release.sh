#
# MIT License

# Copyright (C) 2018 Daniel Morandini

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#

#!/bin/bash

set -e

conf=.goreleaser.yml

function release {
	echo "Starting release pipeline..."
	prepare

	echo "Please insert git tag to be used for the release: "
	read version

	echo "Proceding will remove dist/ folder & will add/substitute a new release/tag $version. Continue?"
	echo "Yes/no"
	read opt
	if [ "$opt" = "no" ]; then
		echo "quitting..."
		return 1
	fi

	tag $version

	echo "Executing goreleaser..."
	goreleaser release --rm-dist
}

function prepare {
	command -v goreleaser >/dev/null 2>&1 || { echo >&2 "goreleaser not installed. Quitting..."; exit 1; }

	if [ ! -f $conf ]; then
		echo file $conf does not exits!
		exit -1
	fi
}

function snapshot {
	echo "Proceding will remove dist/ folder. Continue?"
	echo "Yes/no"
	read opt
	if [ "$opt" = "no" ]; then
		echo "quitting..."
		return 1
	fi

	echo "Executing goreleaser..."
	goreleaser release --rm-dist --snapshot --skip-publish
}

function tag {
	echo "Creating tag $1..."
	git tag -a "$1" -m "Release $version"
	git push origin "$1"
}

function uploadSnaps {
	# allow multiple upload failures
	set +e
	echo "Uploading snaps..."
	local files=( `find dist -name "*.snap" -type f` )
 	echo "Snaps found: ${files[@]}."
	for f in ${files[@]}; do
		snapcraft push $f
	done
 	set -e
 	echo "Remember to execute \`snapcraft release <snap name> revision channel\` for each revision provided!"
	echo "Find revisions using \`snapcraft list-revisions booster\`"
}

# main

OPTS="Snapshot Release UploadSnaps Exit"
select opt in $OPTS; do
	if [ "$opt" = "Snapshot" ]; then
		snapshot
	elif [ "$opt" = "Release" ]; then
		release
	elif [ "$opt" = "UploadSnaps" ]; then
		uploadSnaps
	else
		exit
	fi
done

exit 0
