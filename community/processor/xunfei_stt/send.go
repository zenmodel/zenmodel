package xunfei_stt

/*
// NOTE: Only works on SDL2 2.0.5 and above!

extern void cOnAudio(void *userdata, unsigned char *stream, int len);
*/
import "C"
import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/net/websocket"
)

var globalWsConn *websocket.Conn

func (p *XunfeiSTTProcessor) send(sendChan chan int) {
	globalWsConn = p.conn
	// 分片上传音频
	defer func() {
		sendChan <- 1
	}()

	// 1. mic
	sdl.Init(sdl.INIT_AUDIO)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

	func() {
		dev, err := openAudioDevice() // this will block process
		if err != nil {
			panic(err)
		}
		defer sdl.CloseAudioDevice(dev)

		<-sigchan

		fmt.Println("Exiting..")
	}()

	sdl.Quit()

	// 上传结束符
	if err := websocket.Message.Send(p.conn, EndTag); err != nil {
		println("send string msg err: ", err)
	} else {
		println("send end tag success, ", len(EndTag))
	}

}

func openAudioDevice() (dev sdl.AudioDeviceID, err error) {
	var want, have sdl.AudioSpec

	want.Callback = sdl.AudioCallback(C.cOnAudio)
	want.Channels = 1
	want.Format = sdl.AUDIO_S16LSB
	want.Freq = 16000
	want.Samples = SendSize / 2

	dev, err = sdl.OpenAudioDevice("", true, &want, &have, 0)
	if err != nil {
		return
	}

	sdl.PauseAudioDevice(dev, false)

	return
}

//export onAudio
func onAudio(raw *C.uchar, sz int) {
	data := makeBytes(raw, sz)
	//fmt.Println("Received audio:", len(data), "bytes")

	if err := websocket.Message.Send(globalWsConn, data); err != nil {
		println("send byte msg err: ", err)
		return
	}
	// println("send data success, sleep 40 ms")
	time.Sleep(40 * time.Millisecond)
}
func makeBytes(raw *C.uchar, len int) (out []byte) {
	in := asBytes(raw, len)
	out = make([]byte, len)

	for i := 0; i < len; i++ {
		out[i] = in[i]
	}

	return
}
func asBytes(in *C.uchar, len int) (p []byte) {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(unsafe.Pointer(in))
	return
}
