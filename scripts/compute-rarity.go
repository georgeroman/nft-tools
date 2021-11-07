package scripts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Metadata struct {
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

func ComputeRarity(contractAddress string) {
	metadataDir := fmt.Sprintf("./metadata/%s", contractAddress)

	_, err := os.Stat(metadataDir)
	if err != nil {
		log.Fatalf("No metadata found\n")
	}

	metadataFiles, err := ioutil.ReadDir(metadataDir)
	if err != nil {
		log.Fatalf("Failed to read metadata\n")
	}

	tokenAttributes := make(map[string][]Attribute)
	valueCounts := make(map[string]uint)
	tokenCount := 0

	for _, metadataFile := range metadataFiles {
		rawMetadata, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", metadataDir, metadataFile.Name()))
		if err != nil {
			fmt.Printf("Failed to read metadata file %s\n", metadataFile.Name())
			continue
		}

		var metadata Metadata
		err = json.Unmarshal(rawMetadata, &metadata)
		if err != nil {
			fmt.Printf("Failed to parse metadata file %s: %s\n", metadataFile.Name(), err)
			continue
		}

		tokenId := strings.TrimSuffix(metadataFile.Name(), filepath.Ext(metadataFile.Name()))
		tokenAttributes[tokenId] = metadata.Attributes
		for _, attribute := range metadata.Attributes {
			valueCounts[attribute.Value]++
		}
		tokenCount++
	}

	var maxValueCount uint
	for value := range valueCounts {
		if valueCounts[value] > maxValueCount {
			maxValueCount = valueCounts[value]
		}
	}

	tokenRarities := make(map[string]float64)
	tokens := make([]string, 0, len(tokenRarities))
	for tokenId := range tokenAttributes {
		rarity := float64(0)
		for _, attribute := range tokenAttributes[tokenId] {
			rarity += 1 / (float64(valueCounts[attribute.Value]) / float64(tokenCount)) / (float64(valueCounts[attribute.Value]) / float64(maxValueCount))
		}
		tokenRarities[tokenId] = rarity
		tokens = append(tokens, tokenId)
	}

	sort.Slice(tokens, func(i, j int) bool { return tokenRarities[tokens[i]] <= tokenRarities[tokens[j]] })
	for _, tokenId := range tokens {
		fmt.Println(tokenId, tokenRarities[tokenId])
	}
}
