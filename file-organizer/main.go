package main

import (
	"fmt"
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
