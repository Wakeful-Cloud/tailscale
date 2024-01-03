#!/usr/bin/env sh

# Tools
GO="${GO:-go}"

# Ensure tools are installed
for TOOL in $GO $UPX_BIN; do
	if [ ! -x "$(command -v $TOOL)" ]; then
		echo "Error: $TOOL is not installed." >&2
		exit 1
	fi
done

# Update the environment variables
eval `${GO} run ./cmd/mkversion`
VERSION_LONG="${VERSION_LONG}_minified"
VERSION_SHORT="${VERSION_SHORT}_minified"

# Print the version
echo $VERSION_SHORT