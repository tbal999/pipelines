package workers

import (
	"log"
	"crypto/md5"
	"os"
	"path/filepath"
	"encoding/hex"
	"gopkg.in/yaml.v2"
	pipelines "github.com/tbal999/pipelines"
)

type Deduper struct {
	SourceFolder string
}

func (w *Deduper) Clone() pipelines.Worker {
	clone := *w

	return &clone
}

func (w *Deduper) Initialise(configBytes []byte) error {
	return yaml.Unmarshal(configBytes, &w)
}

func (w *Deduper) Close() error {
	log.Println("Deduper stopped")
	return nil
}

func (w *Deduper) Action(input []byte) ([]byte, bool, error) {
	hash := md5.Sum(input)

	hashString := hex.EncodeToString(hash[:])

	fileName := filepath.Join(w.SourceFolder, hashString)

	if fileExists(fileName) {
		return nil, false, nil
	}

	err := os.WriteFile(filepath.Join(w.SourceFolder, hashString), input, 0644)

	return input, true, err
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
