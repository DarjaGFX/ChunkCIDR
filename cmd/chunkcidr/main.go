package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/DarjaGFX/ChunkCIDR/pkg/chunker"
)

func main() {
	cidr := flag.String("cidr", "", "CIDR to split, e.g. 192.168.0.0/24")
	prefix := flag.Int("prefix", -1, "Target prefix length for chunks (e.g. 26)")
	size := flag.Int("size", -1, "Target chunk size in number of IPs (must be power of two, e.g. 8 for /29)")
	help := flag.Bool("help", false, "Show help")
	// short form -h
	h := flag.Bool("h", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  -cidr string\n    \tCIDR to split, e.g. 192.168.0.0/24 (optional if CIDRs are provided via stdin)\n")
		fmt.Fprintf(os.Stderr, "  -prefix int\n    \tTarget prefix length for chunks (e.g. 26)\n")
		fmt.Fprintf(os.Stderr, "  -size int\n    \tTarget chunk size in number of IPs (must be power of two, e.g. 8 for /29)\n")
		fmt.Fprintf(os.Stderr, "  -h, -help\n    \tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -cidr 192.168.0.0/24 -prefix 26\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  echo 199.66.248.0/29 | %s -size 8\n", os.Args[0])
	}
	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	if *prefix == -1 && *size == -1 {
		fmt.Fprintf(os.Stderr, "usage: -cidr CIDR (-prefix N | -size S)\n")
		os.Exit(2)
	}

	if *prefix != -1 && *size != -1 {
		fmt.Fprintf(os.Stderr, "error: specify either -prefix or -size, not both\n")
		os.Exit(2)
	}

	// If -cidr provided, process single CIDR. Otherwise read CIDRs from stdin (one per line).
	if strings.TrimSpace(*cidr) != "" {
		var subs []string
		var err error
		if *prefix != -1 {
			subs, err = chunker.ChunkCIDR(*cidr, *prefix)
		} else {
			subs, err = chunker.ChunkCIDRBySize(*cidr, *size)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		for _, s := range subs {
			fmt.Println(s)
		}
		return
	}

	// Read from stdin
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if fi.Mode()&os.ModeCharDevice != 0 {
		// No piped input and no -cidr
		fmt.Fprintf(os.Stderr, "usage: -cidr CIDR -prefix N\nOr: echo CIDR | chunkcidr -prefix N\n")
		os.Exit(2)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var subs []string
		var err error
		if *prefix != -1 {
			subs, err = chunker.ChunkCIDR(line, *prefix)
		} else {
			subs, err = chunker.ChunkCIDRBySize(line, *size)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error (%s): %v\n", line, err)
			// continue processing remaining lines rather than exiting
			continue
		}
		for _, s := range subs {
			fmt.Println(s)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}
}
