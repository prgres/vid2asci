package app

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

func resizeVideo(inputPath string, outputPath string) ([]byte, error) {
	return execCmd("ffmpeg",
		"-i", inputPath, // input
		"-vf", fmt.Sprintf("scale=%d:%d", _xw, _xh),
		outputPath, // output
	)
}

func cropBlackBars(inputPath string, outputPath string, blackBarsParam string) ([]byte, error) {
	return execCmd("ffmpeg",
		"-i", inputPath, // input
		"-vf", fmt.Sprintf("crop=%s", blackBarsParam),
		outputPath, // output
	)
}

func getCropBlackBarsParams(inputPath string) ([]byte, error) {
	return execCmd("bash", "-c", fmt.Sprintf("ffmpeg -i %s -vframes 10 -vf cropdetect -f null - 2>&1 | awk '/crop/ { print $NF }' | tail -1 | sed -n 's/crop=//p'", inputPath))
}

func splitVideo(filename string, chunkTime time.Duration, index int) ([]byte, error) {
	return execCmd("ffmpeg",
		"-ss", parseDuration(chunkTime*time.Duration(index)),
		"-i", filename, // input
		"-c", "copy",
		"-t", parseDuration(chunkTime),
		fmt.Sprintf("./cache/chunks/%d.mp4", index), // output
	)
}

func processChunk(filename string, index int) ([]byte, error) {
	return execCmd("ffmpeg",
		"-i", filename, // input
		"-vf", fmt.Sprintf("fps=%d", _fps),
		fmt.Sprintf("./cache/frames/chunk-%d/", index)+"%d.jpg", // output
	)
}

func getVideoDuration(filename string) ([]byte, error) {
	return execCmd("bash", "-c", fmt.Sprintf("ffprobe -i %s -show_format -v quiet | sed -n 's/duration=//p'", filename))
}

func execCmd(commad ...string) ([]byte, error) {
	zap.L().Debug("Executing cmd:" + strings.Join(commad, " "))
	cmd := exec.Command(commad[0], commad[1:]...)
	// cmd.Stderr = cmd.Stdout

	return cmd.CombinedOutput()
}
