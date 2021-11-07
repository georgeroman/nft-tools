package scripts

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/georgeroman/nft-tools/internal/contracts"
)

func MonitorTokens(rpcWsUrl string) {
	client, err := ethclient.Dial(rpcWsUrl)
	if err != nil {
		log.Fatalf("Failed to connect to network: %s\n", err)
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to detect chain id: %s\n", err)
	}

	blockHeaders := make(chan *types.Header)
	subscription, err := client.SubscribeNewHead(context.Background(), blockHeaders)
	if err != nil {
		log.Fatalf("Failed to subscribe to new blocks: %s\n", err)
	}

	for {
		select {
		case err := <-subscription.Err():
			fmt.Printf("Subscription error: %s\n", err)
		case blockHeader := <-blockHeaders:
			fmt.Printf("Got new block %s\n", blockHeader.Number.String())

			block, err := client.BlockByHash(context.Background(), blockHeader.Hash())
			if err != nil {
				fmt.Printf("Failed to fetch block %s: %s\n", blockHeader.Number.String(), err)
				break
			}

			for _, tx := range block.Transactions() {
				if tx.To() == nil {
					msg, err := tx.AsMessage(types.NewLondonSigner(chainId), nil)
					if err != nil {
						fmt.Printf("Failed to handle transaction %s: %s\n", tx.Hash().Hex(), err)
						continue
					}

					contractAddress := crypto.CreateAddress(msg.From(), tx.Nonce())
					fmt.Printf("Detected newly deployed contract %s\n", contractAddress.Hex())

					contractAddress = common.HexToAddress("0x0d8f6F203662a73491bf65Cc4deFE039C843d098")

					erc165Abi, _ := abi.JSON(strings.NewReader(contracts.Erc165Abi))
					var erc721MetadataInterface [4]byte
					copy(erc721MetadataInterface[:], common.FromHex("0x5b5e139f"))
					calldata, _ := erc165Abi.Pack("supportsInterface", erc721MetadataInterface)
					encodedResult, err := client.CallContract(context.Background(), ethereum.CallMsg{To: &contractAddress, Data: calldata}, nil)
					if err == nil {
						result, err := erc165Abi.Unpack("supportsInterface", encodedResult)
						if err == nil && result[0].(bool) {
							fmt.Println("Contract is ERC721 compliant")

							erc721Abi, _ := abi.JSON(strings.NewReader(contracts.Erc721Abi))
							calldata, _ := erc721Abi.Pack("name")
							encodedResult, _ := client.CallContract(context.Background(), ethereum.CallMsg{To: &contractAddress, Data: calldata}, nil)
							result, _ := erc721Abi.Unpack("name", encodedResult)
							fmt.Println(result[0].(string))
						}
					}
					break
				}
			}
		}
	}
}
