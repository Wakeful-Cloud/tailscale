#!/usr/bin/env sh

# Tools
GO="${GO:-go}"

# Ensure tools are installed
for TOOL in $GO; do
	if [ ! -x "$(command -v $TOOL)" ]; then
		echo "Error: $TOOL is not installed." >&2
		exit 1
	fi
done

# Clean
rm -rf dist

# Build and minify binaries
mkdir -p dist
for GOARCH in 386 amd64 arm arm64 mips mipsle mips64 mips64le riscv64; do
	# Log
	echo "Building for ${GOARCH}..."

	# Update the environment variables
	eval `CGO_ENABLED=0 GOOS=$(${GO} env GOHOSTOS) GOARCH=$(${GO} env GOHOSTARCH) ${GO} run ./cmd/mkversion`
	VERSION_LONG="${VERSION_LONG}_small"
	VERSION_SHORT="${VERSION_SHORT}_small"

	# Generate the build information
	# See https://tailscale.com/kb/1207/small-tailscale#step-1-building-tailscale
	TAGS="ts_include_cli,ts_omit_aws,ts_omit_bird,ts_omit_tap,ts_omit_kube,ts_omit_completion,ts_omit_ssh,ts_omit_wakeonlan,ts_omit_capture,ts_omit_relayserver,ts_omit_taildrop,ts_omit_tpm"
	LDFLAGS="-s -w -X tailscale.com/version.longStamp=${VERSION_LONG} -X tailscale.com/version.shortStamp=${VERSION_SHORT}"
	DIST="dist/tailscale_${VERSION_SHORT}_${GOARCH}"
	OUT="${DIST}/tailscaled"

	# Build
	GOOS=linux GOARCH=${GOARCH} ${GO} build -tags "${TAGS}" -ldflags "${LDFLAGS}" -o "${OUT}" ./cmd/tailscaled

	# Log
	echo "Built ${OUT}"

	# Copy systemd files
	mkdir -p "${DIST}/systemd"
	cp cmd/tailscaled/tailscaled.defaults "${DIST}/systemd"
	cp cmd/tailscaled/tailscaled.service "${DIST}/systemd"
done