package handler

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	newPath = ""
)

func LogError(message string) {
	if err := LogToFile(message); err != nil {
		fmt.Println(fmt.Sprintf(`LogError %v`, err))
	}
}

func LogStart() {
	if err := LogToFile("scanner client started"); err != nil {
		LogError(fmt.Sprintf(`LogStart %v`, err))
	}
}

func LogCommon(message any) {
	if err := LogToFile(fmt.Sprint(message)); err != nil {
		LogError(fmt.Sprintf(`LogCommon %v`, err))
	}
}

func LogToFile(text string) error {
	if text == "" {
		return errors.New("invalid data or prefix")
	}

	// open file and create if non-existent
	os.MkdirAll(newPath, os.ModePerm)
	file, err := os.OpenFile(filepath.Join(newPath, fmt.Sprintf("%s.log", filename)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	st, err := file.Stat()
	if err != nil {
		panic(err)
	}

	if st.Size() > 4194304 {
		filename = "scanner" + time.Now().Format("-2006-01-02-15-04-05T-07-00")
		return LogToFile(text)
	} else {
		logger := log.New(file, "", log.LstdFlags)
		logger.Println(text)
	}

	return nil
}
