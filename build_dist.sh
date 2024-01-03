#!/usr/bin/env sh
#
# Runs `go build` with flags configured for binary distribution. All
# it does differently from `go build` is burn git commit and version
# information into the binaries, so that we can track down user
# issues.
#
# If you're packaging Tailscale for a distro, please consider using
# this script, or executing equivalent commands in your
# distro-specific build system.

# Tools
GO="${GO:-go}"
UPX_BIN="${UPX_BIN:-upx}"

# Ensure tools are installed
for TOOL in $GO $UPX_BIN; do
	if [ ! -x "$(command -v $TOOL)" ]; then
		echo "Error: $TOOL is not installed." >&2
		exit 1
	fi
done

# Clean
rm -rf dist

for GOARCH in 386 amd64 arm arm64 mips mipsle mips64 mips64le riscv64; do
	# Log
	echo "Building for $GOARCH..."

	# Update the environment variables
	eval `CGO_ENABLED=0 GOOS=$($GO env GOHOSTOS) GOARCH=$($GO env GOHOSTARCH) $GO run ./cmd/mkversion`
	VERSION_SHORT="$VERSION_SHORT-minified"
	VERSION_LONG="$VERSION_LONG-minified"

	# Generate the build information
	# See https://tailscale.com/kb/1207/small-tailscale#step-1-building-tailscale
	TAGS="ts_include_cli,ts_omit_aws,ts_omit_bird,ts_omit_tap,ts_omit_kube"
	LDFLAGS="-s -w -X tailscale.com/version.longStamp=${VERSION_LONG} -X tailscale.com/version.shortStamp=${VERSION_SHORT}"
	OUT="dist/tailscale-${VERSION_SHORT}-$GOARCH"

	# Build
	GOOS=linux GOARCH=$GOARCH $GO build -tags "$TAGS" -ldflags "$LDFLAGS" -o $OUT ./cmd/tailscaled

	# Log
	echo "Built $OUT"

	# Minify
	OLD_SIZE=$(stat -c%s $OUT)
	if $UPX_BIN --lzma --best $OUT > /dev/null 2>&1 ; then
		NEW_SIZE=$(stat -c%s $OUT)
	
		# Log
		echo "Compressed $OUT from $OLD_SIZE B to $NEW_SIZE B (Savings: ~$((100 * ($OLD_SIZE - $NEW_SIZE) / $OLD_SIZE))%)"
	else
		echo "Failed to compress $OUT (Size: $OLD_SIZE B)"
	fi
done