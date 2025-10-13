package cmd

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/evstack/ev-node/core/da"
	"github.com/evstack/ev-node/da/jsonrpc"
	"github.com/evstack/ev-node/node"
	"github.com/evstack/ev-node/sequencers/single"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/evstack/ev-node/execution/evm"

	"github.com/evstack/ev-node/core/execution"
	rollcmd "github.com/evstack/ev-node/pkg/cmd"
	"github.com/evstack/ev-node/pkg/config"
	genesispkg "github.com/evstack/ev-node/pkg/genesis"
	"github.com/evstack/ev-node/pkg/p2p"
	"github.com/evstack/ev-node/pkg/p2p/key"
	"github.com/evstack/ev-node/pkg/store"
)

var RunCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"node", "run"},
	Short:   "Run the evolve node with EVM execution client",
	RunE: func(cmd *cobra.Command, args []string) error {
		executor, err := createExecutionClient(cmd)
		if err != nil {
			return err
		}

		nodeConfig, err := rollcmd.ParseConfig(cmd)
		if err != nil {
			return err
		}

		logger := rollcmd.SetupLogger(nodeConfig.Log)

		// Attach logger to the EVM engine client if available
		if ec, ok := executor.(*evm.EngineClient); ok {
			ec.SetLogger(logger.With().Str("module", "engine_client").Logger())
		}

		headerNamespace := da.NamespaceFromString(nodeConfig.DA.GetNamespace())
		dataNamespace := da.NamespaceFromString(nodeConfig.DA.GetDataNamespace())

		logger.Info().Str("headerNamespace", headerNamespace.HexString()).Str("dataNamespace", dataNamespace.HexString()).Msg("namespaces")

		daJrpc, err := jsonrpc.NewClient(context.Background(), logger, nodeConfig.DA.Address, nodeConfig.DA.AuthToken, nodeConfig.DA.GasPrice, nodeConfig.DA.GasMultiplier, rollcmd.DefaultMaxBlobSize)
		if err != nil {
			return err
		}

		daAPI := newNamespaceMigrationDAAPI(daJrpc.DA, nodeConfig, migrations)

		datastore, err := store.NewDefaultKVStore(nodeConfig.RootDir, nodeConfig.DBPath, "eden-testnet")
		if err != nil {
			return err
		}

		genesisPath := filepath.Join(filepath.Dir(nodeConfig.ConfigPath()), "genesis.json")
		genesis, err := genesispkg.LoadGenesis(genesisPath)
		if err != nil {
			return fmt.Errorf("failed to load genesis: %w", err)
		}

		if genesis.DAStartHeight == 0 && !nodeConfig.Node.Aggregator {
			logger.Warn().Msg("da_start_height is not set in genesis.json, ask your chain developer")
		}

		singleMetrics, err := single.DefaultMetricsProvider(nodeConfig.Instrumentation.IsPrometheusEnabled())(genesis.ChainID)
		if err != nil {
			return err
		}

		sequencer, err := single.NewSequencer(
			context.Background(),
			logger,
			datastore,
			daAPI,
			[]byte(genesis.ChainID),
			nodeConfig.Node.BlockTime.Duration,
			singleMetrics,
			nodeConfig.Node.Aggregator,
		)
		if err != nil {
			return err
		}

		nodeKey, err := key.LoadNodeKey(filepath.Dir(nodeConfig.ConfigPath()))
		if err != nil {
			return err
		}

		p2pClient, err := p2p.NewClient(nodeConfig.P2P, nodeKey.PrivKey, datastore, genesis.ChainID, logger, nil)
		if err != nil {
			return err
		}

		return rollcmd.StartNode(logger, cmd, executor, sequencer, daAPI, p2pClient, datastore, nodeConfig, genesis, node.NodeOptions{})
	},
}

func init() {
	config.AddFlags(RunCmd)
	addFlags(RunCmd)
}

func createExecutionClient(cmd *cobra.Command) (execution.Executor, error) {
	// Read execution client parameters from flags
	ethURL, err := cmd.Flags().GetString(evm.FlagEvmEthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s' flag: %w", evm.FlagEvmEthURL, err)
	}
	engineURL, err := cmd.Flags().GetString(evm.FlagEvmEngineURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s' flag: %w", evm.FlagEvmEngineURL, err)
	}
	jwtSecret, err := cmd.Flags().GetString(evm.FlagEvmJWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s' flag: %w", evm.FlagEvmJWTSecret, err)
	}
	genesisHashStr, err := cmd.Flags().GetString(evm.FlagEvmGenesisHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s' flag: %w", evm.FlagEvmGenesisHash, err)
	}
	feeRecipientStr, err := cmd.Flags().GetString(evm.FlagEvmFeeRecipient)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s' flag: %w", evm.FlagEvmFeeRecipient, err)
	}

	// Convert string parameters to Ethereum types
	genesisHash := common.HexToHash(genesisHashStr)
	feeRecipient := common.HexToAddress(feeRecipientStr)

	return evm.NewEngineExecutionClient(ethURL, engineURL, jwtSecret, genesisHash, feeRecipient)
}

// addFlags adds flags related to the EVM execution client
func addFlags(cmd *cobra.Command) {
	cmd.Flags().String(evm.FlagEvmEthURL, "http://localhost:8545", "URL of the Ethereum JSON-RPC endpoint")
	cmd.Flags().String(evm.FlagEvmEngineURL, "http://localhost:8551", "URL of the Engine API endpoint")
	cmd.Flags().String(evm.FlagEvmJWTSecret, "", "The JWT secret for authentication with the execution client")
	cmd.Flags().String(evm.FlagEvmGenesisHash, "", "Hash of the genesis block")
	cmd.Flags().String(evm.FlagEvmFeeRecipient, "", "Address that will receive transaction fees")
}

var migrations = map[uint64]namespaces{
	8130490: {
		namespace:     "rollkit-headers",
		dataNamespace: "rollkit-data",
	},
}

// namespaces defines the namespace used for namespace migration
type namespaces struct {
	namespace     string
	dataNamespace string
}

func (n namespaces) GetNamespace() string {
	return n.namespace
}

func (n namespaces) GetDataNamespace() string {
	if n.dataNamespace == "" {
		return n.namespace
	}

	return n.dataNamespace
}

// namespaceMigrationDAAPI is wrapper around the da json rpc to use when handling namespace migrations
type namespaceMigrationDAAPI struct {
	jsonrpc.API

	migrations map[uint64]namespaces

	currentNamespace     []byte
	currentDataNamespace []byte
}

func newNamespaceMigrationDAAPI(api jsonrpc.API, cfg config.Config, migrations map[uint64]namespaces) *namespaceMigrationDAAPI {
	return &namespaceMigrationDAAPI{
		API:                  api,
		migrations:           migrations,
		currentNamespace:     da.NamespaceFromString(cfg.DA.GetNamespace()).Bytes(),
		currentDataNamespace: da.NamespaceFromString(cfg.DA.GetDataNamespace()).Bytes(),
	}
}

// isDataNS returns true if the provided namespace matches the current data namespace
// or any historical data namespace defined in migrations.
func (api *namespaceMigrationDAAPI) isDataNS(ns []byte) bool {
	if bytes.Equal(ns, api.currentDataNamespace) {
		return true
	}
	for _, m := range api.migrations {
		if bytes.Equal(ns, da.NamespaceFromString(m.GetDataNamespace()).Bytes()) {
			return true
		}
	}
	return false
}

// orderedMigrationNamespaces returns the migration namespaces in a deterministic order (by height asc).
func (api *namespaceMigrationDAAPI) orderedMigrationNamespaces(isData bool) [][]byte {
	if len(api.migrations) == 0 {
		return nil
	}
	keys := make([]uint64, 0, len(api.migrations))
	for k := range api.migrations {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	out := make([][]byte, 0, len(keys))
	for _, k := range keys {
		m := api.migrations[k]
		if isData {
			out = append(out, da.NamespaceFromString(m.GetDataNamespace()).Bytes())
		} else {
			out = append(out, da.NamespaceFromString(m.GetNamespace()).Bytes())
		}
	}
	return out
}

// findNamespaceForHeight determines the correct namespace to use for a given height.
// Migrations are defined with "until" heights - the namespace is used until that height (inclusive).
// For example, a migration at untilHeight=100 means the namespace is used for heights 0-100.
func (api *namespaceMigrationDAAPI) findNamespaceForHeight(height uint64, isDataNamespace bool) []byte {
	if len(api.migrations) == 0 {
		if isDataNamespace {
			return api.currentDataNamespace
		}
		return api.currentNamespace
	}

	// Find the migration with the lowest untilHeight that is >= requested height
	var selectedUntilHeight uint64
	var found bool
	for untilHeight := range api.migrations {
		if untilHeight >= height && (!found || untilHeight < selectedUntilHeight) {
			selectedUntilHeight = untilHeight
			found = true
		}
	}

	// If no migration applies to this height, use current namespace
	if !found {
		if isDataNamespace {
			return api.currentDataNamespace
		}
		return api.currentNamespace
	}

	// Use the namespace from the migration
	migration := api.migrations[selectedUntilHeight]
	if isDataNamespace {
		return da.NamespaceFromString(migration.GetDataNamespace()).Bytes()
	}
	return da.NamespaceFromString(migration.GetNamespace()).Bytes()
}

// GetIDs returns IDs of all Blobs located in DA at given height.
// This method handles namespace migrations by determining the correct namespace based on height
func (api *namespaceMigrationDAAPI) GetIDs(ctx context.Context, height uint64, ns []byte) (*da.GetIDsResult, error) {
	isDataNamespace := api.isDataNS(ns)
	ns = api.findNamespaceForHeight(height, isDataNamespace)
	return api.API.GetIDs(ctx, height, ns)
}

// Get retrieves blobs by their IDs from the DA layer.
// This method tries the provided namespace first, then falls back to historical namespaces
// from migrations if the blob is not found.
func (api *namespaceMigrationDAAPI) Get(ctx context.Context, ids []da.ID, ns []byte) ([]da.Blob, error) {
	// Try with the provided namespace first
	blobs, err := api.API.Get(ctx, ids, ns)
	if err == nil {
		return blobs, nil
	}

	// If no migrations, return the original error
	if len(api.migrations) == 0 {
		return nil, err
	}

	// Determine if we're looking for data or header namespace
	isDataNamespace := api.isDataNS(ns)

	// Build deterministic fallback namespaces: current first, then migrations by height asc.
	candidates := make([][]byte, 0, len(api.migrations)+1)
	if isDataNamespace {
		candidates = append(candidates, api.currentDataNamespace)
	} else {
		candidates = append(candidates, api.currentNamespace)
	}
	candidates = append(candidates, api.orderedMigrationNamespaces(isDataNamespace)...)

	for _, candidate := range candidates {
		// Skip if this is the same as what we already tried
		if bytes.Equal(candidate, ns) {
			continue
		}
		blobs, err = api.API.Get(ctx, ids, candidate)
		if err == nil {
			return blobs, nil
		}
	}

	// Return the last error if nothing worked
	return nil, err
}

// GetProofs retrieves proofs for blobs by their IDs.
// This method tries the provided namespace first, then falls back to historical namespaces.
func (api *namespaceMigrationDAAPI) GetProofs(ctx context.Context, ids []da.ID, ns []byte) ([]da.Proof, error) {
	// Try with the provided namespace first
	proofs, err := api.API.GetProofs(ctx, ids, ns)
	if err == nil {
		return proofs, nil
	}

	// If no migrations, return the original error
	if len(api.migrations) == 0 {
		return nil, err
	}

	// Determine if we're looking for data or header namespace
	isDataNamespace := api.isDataNS(ns)

	// Build deterministic fallback namespaces: current first, then migrations by height asc.
	candidates := make([][]byte, 0, len(api.migrations)+1)
	if isDataNamespace {
		candidates = append(candidates, api.currentDataNamespace)
	} else {
		candidates = append(candidates, api.currentNamespace)
	}
	candidates = append(candidates, api.orderedMigrationNamespaces(isDataNamespace)...)

	for _, candidate := range candidates {
		// Skip if this is the same as what we already tried
		if bytes.Equal(candidate, ns) {
			continue
		}

		proofs, err = api.API.GetProofs(ctx, ids, candidate)
		if err == nil {
			return proofs, nil
		}
	}

	// Return the last error if nothing worked
	return nil, err
}

// Commit computes commitments for blobs.
// This method doesn't rewrite namespaces as it's a local operation.
func (api *namespaceMigrationDAAPI) Commit(ctx context.Context, blobs []da.Blob, ns []byte) ([]da.Commitment, error) {
	return api.API.Commit(ctx, blobs, ns)
}

// Validate validates blob proofs.
// This method tries the provided namespace first, then falls back to historical namespaces.
func (api *namespaceMigrationDAAPI) Validate(ctx context.Context, ids []da.ID, proofs []da.Proof, ns []byte) ([]bool, error) {
	// Try with the provided namespace first
	results, err := api.API.Validate(ctx, ids, proofs, ns)
	if err == nil {
		return results, nil
	}

	// If no migrations, return the original error
	if len(api.migrations) == 0 {
		return nil, err
	}

	// Determine if we're looking for data or header namespace
	isDataNamespace := api.isDataNS(ns)

	// Build deterministic fallback namespaces: current first, then migrations by height asc.
	candidates := make([][]byte, 0, len(api.migrations)+1)
	if isDataNamespace {
		candidates = append(candidates, api.currentDataNamespace)
	} else {
		candidates = append(candidates, api.currentNamespace)
	}
	candidates = append(candidates, api.orderedMigrationNamespaces(isDataNamespace)...)

	for _, candidate := range candidates {
		// Skip if this is the same as what we already tried
		if bytes.Equal(candidate, ns) {
			continue
		}

		results, err = api.API.Validate(ctx, ids, proofs, candidate)
		if err == nil {
			return results, nil
		}
	}

	// Return the last error if nothing worked
	return nil, err
}

// Submit submits blobs to the DA layer.
// This method uses the current namespace (no migration rewriting) as it's for new submissions.
func (api *namespaceMigrationDAAPI) Submit(ctx context.Context, blobs []da.Blob, gasPrice float64, ns []byte) ([]da.ID, error) {
	return api.API.Submit(ctx, blobs, gasPrice, ns)
}

// SubmitWithOptions submits blobs to the DA layer with additional options.
// This method uses the current namespace (no migration rewriting) as it's for new submissions.
func (api *namespaceMigrationDAAPI) SubmitWithOptions(ctx context.Context, blobs []da.Blob, gasPrice float64, ns []byte, options []byte) ([]da.ID, error) {
	return api.API.SubmitWithOptions(ctx, blobs, gasPrice, ns, options)
}
