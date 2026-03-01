// api/upload.go — загрузка и валидация логотипов токенов (250x250 px, тип картинка).

package api

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

// MaxLogoWidth и MaxLogoHeight — максимальные размеры логотипа токена в пикселях.
const MaxLogoWidth, MaxLogoHeight = 250, 250

// AllowedLogoTypes — MIME-типы изображений для логотипа.
var AllowedLogoTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
}

// ValidateTokenLogo проверяет, что данные являются изображением и размер не превышает 250x250.
// Возвращает width, height и ошибку.
func ValidateTokenLogo(data []byte, contentType string) (width, height int, err error) {
	ct := strings.TrimSpace(strings.Split(contentType, ";")[0])
	if !AllowedLogoTypes[ct] {
		return 0, 0, fmt.Errorf("недопустимый тип файла: %s (ожидается image/png, image/jpeg, image/gif)", ct)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("файл не является изображением или повреждён: %w", err)
	}
	if cfg.Width > MaxLogoWidth || cfg.Height > MaxLogoHeight {
		return cfg.Width, cfg.Height, fmt.Errorf("размер изображения %dx%d превышает допустимый %dx%d", cfg.Width, cfg.Height, MaxLogoWidth, MaxLogoHeight)
	}
	return cfg.Width, cfg.Height, nil
}

// LogoExtByContentType возвращает расширение файла по MIME-типу.
func LogoExtByContentType(contentType string) string {
	ct := strings.TrimSpace(strings.Split(contentType, ";")[0])
	switch ct {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpeg"
	case "image/gif":
		return ".gif"
	default:
		return ".bin"
	}
}

// SafeLogoFilename возвращает безопасное имя файла для логотипа (без path traversal).
func SafeLogoFilename(tokenIDOrSymbol, suffix, ext string) string {
	base := strings.TrimSpace(tokenIDOrSymbol)
	if base == "" {
		base = "token"
	}
	base = strings.ReplaceAll(base, "/", "")
	base = strings.ReplaceAll(base, "\\", "")
	base = strings.ReplaceAll(base, "..", "")
	if len(base) > 64 {
		base = base[:64]
	}
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	return base + "_" + suffix + ext
}
