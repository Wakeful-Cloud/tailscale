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

[^1]: UPX is not supported for these architectures, therefore the binaries are not as small as the others.