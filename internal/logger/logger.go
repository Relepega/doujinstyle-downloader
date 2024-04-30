package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/slog"
)

const dateFormatting = "2006-01-02"

func createLogFile(logdir string) (*os.File, error) {
	fn := fmt.Sprintf("%v.log", time.Now().Format(dateFormatting))
	fp := filepath.Join(logdir, fn)

	file, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error creating log file: %v", err)
	}

	return file, nil
}

/*
Initializes the logger.

# Parameters

- `dir` - The directory where the logs will be saved
*/
func InitLogger(dir string) {
	fileHandle, err := createLogFile(dir)
	if err != nil {
		log.Fatalln(err)
	}

	logger := slog.New(slog.NewTextHandler(fileHandle, nil)) // or os.Stdout
	slog.SetDefault(logger)
}
