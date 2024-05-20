package xunfei_stt

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/zenmodel/zenmodel"
	"golang.org/x/net/websocket"
)

func NewProcessor() *XunfeiSTTProcessor {
	return &XunfeiSTTProcessor{
		MemoryKeyTextQueue:     "text_queue",
		MemoryKeyTextQueueSize: 10,
		appCfg: appConfig{
			appID:  os.Getenv("XUNFEI_APP_ID"),
			appKey: os.Getenv("XUNFEI_APP_KEY"),
		},
	}
}

type XunfeiSTTProcessor struct { // nolint
	MemoryKeyTextQueue     string
	MemoryKeyTextQueueSize int

	appCfg appConfig
	conn   *websocket.Conn
}

type appConfig struct {
	appID  string
	appKey string
}

func (p *XunfeiSTTProcessor) Process(brain zenmodel.BrainRuntime) error {
	if err := p.wsConnect(); err != nil {
		return fmt.Errorf("ws connect error:%v\n", err)
	}
	defer p.conn.Close()
	if err := wsHandShake(p.conn); err != nil {
		return fmt.Errorf("ws handshake error:%v\n", err)
	}

	if err := p.ensureTextQueue(brain); err != nil {
		return err
	}

	sendChan := make(chan int, 1)
	readChan := make(chan int, 1)
	defer close(sendChan)
	defer close(readChan)
	go p.send(sendChan)
	go p.receive(readChan, brain)
	<-sendChan
	<-readChan

	return nil
}

func (p *XunfeiSTTProcessor) DeepCopy() zenmodel.Processor {
	return &XunfeiSTTProcessor{
		MemoryKeyTextQueue: p.MemoryKeyTextQueue,
		appCfg: appConfig{
			appID:  p.appCfg.appID,
			appKey: p.appCfg.appKey,
		},
	}
}

func (p *XunfeiSTTProcessor) WithMemoryKeyTextQueue(memKeyTextQueue string) zenmodel.Processor {
	p.MemoryKeyTextQueue = memKeyTextQueue
	return p
}

func (p *XunfeiSTTProcessor) WithAppConfig(appID, appKey string) zenmodel.Processor {
	p.appCfg.appID = appID
	p.appCfg.appKey = appKey
	return p
}

func (p *XunfeiSTTProcessor) wsConnect() error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha1.New, []byte(p.appCfg.appKey))
	strByte := []byte(p.appCfg.appID + ts)
	strMd5Byte := md5.Sum(strByte)
	strMd5 := fmt.Sprintf("%x", strMd5Byte)
	mac.Write([]byte(strMd5))
	signa := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	requestParam := "appid=" + p.appCfg.appID + "&ts=" + ts + "&signa=" + signa

	conn, err := websocket.Dial("ws://"+XunfeiSTTHost+"?"+requestParam, websocket.SupportedProtocolVersion, "http://"+XunfeiSTTHost)
	if err != nil {
		return err
	}
	p.conn = conn

	return nil
}

func wsHandShake(conn *websocket.Conn) error {
	var message string
	websocket.Message.Receive(conn, &message)
	var m map[string]string
	err := json.Unmarshal([]byte(message), &m)
	println(message)
	if err != nil {
		return err
	} else if m["code"] != "0" {
		return fmt.Errorf("handshake fail!" + message)
	}

	return nil
}

func (p *XunfeiSTTProcessor) ensureTextQueue(brain zenmodel.BrainRuntime) error {
	if !brain.ExistMemory(p.MemoryKeyTextQueue) {
		q := make(chan string, p.MemoryKeyTextQueueSize)
		if err := brain.SetMemory(p.MemoryKeyTextQueue, q); err != nil {
			return err
		}
	}

	return nil
}
