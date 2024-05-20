package xunfei_tts

import (
	"encoding/base64"
)

func (p *XunfeiTTSProcessor) send() {
	frameData := map[string]interface{}{
		"common": map[string]interface{}{
			"app_id": p.appCfg.appID, //appid 必须带上，只需第一帧发送
		},
		"business": map[string]interface{}{ //business 参数，只需一帧发送
			"vcn":   "xiaoyan",
			"aue":   "lame",
			"speed": 50,
			"tte":   "UTF8",
			"sfl":   1,
		},
		"data": map[string]interface{}{
			"status":   STATUS_LAST_FRAME,
			"encoding": "UTF8",
			"text":     base64.StdEncoding.EncodeToString([]byte(p.text)),
		},
	}

	p.conn.WriteJSON(frameData)
}
