package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
	"github.com/spf13/pflag"
	"go.coder.com/cli"
	"go.coder.com/flog"
)

var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

type recordCmd struct {
	outFile string
	port    int
}

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
	fl.IntVarP(&cmd.port, "port", "p", cmd.port, "Port to run the replay server on.")
	fl.StringVarP(&cmd.outFile, "out", "o", cmd.outFile, "Filename to write audio data")
}

// Run starts the screen-recording and stops the recording when the user inputs anything to stdin.
func (cmd *recordCmd) Run(fl *pflag.FlagSet) {
	if cmd.outFile == "" {
		flog.Error("you forgot to name the video")
		fl.Usage()
		return
	}

	if cmd.port < 1 || cmd.port > 65536 {
		flog.Error("%d is an invalid port number", cmd.port)
		fl.Usage()
		return
	}

	in := fmt.Sprintf("%s.avi", cmd.outFile)
	out, err := createRecording(in)
	if err != nil {
		flog.Error("failed to create recording : %v", err)
		fl.Usage()
		return
	}
	flog.Success("successfully created %s", out)

	errs := make(chan error, 1)
	port := fmt.Sprintf(":%d", cmd.port)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, signals...)

	flog.Info("starting replay server")
	server := startReplayServer(port, out, errs)
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			flog.Error("failed to shutdown replay server", "error", err.Error())
			return
		}
		flog.Success("successfully shutdown replay server")
	}()

	flog.Info("opening browser")
	if err := openbrowser(fmt.Sprintf("http://localhost%s", port)); err != nil {
		flog.Fatal(err.Error())
	}

	for {
		select {
		case err := <-errs:
			flog.Error(err.Error())
			fl.Usage()
			return
		case <-interrupt:
			println()
			flog.Info("stopping replay server")
			return
		}
	}
}

func startReplayServer(port, out string, errs chan error) *http.Server {
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "video/mp4")
		http.ServeFile(w, r, out)
	})

	server := &http.Server{Addr: port, Handler: r}

	go func() {
		errs <- server.ListenAndServe()
	}()

	return server
}

func openbrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}

	return err
}

func createRecording(fileName string) (string, error) {
	var out string
	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan bool)

	video, err := mjpeg.New(fileName, 200, 100, 2)
	if err != nil {
		return out, err
	}
	defer video.Close()

	flog.Success("recording started")

	go func() {
		flog.Info("press enter to stop recording")
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

	ext := path.Ext(fileName)
	out = fmt.Sprintf("%s.mp4", strings.TrimSuffix(fileName, ext))
	flog.Info("creating %s", out)

	convert := exec.Command("ffmpeg", "-i", fileName, out)
	if err := convert.Start(); err != nil {
		return out, err
	}
	if err := convert.Wait(); err != nil {
		return out, err
	}

	if err := os.Remove(fileName); err != nil {
		return out, err
	}
	return out, nil
}
