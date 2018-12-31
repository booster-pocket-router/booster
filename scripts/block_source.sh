#!/bin/bash

METHOD="$1"
NAME="$2"
HOST="http://booster.local:7764"

if [ -z "$METHOD" ]; then
	echo "You have to pass the name of the method that you want to execute for this to work out:"
	echo "Only POST or DELETE will produce valid results."
	echo "POST: creates the block policy"
	echo "DELETE: removes a previously created block policy"
	echo "Example: $0 POST en0"
	exit 0
fi
if [ -z "$NAME" ]; then
	echo "You have to pass the name of the interface that you want to block for this to work out:"
	echo "Example: $0 POST en0"
	exit 0
fi

curl -H "Content-Type: application/json" -d "{\"reason\": \"blocked by $0 script\"}" -X "$METHOD" "$HOST"/sources/"$NAME"/block.json
