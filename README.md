chunkcidr
=========

Simple Go tool and package to chunk a CIDR into smaller CIDRs.

Usage (CLI):

Build:

    go build ./cmd/chunkcidr

Run:

    chunkcidr -cidr 192.168.0.0/24 -prefix 26

This prints the list of /26 subnets that cover the /24.

You can also pass a CIDR via stdin (one per line). Example:

    echo 199.66.248.0/29 | chunkcidr -prefix 30

This will print the /30 subnets that cover the /29.

Library usage:

    import "chunkcidr/pkg/chunker"

    subs, err := chunker.ChunkCIDR("192.168.0.0/24", 26)

Notes:
- The tool expects the target chunk size as a prefix length (e.g. 26 for /26).
- It supports both IPv4 and IPv6.

Additional option:
- You can specify a chunk by number of IPs using `-size`. Size must be a power-of-two (1,2,4,8...). For example `-size 8` will produce /29 subnets for IPv4.

Examples:

    chunkcidr -cidr 192.168.0.0/24 -prefix 26
    echo 199.66.248.0/29 | chunkcidr -size 8

Installation
------------

```bash
go install github.com/DarjaGFX/ChunkCIDR/cmd/chunkcidr@latest
```
