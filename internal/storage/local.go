package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/felip/api-fidelidade/internal/id"
)

type Local struct {
	baseDir       string
	publicBaseURL string
}

func NewLocal(baseDir string, publicBaseURL string) (*Local, error) {
	profilesDir := filepath.Join(baseDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	return &Local{baseDir: baseDir, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

func (l *Local) SaveProfileImage(userID string, contentType string, data []byte) (string, error) {
	var extension string
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	default:
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	fileID, err := id.New("img_")
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s%s", userID, strings.TrimPrefix(fileID, "img_"), extension)
	absolutePath := filepath.Join(l.baseDir, "profiles", filename)
	if err := os.WriteFile(absolutePath, data, 0o644); err != nil {
		return "", fmt.Errorf("write image file: %w", err)
	}

	return l.publicBaseURL + "/uploads/profiles/" + filename, nil
}
