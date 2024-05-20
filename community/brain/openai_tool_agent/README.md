# OpenAI Agent With Tools

### Usage

1. write main.go
```go
import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel/community/brain/openai_tool_agent"
	"github.com/zenmodel/zenmodel/community/tools"
)

func main() {
    // clone community shared brainprint, and set some tool cal definitions(support multi definitions)
	bp := openai_tool_agent.CloneBrainprint(tools.OpenWeatherToolCallDefinition())
	// build brain
	brain := bp.Build()
	// set memory and trig all entry links
	_ = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})

	// block process util brain sleeping
	brain.Wait()

    // get messages finally 
	messages, _ := json.Marshal(brain.GetMemory("messages"))
	fmt.Printf("messages: %s\n", messages)
}
```
2. run it !

```shell
export OPENAI_API_KEY=chage_it
export OPEN_WEATHER_API_KEY=chage_it
go run main.go
```