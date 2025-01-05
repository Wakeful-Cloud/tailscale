# Tailscale

[![GitHub Synchronization Status](https://img.shields.io/github/actions/workflow/status/wakeful-cloud/tailscale/sync.yml?label=Synchronization&style=flat-square)](https://github.com/wakeful-cloud/tailscale/actions/workflows/sync.yml)
[![GitHub Build Status](https://img.shields.io/github/actions/workflow/status/wakeful-cloud/tailscale/build.yml?label=Build&style=flat-square)](https://github.com/wakeful-cloud/tailscale/actions/workflows/build.yml)

Fork of the [Tailscale](https://github.com/tailscale/tailscale) that auto-synchronizes with upstream and
auto-builds a highly [space-optimized Tailscale client](https://tailscale.com/kb/1207/small-tailscale) for a variety of architectures (running Linux).

To download, [find the artifacts](https://github.com/actions/upload-artifact?tab=readme-ov-file#where-does-the-upload-go) for the latest build [here](https://github.com/wakeful-cloud/tailscale/actions/workflows/build.yml).

## FAQ

### What are the differences between this fork and upstream?
This fork uses a modified [`build_dist.sh`](./build_dist.sh) to support cross-compilation and [UPX](https://upx.github.io) for minification. Beyond
that, this fork adds the [GitHub actions](.github/workflows) required for automatic synchronization
and building.

### What architectures are supported?
* i386 (`linux/386`)
* X86 64-bit (`linux/amd64`)
* Arm 32-bit (`linux/arm`)
* Arm 64-bit (`linux/arm64`)
* MIPS 32-bit big endian (`linux/mips`)
* MIPS 32-bit little endian (`linux/mipsle`)
* MIPS 64-bit big endian (`linux/mips64`)[^1]
* MIPS 64-bit little endian (`linux/mips64le`)[^1]
* RISC-V 64-bit (`linux/riscv64`)[^1]

For background on which parts of Tailscale are open source and why,
see [https://tailscale.com/opensource/](https://tailscale.com/opensource/).

## Using

We serve packages for a variety of distros and platforms at
[https://pkgs.tailscale.com](https://pkgs.tailscale.com/).

## Other clients

The [macOS, iOS, and Windows clients](https://tailscale.com/download)
use the code in this repository but additionally include small GUI
wrappers. The GUI wrappers on non-open source platforms are themselves
not open source.

## Building

We always require the latest Go release, currently Go 1.23. (While we build
releases with our [Go fork](https://github.com/tailscale/go/), its use is not
required.)

```
go install tailscale.com/cmd/tailscale{,d}
```

If you're packaging Tailscale for distribution, use `build_dist.sh`
instead, to burn commit IDs and version info into the binaries:

```
./build_dist.sh tailscale.com/cmd/tailscale
./build_dist.sh tailscale.com/cmd/tailscaled
```

If your distro has conventions that preclude the use of
`build_dist.sh`, please do the equivalent of what it does in your
distro's way, so that bug reports contain useful version information.

## Bugs

Please file any issues about this code or the hosted service on
[the issue tracker](https://github.com/tailscale/tailscale/issues).

## Contributing

PRs welcome! But please file bugs. Commit messages should [reference
bugs](https://docs.github.com/en/github/writing-on-github/autolinked-references-and-urls).

We require [Developer Certificate of
Origin](https://en.wikipedia.org/wiki/Developer_Certificate_of_Origin)
`Signed-off-by` lines in commits.

See `git log` for our commit message style. It's basically the same as
[Go's style](https://go.dev/wiki/CommitMessage).

## About Us

[Tailscale](https://tailscale.com/) is primarily developed by the
people at https://github.com/orgs/tailscale/people. For other contributors,
see:

* https://github.com/tailscale/tailscale/graphs/contributors
* https://github.com/tailscale/tailscale-android/graphs/contributors

## Legal

WireGuard is a registered trademark of Jason A. Donenfeld.
