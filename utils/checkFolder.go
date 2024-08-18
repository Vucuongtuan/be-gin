package utils

import (
	"strings"
)

func CheckFolder(fileType string) (string,string) {
	var folder string
	var typeFile string
	switch {
    case strings.HasPrefix(fileType, "video"):
        folder = "uploads/video/"
		typeFile = strings.TrimPrefix(fileType, "video/")
    case strings.HasPrefix(fileType, "image"):
        folder = "uploads/image/"
		typeFile = strings.TrimPrefix(fileType, "image/")
    case strings.HasPrefix(fileType, "audio"):
        folder = "uploads/audio/"
		typeFile = strings.TrimPrefix(fileType, "audio/")
		
    default:
        folder = "uploads/other/"
		typeFile = fileType
    }
	return folder, typeFile
}
