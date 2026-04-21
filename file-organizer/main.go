package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var DefaultRules = map[string]string{
		".jpg":  "Images",
		".jpeg": "Images",
		".pdf":  "Documents",
		".mp3":  "Music",
		".zip":  "Archives",
		".rar":  "Archives",
		".mp4":  "Video",
		".avi":  "Video",
		".wav":  "Music",
		".doc":  "Documents",
		".docx": "Documents",
		".txt":  "Documents",
		".png":  "Images",
	}

	for key, value := range DefaultRules {
		fmt.Printf("%s: %s\n", strings.ToLower(key), strings.ToLower(value))
	}

}

type FileOrganizer struct {
	sourceDir      string
	rulesMap       map[string]string
	processedFiles int
	logFile        *os.File
}

func NewFileOrganizer(sourceDir string) (*FileOrganizer, error) {
	if sourceDir == "" {
		return nil, errors.New("Empty directory")

	}
	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("Source is not a directory")
	}

	return &FileOrganizer{sourceDir: sourceDir}, nil

}

func (fo *FileOrganizer) initLog() error {
	file, err := os.OpenFile("oranizer.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fo.logFile = file

	log.SetOutput(file)
	return nil

}

func (fo *FileOrganizer) logSuccess(message string) {
	log.Println("[SUCCESS]", message)

}

func (fo *FileOrganizer) logError(message string) {
	log.Println("[ERROR]", message)

}

func (fo *FileOrganizer) Close() error {
	if fo.logFile != nil {
		return fo.logFile.Close()

	}
	return nil

}

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string) error {
	fullPath := filepath.Join(fo.sourceDir, targetDir)
	CreateDirectory := os.MkdirAll(fullPath, 0755)
	if CreateDirectory != nil {
		return CreateDirectory
	}

	return nil

}
