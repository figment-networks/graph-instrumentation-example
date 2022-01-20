package main

import (
	"github.com/spf13/cobra"
	"github.com/streamingfast/dlauncher/launcher"
)

var (
	// GRPC Service Addresses
	BlockStreamServingAddr  = ":9000"
	RelayerServingAddr      = ":9010"
	MergerServingAddr       = ":9020"
	FirehoseGRPCServingAddr = ":9030"

	// Blocks store
	MergedBlocksStoreURL string = "file://{sf-data-dir}/storage/merged-blocks"
	OneBlockStoreURL     string = "file://{sf-data-dir}/storage/one-blocks"

	// Protocol defaults
	FirstStreamableBlock = 0
	GenesisBlock         = 0
)

func init() {
	launcher.RegisterCommonFlags = func(cmd *cobra.Command) error {
		//Common stores configuration flags
		cmd.Flags().String("common-blocks-store-url", MergedBlocksStoreURL, "Store URL (with prefix) where to read/write")
		cmd.Flags().String("common-oneblock-store-url", OneBlockStoreURL, "Store URL (with prefix) to read/write one-block files")
		cmd.Flags().String("common-blockstream-addr", RelayerServingAddr, "GRPC endpoint to get real-time blocks")
		cmd.Flags().Int("common-first-streamable-block", FirstStreamableBlock, "First streamable block number")
		cmd.Flags().Int("common-genesis-block", GenesisBlock, "Genesis block number")

		// Authentication, metering and rate limiter plugins
		cmd.Flags().String("common-auth-plugin", "null://", "Auth plugin URI, see streamingfast/dauth repository")
		cmd.Flags().String("common-metering-plugin", "null://", "Metering plugin URI, see streamingfast/dmetering repository")

		// System Behavior
		cmd.Flags().Duration("common-shutdown-delay", 5, "Add a delay between receiving SIGTERM signal and shutting down apps. Apps will respond negatively to /healthz during this period")

		return nil
	}
}
