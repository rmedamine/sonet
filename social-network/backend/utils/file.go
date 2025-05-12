package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
)

func GetFileExtension(filename string) string {
	return filename[strings.LastIndex(filename, "."):]
}

func SaveFile(file multipart.File, filePath string) error {
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

func GenerateRandomFilename(ext string) string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Println("Error generating random filename:", err)
		return "default_avatar" + ext
	}
	return hex.EncodeToString(bytes) + ext
}