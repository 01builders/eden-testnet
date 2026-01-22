package cmd

import rollconf "github.com/evstack/ev-node/pkg/config"

const (
	flagBackwardPassphrase = rollconf.FlagPrefixEvnode + "signer.passphrase"
	flagBackwardJWTSecret  = rollconf.FlagPrefixEvnode + "evm.jwt-secret"
)
