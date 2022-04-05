package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
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

	_resolution = "60"
	_xw         = resolutionMap[_resolution]["_xw"]
	_xh         = resolutionMap[_resolution]["_xh"]
	_fps        = 20
	_vb         = 10.0
	_vc         = 10.0
	_vs         = 10.0
	_chunkTime  = 10 * time.Second
)

func Play() error {
	frames, err := getFrames()
	if err != nil {
		return err
	}
	framesNum := len(frames)
	if framesNum == 0 {
		return errors.New("no frames found, exec 'vid2asci render // start' first")
	}

	sortFrames(frames)

	fmt.Println("ENTER to play")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	for iFrame, frame := range frames {
		fmt.Printf("\x1bc")
		// fmt.Printf("\x1b[2J") //clear term  https://imzye.com/Go/golang-clear-screen/

		if err := printFrame(fmt.Sprintf("./asci/%s", frame.Name())); err != nil {
			return err
		}
		fmt.Printf("\n\n******\nframe %d/%d\n", iFrame, framesNum)
		time.Sleep(((1000 * time.Millisecond) / time.Duration(_fps)))
	}

	return nil
}

func getFrames() ([]fs.FileInfo, error) {
	result := make([]fs.FileInfo, 0)

	files, err := ioutil.ReadDir("./asci")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		result = append(result, file)
	}

	return result, nil
}

func sortFrames(files []os.FileInfo) {
	fileNameToInt := func(fileName string) int {
		zap.S().Debugf("filename: %s", fileName)
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

func printFrame(path string) error {
	frameFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer frameFile.Close()

	buffer := make([]byte, 4096)

	for {

		b, err := frameFile.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		fmt.Print(string(buffer[:b]))
	}

	return nil
}
