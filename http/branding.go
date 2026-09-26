package fbhttp

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// brandingKind describes one uploadable branding file. The frontend always
// asks for the first name in variants; any of them answers that request, so
// an uploaded logo.png also serves /static/img/logo.svg.
type brandingKind struct {
	variants []string
}

var brandingKinds = map[string]brandingKind{
	"logo":               {variants: []string{"img/logo.svg", "img/logo.png", "img/logo.jpg", "img/logo.jpeg"}},
	"favicon":            {variants: []string{"img/icons/favicon.svg", "img/icons/favicon.ico", "img/icons/favicon.png"}},
	"apple-touch-icon":   {variants: []string{"img/icons/apple-touch-icon.png"}},
	"android-chrome-192": {variants: []string{"img/icons/android-chrome-192x192.png"}},
	"android-chrome-512": {variants: []string{"img/icons/android-chrome-512x512.png"}},
}

// target returns where an upload with the given extension is stored.
func (k brandingKind) target(ext string) (string, bool) {
	for _, variant := range k.variants {
		if strings.EqualFold(filepath.Ext(variant), ext) {
			return variant, true
		}
	}
	return "", false
}

// find returns the uploaded file of this kind in dir, or "".
func (k brandingKind) find(dir string) (string, os.FileInfo) {
	for _, variant := range k.variants {
		fPath := filepath.Join(dir, filepath.FromSlash(variant))
		info, err := os.Stat(fPath)
		if err == nil && !info.IsDir() {
			return fPath, info
		}
		if err != nil && !os.IsNotExist(err) {
			log.Printf("could not load branding file override: %v", err)
		}
	}
	return "", nil
}

// remove deletes every variant, so an old logo.svg cannot shadow a new logo.png.
func (k brandingKind) remove(dir string) error {
	for _, variant := range k.variants {
		err := os.Remove(filepath.Join(dir, filepath.FromSlash(variant)))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// brandingOverride returns the file in the branding directory that should
// answer a request for urlPath, or "" when the built-in asset should be used.
func brandingOverride(dir, urlPath string) string {
	for _, kind := range brandingKinds {
		for _, variant := range kind.variants {
			if variant == urlPath {
				fPath, _ := kind.find(dir)
				return fPath
			}
		}
	}

	fPath := filepath.Join(dir, filepath.FromSlash(urlPath))
	if info, err := os.Stat(fPath); err == nil && !info.IsDir() {
		return fPath
	} else if err != nil && !os.IsNotExist(err) {
		log.Printf("could not load branding file override: %v", err)
	}
	return ""
}

// brandingVersion changes whenever a branding file is uploaded or removed.
// The frontend appends it to logo and favicon URLs so browsers do not keep
// showing a cached older image.
func brandingVersion(dir string) string {
	if dir == "" {
		return "0"
	}

	var latest int64
	for _, kind := range brandingKinds {
		if _, info := kind.find(dir); info != nil && info.ModTime().UnixNano() > latest {
			latest = info.ModTime().UnixNano()
		}
	}
	if info, err := os.Stat(filepath.Join(dir, ".removed")); err == nil && info.ModTime().UnixNano() > latest {
		latest = info.ModTime().UnixNano()
	}
	return strconv.FormatInt(latest/int64(1e6), 36)
}

type brandingStatus struct {
	Custom map[string]bool `json:"custom"`
	// Version is the same value the page gets as BrandingVersion.
	Version string `json:"version"`
	// Configured is false while no branding directory is set in the settings.
	Configured bool `json:"configured"`
}

var brandingGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	dir := d.settings.Branding.Files
	status := brandingStatus{Custom: map[string]bool{}, Version: brandingVersion(dir), Configured: dir != ""}
	for name, kind := range brandingKinds {
		fPath := ""
		if dir != "" {
			fPath, _ = kind.find(dir)
		}
		status.Custom[name] = fPath != ""
	}
	return renderJSON(w, r, status)
})

var brandingDeleteHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	kind, ok := brandingKinds[mux.Vars(r)["type"]]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("invalid fileType")
	}

	dir := d.settings.Branding.Files
	if dir == "" {
		return http.StatusBadRequest, fmt.Errorf("branding files directory not configured")
	}

	if err := kind.remove(dir); err != nil {
		return http.StatusInternalServerError, err
	}

	// Removing lowers no modification time, so remember the moment to still
	// change the version and push browsers back to the default image.
	if err := os.WriteFile(filepath.Join(dir, ".removed"), nil, 0o644); err != nil {
		log.Printf("could not mark branding removal: %v", err)
	}
	return http.StatusOK, nil
})

var brandingUploadHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	dir := d.settings.Branding.Files
	if dir == "" {
		return http.StatusBadRequest, fmt.Errorf("branding files directory not configured")
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return http.StatusBadRequest, err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer file.Close()

	kind, ok := brandingKinds[r.FormValue("fileType")]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("invalid fileType")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	variant, ok := kind.target(ext)
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("invalid file extension: %s", ext)
	}

	targetPath := filepath.Join(dir, filepath.FromSlash(variant))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return http.StatusInternalServerError, err
	}

	// Write next to the target first so a failed upload keeps the old file.
	tmpPath := targetPath + ".upload"
	dst, err := os.Create(tmpPath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		return http.StatusInternalServerError, err
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmpPath)
		return http.StatusInternalServerError, err
	}

	if err := kind.remove(dir); err != nil {
		os.Remove(tmpPath)
		return http.StatusInternalServerError, err
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})
