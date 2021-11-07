package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"reflect"

	"github.com/georgeroman/nft-tools/scripts"
)

func main() {
	// Compute rarity
	computeRarityCmd := flag.NewFlagSet("compute-rarity", flag.ExitOnError)
	crContractAddress := computeRarityCmd.String("contract-address", "", "Contract address for collection")

	// Fetch metadata
	fetchMetadataCmd := flag.NewFlagSet("fetch-metadata", flag.ExitOnError)
	fmRpcHttpUrl := fetchMetadataCmd.String("rpc-http-url", "", "RPC HTTP URL for connecting to network")
	fmContractAddress := fetchMetadataCmd.String("contract-address", "", "Contract address for collection")
	fmLowerTokenId := fetchMetadataCmd.Int64("lower-token-id", 0, "Lower token id bound")
	fmUpperTokenId := fetchMetadataCmd.Int64("upper-token-id", 10000, "Upper token id bound")

	// Monitor tokens
	monitorTokensCmd := flag.NewFlagSet("monitor-tokens", flag.ExitOnError)
	mtRpcWsUrl := monitorTokensCmd.String("rpc-ws-url", "", "RPC WS URL for connecting to network")

	cmds := map[string](*flag.FlagSet){
		computeRarityCmd.Name(): computeRarityCmd,
		fetchMetadataCmd.Name(): fetchMetadataCmd,
		monitorTokensCmd.Name(): monitorTokensCmd,
	}

	if len(os.Args) < len(cmds)+1 {
		fmt.Printf("Expected a command in the following list: %v\n", reflect.ValueOf(cmds).MapKeys())
		os.Exit(1)
	}

	if cmd, found := cmds[os.Args[1]]; found {
		cmd.Parse(os.Args[2:])

		switch os.Args[1] {
		case computeRarityCmd.Name():
			scripts.ComputeRarity(*crContractAddress)
		case fetchMetadataCmd.Name():
			scripts.FetchMetadata(*fmRpcHttpUrl, *fmContractAddress, big.NewInt(*fmLowerTokenId), big.NewInt(*fmUpperTokenId))
		case monitorTokensCmd.Name():
			scripts.MonitorTokens(*mtRpcWsUrl)
		}
	} else {
		fmt.Printf("Invalid command\n")
	}
}
