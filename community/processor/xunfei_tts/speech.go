package xunfei_tts

import (
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/hajimehoshi/oto"
	"github.com/tosone/minimp3"
)

func (p *XunfeiTTSProcessor) speech() {
	log.Printf("-----------> Start audio play\n")

	var err error
	var dec *minimp3.Decoder
	if dec, err = minimp3.NewDecoder(p.audioReader); err != nil {
		log.Fatal(err)
	}
	<-dec.Started()

	log.Printf("Convert audio sample rate: %d, channels: %d\n", dec.SampleRate, dec.Channels)

	var context *oto.Context
	if context, err = oto.NewContext(dec.SampleRate, dec.Channels, 2, 4096); err != nil {
		log.Fatal(err)
	}

	var waitForPlayOver = new(sync.WaitGroup)
	waitForPlayOver.Add(1)

	var player = context.NewPlayer()

	go func() {
		for {
			var data = make([]byte, 512)
			_, err = dec.Read(data)
			if err == io.EOF {
				fmt.Printf("----------------->End audio play\n")
				break
			}
			if err != nil {
				log.Fatal(err)

				break
			}
			player.Write(data)
		}
		log.Println("over play.")
		waitForPlayOver.Done()
	}()

	waitForPlayOver.Wait()

	<-time.After(time.Second)
	dec.Close()
	player.Close()
}
