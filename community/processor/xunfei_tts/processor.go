package xunfei_tts

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zenmodel/zenmodel/processor"
)

func NewProcessor() *XunfeiTTSProcessor {
	r, w := io.Pipe()
	return &XunfeiTTSProcessor{
		MemoryKeyTextToSpeech: "text_to_speech",
		appCfg: appConfig{
			appID:     os.Getenv("XUNFEI_APP_ID"),
			apiKey:    os.Getenv("XUNFEI_API_KEY"),
			apiSecret: os.Getenv("XUNFEI_API_SECRET"),
		},
		audioReader: r,
		audioWriter: w,
	}
}

type XunfeiTTSProcessor struct { // nolint
	MemoryKeyTextToSpeech string

	appCfg      appConfig
	conn        *websocket.Conn
	text        string
	audioReader io.ReadCloser
	audioWriter io.WriteCloser
}

type appConfig struct {
	appID     string
	apiKey    string
	apiSecret string
}

func (p *XunfeiTTSProcessor) Process(brain processor.BrainContext) error {
	if !brain.ExistMemory(p.MemoryKeyTextToSpeech) {
		// TODO log no need to process
		return nil
	} else {
		text, ok := brain.GetMemory(p.MemoryKeyTextToSpeech).(string)
		if !ok {
			return fmt.Errorf("memory %s is not string", p.MemoryKeyTextToSpeech)
		}
		if text == "" {
			// TODO log no need to process
			return nil
		}
		p.text = text
	}

	if err := p.wsConn(); err != nil {
		return fmt.Errorf("failed to connect to xunfei websocket: %w", err)
	}
	defer p.conn.Close()

	go p.send()
	go p.receive()

	p.speech()

	return nil
}

func (p *XunfeiTTSProcessor) Clone() processor.Processor {
	return &XunfeiTTSProcessor{
		MemoryKeyTextToSpeech: p.MemoryKeyTextToSpeech,
		appCfg: appConfig{
			appID:     p.appCfg.appID,
			apiKey:    p.appCfg.apiKey,
			apiSecret: p.appCfg.apiSecret,
		},
	}
}

func (p *XunfeiTTSProcessor) WithMemoryKeyTextToSpeech(memoryKeyTextToSpeech string) processor.Processor {
	p.MemoryKeyTextToSpeech = memoryKeyTextToSpeech
	return p
}

func (p *XunfeiTTSProcessor) WithAppConfig(appID, apiKey, apiSecret string) processor.Processor {
	p.appCfg.appID = appID
	p.appCfg.apiKey = apiKey
	p.appCfg.apiSecret = apiSecret
	return p
}

func (p *XunfeiTTSProcessor) wsConn() error {
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := d.Dial(assembleAuthUrl(XunfeiTTSURL, p.appCfg.apiKey, p.appCfg.apiSecret), nil)
	if err != nil {
		return fmt.Errorf("response: %s, error: %v", readResp(resp), err)
	} else if resp.StatusCode != 101 {
		return fmt.Errorf("response: %s, error: %v", readResp(resp), err)
	}
	p.conn = conn

	return nil
}

// 创建鉴权url  apikey 即 hmac username
func assembleAuthUrl(hosturl string, apiKey, apiSecret string) string {
	ul, err := url.Parse(hosturl)
	if err != nil {
		fmt.Println(err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	fmt.Println(sgin)
	//签名结果
	sha := HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	fmt.Println(sha)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	callurl := hosturl + "?" + v.Encode()
	return callurl
}

func HmacWithShaTobase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

func readResp(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("code=%d,body=%s", resp.StatusCode, string(b))
}
