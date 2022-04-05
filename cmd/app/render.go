package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/prgres/img2asci/pkg/img2asci"
	"go.uber.org/zap"
)

var (
	_filenameScaled          = "./cache/preprocess/scaled.mp4"
	_filenameScaledBlackBars = "./cache/preprocess/scaled-blackbars.mp4"
)

func Render(inputPath string) error {
	clear()
	if err := preprocess(inputPath); err != nil {
		return err
	}

	frames := 1
	startTime := time.Now()
	_filename := _filenameScaledBlackBars

	duration, err := videoDuration(_filename)
	if err != nil {
		return err
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
		return err
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
			return err
		}

		fr := fmt.Sprintf("./cache/frames/chunk-%d", chunkIndex)
		if err := os.MkdirAll(fr, os.ModePerm); err != nil {
			return err
		}

		if _, err = processChunk(fmt.Sprintf("./cache/chunks/%d.mp4", chunkIndex), chunkIndex); err != nil {
			return err
		}

		files, err := ioutil.ReadDir(fr)
		if err != nil {
			return err
		}

		for _, file := range files {
			inputPath := fmt.Sprintf("%s/%s", fr, file.Name())
			outputPath := fmt.Sprintf("./asci/%d.txt", frames)

			if err := frameToAsci(asciCfg, inputPath, outputPath); err != nil {
				return err
			}

			frames++
		}
	}

	clearCache()
	zap.S().Infof("Time: %d", time.Since(startTime))
	return nil
}

func frameToAsci(cfg *img2asci.Config, inputPath string, outputPath string) error {
	return cfg.Process(inputPath, outputPath)
}

func preprocess(inputPath string) error {
	if _, err := resizeVideo(inputPath, _filenameScaled); err != nil {
		return err
	}

	blackBarsParam, err := getCropBlackBarsParams(_filenameScaled)
	if err != nil {
		return err
	}

	if _, err := cropBlackBars(_filenameScaled, _filenameScaledBlackBars, strings.TrimSpace(string(blackBarsParam))); err != nil {
		return err
	}

	return nil
}

func videoDuration(path string) (time.Duration, error) {
	o, err := getVideoDuration(path)
	if err != nil {
		return time.Duration(-1), err
	}

	duration, err := time.ParseDuration(strings.TrimSpace(string(o)) + "s")
	if err != nil {
		return time.Duration(-1), err
	}

	return duration, nil
}
