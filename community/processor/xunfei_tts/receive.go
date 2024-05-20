package xunfei_tts

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
)

func (p *XunfeiTTSProcessor) receive() {
	for {
		var resp = RespData{}
		_, msg, err := p.conn.ReadMessage()
		if err != nil {
			fmt.Println("read message error:", err)
			break
		}
		json.Unmarshal(msg, &resp)
		if resp.Code != 0 {
			fmt.Println(resp.Code, resp.Message)
			return
		}

		audiobytes, err := base64.StdEncoding.DecodeString(resp.Data.Audio)
		if err != nil {
			panic(err)
		}

		log.Println("received audio length:", len(audiobytes))

		_, err = p.audioWriter.Write(audiobytes)
		if err != nil {
			fmt.Println("write to buffer error:", err)
			return
		}

		if resp.Data.Status == 2 {
			fmt.Println(resp.Code, resp.Message)
			break
		}
	}
	p.audioWriter.Close()
}

type RespData struct {
	Sid     string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	Audio  string `json:"audio,omitempty"`
	Ced    int    `json:"ced,omitempty"`
	Status int    `json:"status,omitempty"`
}
