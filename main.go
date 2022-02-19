package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tm "github.com/buger/goterm"
	"github.com/prgres/img2asci/pkg/img2asci"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	resolutionMap = map[string]map[string]int{
		"1080": {
			"_xw": 1920,
			"_xh": 1080,
		},
		"720": {
			"_xw": 1280,
			"_xh": 720,
		},
		"480": {
			"_xw": 640,
			"_xh": 480,
		},
		"360": {
			"_xw": 480,
			"_xh": 360,
		},
		"240": {
			"_xw": 320,
			"_xh": 240,
		},
		"120": {
			"_xw": 160,
			"_xh": 120,
		},
		"60": {
			"_xw": 80,
			"_xh": 60,
		},
	}

	_resolution  = "60"
	_xw          = resolutionMap[_resolution]["_xw"]
	_xh          = resolutionMap[_resolution]["_xh"]
	_fps         = 20
	_vb          = 10.0
	_vc          = 10.0
	_vs          = 10.0
	_input_video = "./video.mp4"
	_chunkTime   = 10 * time.Second
)

func removeGlob(path string) error {
	items, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	for _, item := range items {
		if strings.Contains(item, ".gitkeep") {
			continue
		}

		if err = os.RemoveAll(item); err != nil {
			return err
		}
	}

	return nil
}

func clear() {
	if err := removeGlob("./asci/*"); err != nil {
		panic(err)
	}

	if err := removeGlob("./chunks/*"); err != nil {
		panic(err)
	}

	if err := removeGlob("./frames/*"); err != nil {
		panic(err)
	}

	if err := removeGlob("./preprocess/*"); err != nil {
		panic(err)
	}
}

func logger(debug bool) error {
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}

	cfg := zap.Config{
		Encoding:    "console",
		Level:       zap.NewAtomicLevelAt(logLevel),
		OutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			TimeKey:     "time",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
	}

	log, err := cfg.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(log)

	return nil
}

func main() {
	startTime := time.Now()
	if err := logger(true); err != nil {
		log.Fatal(err)
	}

	clear()

	frames := 1
	_filenameScaled := "./preprocess/scaled.mp4"

	if _, err := resizeVideo(_input_video, _filenameScaled); err != nil {
		zap.L().Fatal(err.Error())
	}
	// _filename := _filenameScaled

	_filenameScaledBlackBars := "./preprocess/scaled-blackbars.mp4"

	blackBarsParam, err := getCropBlackBarsParams(_filenameScaled)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := cropBlackBars(_filenameScaled, _filenameScaledBlackBars, strings.TrimSpace(string(blackBarsParam))); err != nil {
		zap.L().Fatal(err.Error())
	}
	_filename := _filenameScaledBlackBars

	o, err := getVideoDuration(_filename)
	if err != nil {
		log.Fatal(err)
	}

	duration, err := time.ParseDuration(strings.TrimSpace(string(o)) + "s")
	if err != nil {
		log.Fatal(err)
	}

	durationRound := duration.Round(time.Second)
	secs := int(time.Duration(durationRound) / time.Second)
	chunkTimeInt := int(time.Duration(_chunkTime) / time.Second)

	zap.L().Info("Video duration: " + durationRound.String())

	maxIndex := secs / chunkTimeInt
	if secs%chunkTimeInt != 0 {
		maxIndex += 1
	}

	dLog, err := zap.NewStdLogAt(zap.L(), zap.DebugLevel)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	asciCfg := &img2asci.Config{
		Term: false,
		Log:  dLog,
		ProcessingValues: &img2asci.ProcessingValues{
			Width:              _xw,
			Height:             _xh,
			Sharp:              _vs,
			Bright:             _vb,
			Contrast:           _vc,
			GrayScaleAsciTable: img2asci.DeafultGrayScale,
		},
	}

	for i := 1; i <= maxIndex; i++ {
		chunkIndex := i - 1
		if _, err := splitVideo(_filename, _chunkTime, chunkIndex); err != nil {
			log.Fatal(err.Error())
		}
		fr := fmt.Sprintf("./frames/chunk-%d", chunkIndex)
		if err := os.MkdirAll(fr, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if _, err = processChunk(fmt.Sprintf("./chunks/%d.mp4", chunkIndex), chunkIndex); err != nil {
			log.Fatal(err)
		}

		files, err := ioutil.ReadDir(fr)
		if err != nil {
			log.Fatal(err)
		}

		sortFrames(files)

		for _, file := range files {
			inputPath := fmt.Sprintf("%s/%s", fr, file.Name())
			outputPath := fmt.Sprintf("./asci/%d.txt", frames)

			if err := frameToAsci(asciCfg, inputPath, outputPath); err != nil {
				log.Fatal(err)
			}

			frames++
		}
	}

	allFrames, err := ioutil.ReadDir("./asci")
	if err != nil {
		log.Fatal(err)
	}

	sortFrames(allFrames)

	zap.S().Infof("Time: %d", time.Since(startTime))
	fmt.Println("ENTER to play")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	for _, frame := range allFrames {
		if err := printScreen(fmt.Sprintf("./asci/%s", frame.Name())); err != nil {
			log.Fatal(err)
		}
	}
}

func sortFrames(files []os.FileInfo) {
	fileNameToInt := func(fileName string) int {
		raw := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		intVal, err := strconv.Atoi(raw)
		if err != nil {
			panic(err)
		}

		return intVal
	}

	sort.Slice(files, func(i, j int) bool {
		return fileNameToInt(files[i].Name()) < fileNameToInt(files[j].Name())
	})
}

func printScreen(path string) error {
	frameFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer frameFile.Close()

	buffer := make([]byte, 4096)

	tm.Clear() // Clear current screen

	// By moving cursor to top-left position we ensure that console output
	// will be overwritten each time, instead of adding new.
	tm.MoveCursor(1, 1)

	tm.Println(path)

	for {
		b, err := frameFile.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		tm.Print(string(buffer[:b]))
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(((1000 * time.Millisecond) / time.Duration(_fps)))

	return nil
}

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
		fmt.Sprintf("./chunks/%d.mp4", index), // output
	)
}

func parseDuration(duration time.Duration) string {
	h := duration / time.Hour
	m := (duration - h*time.Hour) / time.Minute
	s := (duration - h*time.Hour - m*time.Minute) / time.Second
	d := fmt.Sprintf("%d:%d:%d", h, m, s)

	zap.L().Debug("Duration:" + d)

	return d
}

func processChunk(filename string, index int) ([]byte, error) {
	return execCmd("ffmpeg",
		"-i", filename, // input
		"-vf", fmt.Sprintf("fps=%d", _fps),
		fmt.Sprintf("./frames/chunk-%d/", index)+"%d.jpg", // output
	)
}

func getVideoDuration(filename string) ([]byte, error) {
	return execCmd("bash", "-c", fmt.Sprintf("ffprobe -i %s -show_format -v quiet | sed -n 's/duration=//p'", filename))
}

func frameToAsci(cfg *img2asci.Config, inputPath string, outputPath string) error {
	return cfg.Process(inputPath, outputPath)
}

func execCmd(commad ...string) ([]byte, error) {
	zap.L().Debug("Executing cmd:" + strings.Join(commad, " "))
	cmd := exec.Command(commad[0], commad[1:]...)
	// cmd.Stderr = cmd.Stdout

	return cmd.CombinedOutput()
}
