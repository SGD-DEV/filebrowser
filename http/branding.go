package fbhttp

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type brandingUploadRequest struct {
	FileType string `json:"fileType"`
}

var brandingUploadHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}

	if d.settings.Branding.Files == "" {
		return http.StatusBadRequest, fmt.Errorf("branding files directory not configured")
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return http.StatusBadRequest, err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer file.Close()

	fileType := r.FormValue("fileType")
	if fileType == "" {
		return http.StatusBadRequest, fmt.Errorf("fileType is required")
	}

	allowedTypes := map[string][]string{
		"logo":               {".svg", ".png", ".jpg", ".jpeg"},
		"favicon":            {".ico", ".svg", ".png"},
		"apple-touch-icon":   {".png"},
		"android-chrome-192": {".png"},
		"android-chrome-512": {".png"},
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts, ok := allowedTypes[fileType]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("invalid fileType")
	}

	valid := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			valid = true
			break
		}
	}
	if !valid {
		return http.StatusBadRequest, fmt.Errorf("invalid file extension for %s: %s", fileType, ext)
	}

	err = os.MkdirAll(d.settings.Branding.Files, 0755)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	imgDir := filepath.Join(d.settings.Branding.Files, "img")
	iconsDir := filepath.Join(imgDir, "icons")

	var targetPath string
	switch fileType {
	case "logo":
		err = os.MkdirAll(imgDir, 0755)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		removeBrandingVariants(d.settings.Branding.Files, "img/logo")
		targetPath = filepath.Join(imgDir, "logo"+ext)
	case "favicon":
		err = os.MkdirAll(iconsDir, 0755)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		removeBrandingVariants(d.settings.Branding.Files, "img/icons/favicon")
		targetPath = filepath.Join(iconsDir, "favicon"+ext)
	case "apple-touch-icon":
		err = os.MkdirAll(iconsDir, 0755)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		targetPath = filepath.Join(iconsDir, "apple-touch-icon.png")
	case "android-chrome-192":
		err = os.MkdirAll(iconsDir, 0755)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		targetPath = filepath.Join(iconsDir, "android-chrome-192x192.png")
	case "android-chrome-512":
		err = os.MkdirAll(iconsDir, 0755)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		targetPath = filepath.Join(iconsDir, "android-chrome-512x512.png")
	default:
		return http.StatusBadRequest, fmt.Errorf("unknown fileType")
	}

	dst, err := os.Create(targetPath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})

// brandingVariants lists the file names an upload of each kind can have. The
// frontend always asks for the first one; any of them answers that request.
var brandingVariants = map[string][]string{
	"img/logo":          {"img/logo.svg", "img/logo.png", "img/logo.jpg", "img/logo.jpeg"},
	"img/icons/favicon": {"img/icons/favicon.svg", "img/icons/favicon.ico", "img/icons/favicon.png"},
}

// brandingOverride returns the file in the branding directory that should
// answer a request for urlPath, or "" when the built-in asset should be used.
// A request for logo.svg is answered by an uploaded logo.png as well.
func brandingOverride(dir, urlPath string) string {
	candidates := []string{urlPath}
	if variants, ok := brandingVariants[strings.TrimSuffix(urlPath, filepath.Ext(urlPath))]; ok {
		candidates = append(candidates, variants...)
	}

	for _, candidate := range candidates {
		fPath := filepath.Join(dir, filepath.FromSlash(candidate))
		info, err := os.Stat(fPath)
		if err == nil && !info.IsDir() {
			return fPath
		}
		if err != nil && !os.IsNotExist(err) {
			log.Printf("could not load branding file override: %v", err)
		}
	}

	return ""
}

// removeBrandingVariants deletes earlier uploads of the same kind with another
// extension, so an old logo.svg cannot shadow a newly uploaded logo.png.
func removeBrandingVariants(dir, kind string) {
	for _, variant := range brandingVariants[kind] {
		if err := os.Remove(filepath.Join(dir, filepath.FromSlash(variant))); err != nil && !os.IsNotExist(err) {
			log.Printf("could not remove old branding file: %v", err)
		}
	}
}
