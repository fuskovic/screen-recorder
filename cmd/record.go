package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"image/jpeg"
	"os"
	"os/exec"
	"time"

	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
	"github.com/spf13/pflag"
	"go.coder.com/cli"
	"go.coder.com/flog"
)

type recordCmd struct{ name string }

// Spec returns a command spec containing a description of it's usage.
func (cmd *recordCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "record",
		Usage: "[flags]",
		Desc:  "Start a screen-recording.",
	}
}

// RegisterFlags initializes how a flag set is processed for a particular command.
func (cmd *recordCmd) RegisterFlags(fl *pflag.FlagSet) {
	fl.StringVarP(&cmd.name, "name", "n", cmd.name, "Name to give screen-recording.")
}

// Run starts the screen-recording and stops the recording when the user inputs anything to stdin.
func (cmd *recordCmd) Run(fl *pflag.FlagSet) {
	if cmd.name == "" {
		flog.Error("you forgot to name the video")
		fl.Usage()
		return
	}

	in := fmt.Sprintf("%s.avi", cmd.name)

	video, err := mjpeg.New(in, 200, 100, 2)
	if err != nil {
		flog.Fatal("failed to create video.avi", "error", err)
	}
	defer video.Close()

	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan bool)
	flog.Success("recording started")
	flog.Info("press enter to stop recording")

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
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

	for scanner.Scan() {
		flog.Info("stopped recording")
		done <- true
		break
	}

	out := fmt.Sprintf("%s.mp4", cmd.name)
	flog.Info("creating %s", out)

	convert := exec.Command("ffmpeg", "-i", in, out)
	if err := convert.Start(); err != nil {
		flog.Fatal(err.Error())
	}
	if err := convert.Wait(); err != nil {
		flog.Fatal(err.Error())
	}

	if err := os.Remove(in); err != nil {
		flog.Fatal(err.Error())
	}

	flog.Success("%s successfully created", out)
}
