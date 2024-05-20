package tools

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func OpenWeatherToolCallDefinition() ToolCallDefinition {
	return ToolCallDefinition{
		Type: ToolTypeFunction,
		Function: &FunctionDefinition{
			Name:        "get_current_weather",
			Description: "Get the current weather in a given location",
			Parameters: json.RawMessage(`{
	"type": "object",
	"properties": {
		"location": {
			"type": "string",
			"description": "The city name, e.g. San Francisco",
			"properties": {}
		},
		"unit": {
			"type": "string",
			"description": "Unit standard: Kelvin, metric: Celsius, imperial: Fahrenheit",
			"enum": ["standard", "metric", "imperial"],
			"properties": {}
		}
	},
	"required": ["location"]
}`),
		},
		CallFunc: getCurrentWeather,
	}
}

func getCurrentWeather(args string) (resp string, err error) {
	// Define a struct to hold the arguments provided in JSON format.
	type weatherArgs struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}

	var arguments weatherArgs

	// Unmarshal the JSON string into the struct.
	err = json.Unmarshal([]byte(args), &arguments)
	if err != nil {
		return "", fmt.Errorf("error parsing args: %v", err)
	}

	// Here you replace with your actual OpenWeather API key.
	apiKey := os.Getenv("OPEN_WEATHER_API_KEY")

	// Construct the query with proper URL encoding.
	baseURL := "https://api.openweathermap.org/data/2.5/weather"
	query := fmt.Sprintf("?appid=%s&q=%s&units=%s", url.QueryEscape(apiKey), url.QueryEscape(arguments.Location), url.QueryEscape(arguments.Unit))

	// Build the full request URL.
	requestURL := baseURL + query

	// Instantiate the HTTP client and make the request.
	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", err
	}

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Check if the status code is 200 OK.
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request to OpenWeather API did not return a successful status code")
	}

	// Read the response body.
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Convert the body to a string and return it.
	resp = string(bodyBytes)
	return resp, nil
}
