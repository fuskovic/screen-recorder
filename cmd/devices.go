package cmd

import (
	"bufio"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"go.coder.com/cli"
	"go.coder.com/flog"
)

type (
	devicesCmd struct{}
	device     struct {
		name, kind string
		index      int
	}
)

// Spec returns a command spec containing a description of it's usage.
func (cmd *devicesCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "devices",
		Usage: "[flags]",
		Desc:  "List connected video and audio devices.",
	}
}

// RegisterFlags initializes how a flag set is processed for a particular command.
func (cmd *devicesCmd) RegisterFlags(fl *pflag.FlagSet) {}

// Run lists video and audio devices.
func (cmd *devicesCmd) Run(fl *pflag.FlagSet) {
	var list *exec.Cmd
	var devices []device
	sys := runtime.GOOS

	switch sys {
	case "darwin":
		list = exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "\"\"")
		out, _ := list.CombinedOutput()
		devices = getDarwinDevices(out)
	case "linux":
		// TODO: add linux support
	case "windows":
		// TODO: add windows support
	default:
		flog.Error("OPERATING SYSTEM %s IS NOT SUPPORTED", sys)
		fl.Usage()
		return
	}

	sep := func(n int) string { return strings.Repeat(" ", n) }
	flog.Success("FOUND %d DEVICES", len(devices))
	flog.Info("KIND%sINDEX%sNAME", sep(8), sep(8))

	for _, d := range devices {
		flog.Info("%s%s%d%s%s", d.kind, sep(7), d.index, sep(11), d.name)
	}
}

func getDarwinDevices(out []byte) []device {
	var devices []device
	var isRangingVideoDevices, isRangingAudioDevices bool

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	for scanner.Scan() {
		var kind string
		line := scanner.Text()

		if !strings.Contains(line, "AVFoundation input device") {
			continue
		}

		if isRangingVideoDevices {
			kind = "video"
		} else if isRangingAudioDevices {
			kind = "audio"
		} else {
			kind = ""
		}

		if kind != "" && strings.Contains(line, "AVFoundation") && !strings.Contains(line, ":") {
			name := line[45:][3:]
			indexInBrkts := line[45:][0:3]
			indexStr := strings.Replace(indexInBrkts, "[", "", 1)
			indexStr = strings.Replace(indexStr, "]", "", 1)
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				flog.Error(err.Error())
				return devices
			}

			devices = append(devices, device{
				name:  name,
				index: index,
				kind:  kind,
			})
		}

		if strings.Contains(line, "AVFoundation video devices:") {
			isRangingVideoDevices = true
			continue
		}

		if strings.Contains(line, "AVFoundation audio devices:") {
			isRangingVideoDevices = false
			isRangingAudioDevices = true
		}
	}
	return devices
}
