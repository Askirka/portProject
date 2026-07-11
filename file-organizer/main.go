package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var DefaultRules = map[string]string{
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".pdf":  "Documents",
	".doc":  "Documents",
	".docx": "Documents",
	".txt":  "Documents",
	".mp3":  "Music",
	".wav":  "Music",
	".zip":  "Archives",
	".rar":  "Archives",
	".mp4":  "Video",
	".avi":  "Video",
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

func main() {
	organizer, err := NewFileOrganizer("./files")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := organizer.Close()
		if err != nil {
			fmt.Println("Ошибка закрытия лог-файла:", err)
		}
	}()

	err = organizer.Organize()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(organizer.generateReport())
}

func NewFileOrganizer(sourceDir string) (*FileOrganizer, error) {
	if sourceDir == "" {
		return nil, errors.New("путь к директории не может быть пустым")
	}

	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("указанный путь не является директорией")
	}

	return &FileOrganizer{
		sourceDir:  sourceDir,
		rulesMap:   DefaultRules,
		statistics: make(map[string]*FileStats),
	}, nil
}

func (fo *FileOrganizer) initLog() error {
	file, err := os.OpenFile(
		"organizer.log",
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0666,
	)
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

func (fo *FileOrganizer) moveFile(
	sourcePath string,
	targetDir string,
	fileSize int64,
) error {
	fileName := filepath.Base(sourcePath)

	destinationDir := filepath.Join(fo.sourceDir, targetDir)
	destinationPath := filepath.Join(destinationDir, fileName)

	err := os.MkdirAll(destinationDir, 0755)
	if err != nil {
		fo.logError(
			fmt.Sprintf(
				"не удалось создать директорию %s: %v",
				destinationDir,
				err,
			),
		)

		return fmt.Errorf(
			"ошибка создания директории %s: %w",
			destinationDir,
			err,
		)
	}

	_, err = os.Stat(destinationPath)
	if err == nil {
		fileExt := filepath.Ext(fileName)
		fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)

		timestamp := time.Now().Format("2006-01-02_15-04-05")

		newFileName := fmt.Sprintf(
			"%s_%s%s",
			fileNameWithoutExt,
			timestamp,
			fileExt,
		)

		destinationPath = filepath.Join(
			destinationDir,
			newFileName,
		)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(
			"ошибка проверки файла %s: %w",
			destinationPath,
			err,
		)
	}

	err = os.Rename(sourcePath, destinationPath)
	if err != nil {
		fo.logError(
			fmt.Sprintf(
				"не удалось переместить файл %s: %v",
				fileName,
				err,
			),
		)

		return fmt.Errorf(
			"ошибка перемещения файла %s: %w",
			fileName,
			err,
		)
	}

	// Статистика обновляется только после успешного перемещения.
	fo.processedFiles++
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

	fo.logSuccess(
		fmt.Sprintf(
			"Перемещён: %s -> %s",
			sourcePath,
			destinationPath,
		),
	)

	return nil
}

func (fo *FileOrganizer) Organize() error {
	err := fo.initLog()
	if err != nil {
		return err
	}

	err = filepath.WalkDir(
		fo.sourceDir,
		func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if d.IsDir() {
				return nil
			}

			// Обрабатываем только файлы непосредственно
			// внутри sourceDir.
			if filepath.Dir(path) != fo.sourceDir {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))

			category, ok := fo.rulesMap[ext]
			if !ok {
				return nil
			}

			// Получаем информацию о текущем файле через DirEntry.
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf(
					"не удалось получить информацию о файле %s: %w",
					path,
					err,
				)
			}

			err = fo.moveFile(
				path,
				category,
				info.Size(),
			)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return err
}

func (fo *FileOrganizer) generateReport() string {
	var builder strings.Builder

	builder.WriteString("=== Отчёт о перемещении файлов ===\n\n")

	builder.WriteString(
		fmt.Sprintf(
			"Всего обработано файлов: %d\n",
			fo.processedFiles,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Общий размер: %s\n\n",
			formatSize(fo.totalSize),
		),
	)

	builder.WriteString("Статистика по категориям:\n\n")

	categories := make([]string, 0, len(fo.statistics))

	for category := range fo.statistics {
		categories = append(categories, category)
	}

	sort.Strings(categories)

	for _, category := range categories {
		stats := fo.statistics[category]

		builder.WriteString(category)
		builder.WriteString(":\n")

		builder.WriteString(
			fmt.Sprintf("  %s\n\n", stats),
		)
	}

	return builder.String()
}

func (fileStats *FileStats) String() string {
	return fmt.Sprintf(
		"Файлов: %d, Размер: %s",
		fileStats.Count,
		formatSize(fileStats.TotalSize),
	)
}

func formatSize(size int64) string {
	const (
		kilobyte = 1024
		megabyte = 1024 * 1024
	)

	if size >= megabyte {
		sizeInMB := float64(size) / float64(megabyte)

		// Округление вниз до двух знаков после запятой.
		sizeInMB = math.Floor(sizeInMB*100) / 100

		return fmt.Sprintf("%.2f MB", sizeInMB)
	}

	sizeInKB := float64(size) / float64(kilobyte)

	// Округление вниз до двух знаков после запятой.
	sizeInKB = math.Floor(sizeInKB*100) / 100

	return fmt.Sprintf("%.2f KB", sizeInKB)
}
