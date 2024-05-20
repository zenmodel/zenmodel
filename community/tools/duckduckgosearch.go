package tools

import (
	"encoding/json"
	"fmt"
	"os"

	g "github.com/serpapi/google-search-results-golang"
)

func DuckDuckGoSearchToolCallDefinition() ToolCallDefinition {
	return ToolCallDefinition{
		Type: ToolTypeFunction,
		Function: &FunctionDefinition{
			Name:        "duckduckgo_search",
			Description: `A wrapper around DuckDuckGo Search.Useful for when you need to answer questions about current events.`,
			Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "engine": {
      "type": "string",
      "description": "Set parameter to duckduckgo to use the DuckDuckGo API engine.",
      "enum": ["duckduckgo"]
    },
    "q": {
      "type": "string",
      "description": "Parameter defines the query you want to search. You can use anything that you would use in a regular DuckDuckGo search. (e.g., inurl:, site:, intitle:, etc.)"
    },
    "kl": {
      "type": "string",
      "description": "Parameter defines the region to use for the DuckDuckGo search. Region code examples: us-en for the United States, uk-en for United Kingdom, or fr-fr for France. Head to the DuckDuckGo regions for a full list of supported regions.",
      "enum": [
        "xa-ar",
        "xa-en",
        "ar-es",
        "au-en",
        "at-de",
        "be-fr",
        "be-nl",
        "br-pt",
        "bg-bg",
        "ca-en",
        "ca-fr",
        "ct-ca",
        "cl-es",
        "cn-zh",
        "co-es",
        "hr-hr",
        "cz-cs",
        "dk-da",
        "ee-et",
        "fi-fi",
        "fr-fr",
        "de-de",
        "gr-el",
        "hk-tzh",
        "hu-hu",
        "in-en",
        "id-id",
        "id-en",
        "ie-en",
        "il-he",
        "it-it",
        "jp-jp",
        "kr-kr",
        "lv-lv",
        "lt-lt",
        "xl-es",
        "my-ms",
        "my-en",
        "mx-es",
        "nl-nl",
        "nz-en",
        "no-no",
        "pe-es",
        "ph-en",
        "ph-tl",
        "pl-pl",
        "pt-pt",
        "ro-ro",
        "ru-ru",
        "sg-en",
        "sk-sk",
        "sl-sl",
        "za-en",
        "es-es",
        "se-sv",
        "ch-de",
        "ch-fr",
        "ch-it",
        "tw-tzh",
        "th-th",
        "tr-tr",
        "ua-uk",
        "uk-en",
        "us-en",
        "ue-es",
        "ve-es",
        "vn-vi",
        "wt-wt"
        ]
    }
  },
  "required": ["engine", "q", "api_key"]
}`),
		},
		CallFunc: ddgSearch,
	}
}

func ddgSearch(args string) (resp string, err error) {
	// Define a struct to hold the arguments provided in JSON format.
	type searchArgs struct {
		Engine string `json:"engine"`
		Query  string `json:"q"`
		Region string `json:"kl"`
	}

	var arguments searchArgs

	// Unmarshal the JSON string into the struct.
	err = json.Unmarshal([]byte(args), &arguments)
	if err != nil {
		return "", fmt.Errorf("error parsing args: %v", err)
	}

	// Here you replace with your actual SERP API key.
	apiKey := os.Getenv("SERP_API_KEY")

	parameter := map[string]string{
		"engine":  arguments.Engine,
		"q":       arguments.Query,
		"kl":      arguments.Region,
		"api_key": apiKey,
	}
	search := g.NewGoogleSearch(parameter, apiKey)
	results, err := search.GetJSON()

	cutResults := results["organic_results"].([]interface{})
	cutResults = cutResults[:10]
	if err != nil {
		return "", err
	}
	res, err := json.Marshal(cutResults)
	if err != nil {
		return "", err
	}

	return string(res), nil
}
