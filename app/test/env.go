package test

import (
	"log"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/golangci/golangci-worker/app/utils/fsutils"
	"github.com/joho/godotenv"
)

var initOnce sync.Once

func loadEnv() {
	envNames := []string{".env"}
	for _, envName := range envNames {
		fpath := path.Join(fsutils.GetProjectRoot(), envName)
		err := godotenv.Overload(fpath)
		if err != nil {
			log.Fatalf("Can't load %s: %s", envName, err)
		}
	}
}

func Init() {
	initOnce.Do(func() {
		loadEnv()
	})
}

func MarkAsSlow(t *testing.T) {
	if os.Getenv("SLOW_TESTS_ENABLED") != "1" {
		t.SkipNow()
	}
}
