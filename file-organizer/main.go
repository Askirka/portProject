package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	statistics     map[string]*FileStats
	totalSize      int64
}

type FileStats struct {
	Count     int
	TotalSize int64
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

	return &FileOrganizer{sourceDir: sourceDir, statistics: make(map[string]*FileStats)}, nil

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
	fileName := filepath.Base(sourcePath)
	destinationDir := filepath.Join(fo.sourceDir, targetDir)
	destinationPath := filepath.Join(destinationDir, fileName)

	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	fileSize := info.Size()

	fo.totalSize += fileSize

	stats, ok := fo.statistics[targetDir]

	if ok {

		stats.Count++
		stats.TotalSize += fileSize
	} else {

		fo.statistics[targetDir] = &FileStats{
			Count:     1,
			TotalSize: fileSize,
		}
	}

	_, err = os.Stat(destinationPath)
	if err == nil {
		fileExt := filepath.Ext(fileName)
		fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		newFileName := fmt.Sprintf("%s_%s%s", fileNameWithoutExt, timestamp, fileExt)
		destinationPath = filepath.Join(destinationDir, newFileName)
	}

	CreateDirectory := os.MkdirAll(destinationDir, 0755)
	if CreateDirectory != nil {
		fo.logError("не удалось создать директорию +destinationDir, CreateDirectory")
		return fmt.Errorf("ошибка создания директории %s: %w", destinationDir, CreateDirectory) // <
	}
	err = os.Rename(sourcePath, destinationPath)
	if err != nil {
		fo.logError("не удалось переместить файл +fileName, err")

		return fmt.Errorf("ошибка перемещения файла %s: %w", fileName, err)
	}

	fo.logSuccess("\"Перемещён: \" + sourcePath + \" -> \" + destinationPath")

	return nil

}

func (fo *FileOrganizer) Organize() error {

	err := fo.initLog()
	if err != nil {
		return err
	}

	err = filepath.WalkDir(fo.sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil // пропускаем директории
		}

		if filepath.Dir(path) != fo.sourceDir {
			return nil
		}

		ext := filepath.Ext(path)
		ext = strings.ToLower(ext)

		value, ok := fo.rulesMap[ext]
		if ok {
			err = fo.moveFile(path, value)
			if err != nil {
				return err
			}
		}

		fo.processedFiles++
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (fo *FileOrganizer) generateReport() string {

}

func (fs *FileStats) String() string {
	SizeInfo := fmt.Sprintf("Файлов:%d,размер :%.2f KB", fs.Count, float64(fs.TotalSize)/1024)
	return SizeInfo
}
