# Screen-recorder(WIP)

Screen recorder for Mac OSX.

`Doesn't capture audio yet.`


## Installation

    brew update

    brew install ffmpeg

    go get -u github.com/fuskovic/screen-recorder

## Recording

    screen-recorder record --out cool_video_name --port 8000

The recording can be stopped by pressing enter in the same shell that started the program (reads from stdin).

Upon stopping the recording, a replay server will start in the background.

Your default browser should automatically open to the replay of the recording.