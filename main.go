package main

import (
	"bytes"
	"image/jpeg"
	"time"

	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
	"go.coder.com/flog"
)

func main() {
	video, err := mjpeg.New("video.avi", 200, 100, 2)
	if err != nil {
		flog.Fatal("failed to create video.avi", "error", err)
	}
	defer video.Close()

	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				flog.Info("tick")
				buf := &bytes.Buffer{}
				bounds := screenshot.GetDisplayBounds(0)

				img, err := screenshot.CaptureRect(bounds)
				if err != nil {
					flog.Fatal("failed to capture screenshot", "error", err)
				}

				if err := jpeg.Encode(buf, img, nil); err != nil {
					flog.Fatal("failed to encode jpeg", "error", err)
				}

				if err := video.AddFrame(buf.Bytes()); err != nil {
					flog.Fatal("failed to add frame", "error", err)
				}
			}
		}
	}()

	time.Sleep(10 * time.Second)
	ticker.Stop()
	done <- true
}
