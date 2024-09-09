package xunfei_stt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zenmodel/zenmodel/processor"
	"golang.org/x/net/websocket"
)

func (p *XunfeiSTTProcessor) receive(readChan chan int, brain processor.BrainContext) {
	for {
		var msg []byte
		var result map[string]string
		if err := websocket.Message.Receive(p.conn, &msg); err != nil {
			if err.Error() == "EOF" {
				println("receive date end")
			} else {
				fmt.Printf("receive msg error: %v\n", err)
			}

			break
		}

		err := json.Unmarshal(msg, &result)
		if err != nil {
			println(string(msg))
			fmt.Printf("response json parse error\n")
			continue
		}

		if result["code"] == "0" {
			var asrResult AsrResult
			err := json.Unmarshal([]byte(result["data"]), &asrResult)
			if err != nil {
				fmt.Printf("parse asrResult error: %v\n", err)
				println("receive msg: ", string(msg))

				break
			}
			if asrResult.Cn.St.Type == "0" {
				println("------------------------------------------------------------------------------------------------------------------------------------")
				// 最终结果
				var text strings.Builder
				for _, wse := range asrResult.Cn.St.Rt[0].Ws {
					for _, cwe := range wse.Cw {
						print(cwe.W)
						text.WriteString(cwe.W)
					}
				}
				textCh, ok := brain.GetMemory(p.MemoryKeyTextQueue).(chan string)
				if ok {
					textCh <- text.String()
				}
				println("\r\n------------------------------------------------------------------------------------------------------------------------------------")
			} else {
				for _, wse := range asrResult.Cn.St.Rt[0].Ws {
					for _, cwe := range wse.Cw {
						print(cwe.W)
					}
				}
				println()
			}
		} else {
			println("invalid result: ", string(msg))
		}
	}
	readChan <- 1
}

type AsrResult struct {
	Cn    Cn      `json:"cn"`
	SegId float64 `json:"seg_id"`
}

type Cn struct {
	St St `json:"st"`
}

type St struct {
	Bg   string      `json:"bg"`
	Ed   string      `json:"ed"`
	Type string      `json:"type"`
	Rt   []RtElement `json:"rt"`
}

type RtElement struct {
	Ws []WsElement `json:"ws"`
}

type WsElement struct {
	Wb float64     `json:"wb"`
	We float64     `json:"we"`
	Cw []CwElement `json:"cw"`
}

type CwElement struct {
	W  string `json:"w"`
	Wp string `json:"wp"`
}
