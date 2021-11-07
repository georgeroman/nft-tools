package scripts

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/georgeroman/nft-tools/internal/contracts"
)

type Call struct {
	Target   common.Address
	CallData []byte
}

type MetadataRequest struct {
	TokenId  string
	TokenUri string
}

type MetadataResponse struct {
	TokenId  string
	Metadata string
}

func FetchMetadata(rpcUrl string, contractAddress string, lowerTokenId *big.Int, upperTokenId *big.Int) {
	// Connect to the network
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatalf("Failed to connect to network: %s", err)
	}

	// Select multicall address
	var multicallAddress common.Address
	chainId, _ := client.ChainID(context.Background())
	switch chainId.Uint64() {
	case 1:
		multicallAddress = common.HexToAddress("0xeefba1e63905ef1d7acba5a8513c70307c1ce441")
	default:
		log.Fatalf("Unsupported chain id")
	}

	// Make sure the metadata folder exists
	os.MkdirAll(fmt.Sprintf("./metadata/%s", contractAddress), os.ModePerm)

	var reqWg sync.WaitGroup
	var resWg sync.WaitGroup

	// Handle metadata requests concurrently
	reqChan := make(chan MetadataRequest)
	resChan := make(chan MetadataResponse)
	for i := 0; i < 10; i++ {
		reqWg.Add(1)
		go func() {
			defer reqWg.Done()

			for req := range reqChan {
				if _, err := os.Stat(fmt.Sprintf("./metadata/%s/%s.json", contractAddress, req.TokenId)); errors.Is(err, os.ErrNotExist) {
					fmt.Println("req", req.TokenId)

					numRetries := 5
					for numRetries > 0 {
						if strings.HasPrefix(req.TokenUri, "ipfs://") {
							req.TokenUri = fmt.Sprintf("https://gateway.ipfs.io/ipfs/%s", req.TokenUri[7:])
						}

						if strings.HasPrefix(req.TokenUri, "https://") {
							response, err := http.Get(req.TokenUri)
							if err != nil {
								fmt.Printf("Failed to fetch metadata for token id %s, retrying...\n", req.TokenId)
								numRetries--
								time.Sleep(time.Duration((5 - numRetries) * 2 * int(time.Second)))
								continue
							}

							metadata, _ := ioutil.ReadAll(response.Body)
							resChan <- MetadataResponse{
								TokenId:  req.TokenId,
								Metadata: string(metadata),
							}
							break
						}
					}

					if numRetries == 0 {
						fmt.Printf("Could not fetch metadata for token id %s", req.TokenId)
					}
				}
			}
		}()

		resWg.Add(1)
		go func() {
			defer resWg.Done()

			for res := range resChan {
				fmt.Println("res", res.TokenId)
				os.WriteFile(fmt.Sprintf("./metadata/%s/%s.json", contractAddress, res.TokenId), []byte(res.Metadata), 0644)
			}
		}()
	}

	// For efficiency, process the tokens in batches instead of serially
	var processWg sync.WaitGroup
	var BATCH_SIZE = big.NewInt(50)
	for id := lowerTokenId; id.Cmp(upperTokenId) < 1; id.Add(id, BATCH_SIZE) {
		processWg.Add(1)

		var startId big.Int
		startId.Add(id, big.NewInt(0))

		var endId big.Int
		endId.Add(id, BATCH_SIZE)
		if endId.Cmp(upperTokenId) == 1 {
			endId = *upperTokenId
		}

		go func() {
			defer processWg.Done()

			var tokenIds []string
			var calls []Call
			erc721Abi, _ := abi.JSON(strings.NewReader(contracts.Erc721Abi))
			for id := startId; id.Cmp(&endId) < 1; id.Add(&id, big.NewInt(1)) {
				tokenIds = append(tokenIds, id.String())
				calldata, _ := erc721Abi.Pack("tokenURI", &id)
				calls = append(calls, Call{
					Target:   common.HexToAddress(contractAddress),
					CallData: calldata,
				})
			}

			multicallAbi, _ := abi.JSON(strings.NewReader(contracts.MulticallAbi))
			encodedCalldata, _ := multicallAbi.Pack("aggregate", calls)
			encodedResult, _ := client.CallContract(context.Background(), ethereum.CallMsg{To: &multicallAddress, Data: encodedCalldata}, nil)
			decodedResult, _ := multicallAbi.Unpack("aggregate", encodedResult)

			uris := decodedResult[1].([][]byte)
			for i := 0; i < len(tokenIds); i++ {
				result, _ := erc721Abi.Unpack("tokenURI", uris[i])
				uri := result[0].(string)
				reqChan <- MetadataRequest{
					TokenId:  tokenIds[i],
					TokenUri: uri,
				}
			}
		}()
	}

	processWg.Wait()

	close(reqChan)
	reqWg.Wait()

	close(resChan)
	resWg.Wait()
}
