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
	fetchMetadataCmd := flag.NewFlagSet("fetch-metadata", flag.ExitOnError)
	fmRpcUrl := fetchMetadataCmd.String("rpc-url", "", "RPC URL for connecting to network")
	fmContractAddress := fetchMetadataCmd.String("contract-address", "", "Contract address for collection")
	fmLowerTokenId := fetchMetadataCmd.Int64("lower-token-id", 0, "Lower token id bound")
	fmUpperTokenId := fetchMetadataCmd.Int64("upper-token-id", 10000, "Upper token id bound")

	cmds := map[string](*flag.FlagSet){
		fetchMetadataCmd.Name(): fetchMetadataCmd,
	}

	if len(os.Args) < len(cmds)+1 {
		fmt.Printf("Expected a command in the following list: %v\n", reflect.ValueOf(cmds).MapKeys())
		os.Exit(1)
	}

	if cmd, found := cmds[os.Args[1]]; found {
		cmd.Parse(os.Args[2:])

		switch os.Args[1] {
		case fetchMetadataCmd.Name():
			scripts.FetchMetadata(*fmRpcUrl, *fmContractAddress, big.NewInt(*fmLowerTokenId), big.NewInt(*fmUpperTokenId))
		}
	} else {
		fmt.Printf("Invalid command")
	}
}
