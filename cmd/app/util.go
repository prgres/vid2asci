package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func removeGlob(path string) {
	items, err := filepath.Glob(path)
	if err != nil {
		panic(err)
	}

	for _, item := range items {
		if strings.Contains(item, ".gitkeep") {
			continue
		}

		if err = os.RemoveAll(item); err != nil {
			panic(err)
		}
	}
}

func clear() {
	removeGlob("./asci/*")
	clearCache()
}

func clearCache() {
	removeGlob("./cache/chunks/*")
	removeGlob("./cache/frames/*")
	removeGlob("./cache/preprocess/*")
}

func parseDuration(duration time.Duration) string {
	h := duration / time.Hour
	m := (duration - h*time.Hour) / time.Minute
	s := (duration - h*time.Hour - m*time.Minute) / time.Second
	d := fmt.Sprintf("%d:%d:%d", h, m, s)

	return d
}
