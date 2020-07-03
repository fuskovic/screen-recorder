# Screen-recorder(WIP)

An easy to use free screen-recorder.

Doesn't capture audio yet.

## Required

- [ffmpeg](https://ffmpeg.org/download.html)

## Installation

    go get -u github.com/fuskovic/screen-recorder

## Recording

    screen-recorder record --name cool_video_name --port 8000

The recording can be stopped by pressing enter in the same shell that started the program (reads from stdin).

Upon stopping the recording, a replay server will start in the background.

Your default browser should automatically open to the replay of the recording.