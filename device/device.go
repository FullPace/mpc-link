// Package device holds the programs Padbangers Link puts on the MPC.
package device

import _ "embed"

// SSHFS is sshfs 3.7.3 for armv7, statically linked (Alpine, musl; see build-sshfs.sh).
//
//go:embed sshfs
var SSHFS []byte
