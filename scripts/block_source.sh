#!/bin/bash

ifi="$1"
if [ -z "$ifi" ]; then
	echo "You have to pass the name of the interface that you want to block for this to work out:"
	echo "Example: $0 en0"
	exit 0
fi

curl -H "Content-Type: application/json" -d "{\"reason\": \"blocked by $0 script\"}" -X POST http://localhost:7764/sources/"$ifi"/block
