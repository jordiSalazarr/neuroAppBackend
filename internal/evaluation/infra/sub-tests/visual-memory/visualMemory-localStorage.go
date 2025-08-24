package VIMinfra

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

type LocalImageStorage struct {
	Root string // directorio donde guardar
}

func NewLocalImageStorage(root string) *LocalImageStorage {
	_ = os.MkdirAll(root, 0o755)
	return &LocalImageStorage{Root: root}
}

func (l *LocalImageStorage) PutPNGOrJPEG(raw []byte, contentType string) (string, error) {
	ext := ".bin"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		ext = ".jpg"
	}
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	name := hex.EncodeToString(buf) + ext
	path := filepath.Join(l.Root, name)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", err
	}
	// Devuelve una referencia opaca; en local, la ruta absoluta o relativa
	return path, nil
}
