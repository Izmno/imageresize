package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/gographics/imagick.v3/imagick"
)

var (
	ErrNotExist = fs.ErrNotExist
)

const (
	Width    = 200   // Default width of resized image
	CacheTTL = 86400 // Default cache TTL in seconds
)

type Config struct {
	Port      uint
	MediaPath string
}

func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	conf := &Config{}

	app := cli.App{
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Value:       8080,
				Usage:       "Port to listen on",
				EnvVars:     []string{"PORT"},
				Destination: &conf.Port,
			},
			&cli.StringFlag{
				Name:        "mediapath",
				Value:       "/media",
				Usage:       "Path of media directory to serve images from",
				EnvVars:     []string{"MEDIA_PATH"},
				Destination: &conf.MediaPath,
			},
		},
		Action: func(c *cli.Context) error {
			return Run(c.Context, conf)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("applicagtion stopped")
	}
}

func Run(ctx context.Context, conf *Config) error {
	resizer := &Resizer{
		width:     Width,
		cacheTTL:  CacheTTL,
		mediaPath: conf.MediaPath,
	}

	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), resizer)

	return err
}

type Resizer struct {
	width     uint
	cacheTTL  uint
	mediaPath string
}

func (rsz *Resizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = path.Clean(r.URL.Path)

	ra := AnalyzeResponse(w)
	t0 := time.Now()

	rsz.Resize(ra, r)

	t1 := time.Now()
	msg := fmt.Sprintf("[%d] %s %s (%d kb in %d ms)", ra.Status(), r.Method, r.URL.Path, ra.n/1024, t1.Sub(t0).Milliseconds())

	if ra.Status() >= 400 {
		log.Error().Msg(msg)
	} else {
		log.Info().Msg(msg)
	}
}

func (rsz *Resizer) Resize(w http.ResponseWriter, r *http.Request) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var (
		srcFile = path.Join(rsz.mediaPath, r.URL.Path)
		imgType = filepath.Ext(srcFile)[1:]
	)

	if !rsz.isImageTypeAllowed(imgType) {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	if err := readImage(mw, srcFile); err != nil {
		if errors.Is(err, ErrNotExist) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	var (
		srcWidth  = mw.GetImageWidth()
		srcHeight = mw.GetImageHeight()
		dstWidth  = rsz.width
		dstHeight = uint(float64(srcHeight) * (float64(dstWidth) / float64(srcWidth)))
	)

	if err := mw.ResizeImage(dstWidth, dstHeight, imagick.FILTER_LANCZOS); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", rsz.getMime(imgType))
	w.Header().Set("Cache-Control", "max-age="+string(rsz.cacheTTL))

	blob, err := mw.GetImageBlob()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, bytes.NewReader(blob)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (rsz *Resizer) isImageTypeAllowed(ext string) bool {
	allowed := []string{"jpg", "jpeg", "png", "gif", "webp"}

	for _, a := range allowed {
		if a == ext {
			return true
		}
	}

	return false
}

func (rsz *Resizer) getMime(extension string) string {
	mimes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
	}

	if _, ok := mimes[extension]; !ok {
		return "application/octet-stream"
	}

	return mimes[extension]
}

// readImage reads an image from the file system into a MagickWand.
//
// returns ErrNotExist if the file does not exist or is a directory.
func readImage(mw *imagick.MagickWand, f string) error {
	s, err := os.Stat(f)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNotExist
		}

		return err
	}

	if s.IsDir() {
		return ErrNotExist
	}

	return mw.ReadImage(f)
}

type responseAnalysis struct {
	n      int
	status int

	w http.ResponseWriter
}

func (ra *responseAnalysis) Status() int {
	if ra.status == 0 {
		return http.StatusOK
	}

	return ra.status
}

func (ra *responseAnalysis) Header() http.Header {
	return ra.w.Header()
}

func (ra *responseAnalysis) Write(b []byte) (int, error) {
	n, err := ra.w.Write(b)
	ra.n += n

	return n, err
}

func (ra *responseAnalysis) WriteHeader(statusCode int) {
	ra.status = statusCode

	ra.w.WriteHeader(statusCode)
}

func AnalyzeResponse(w http.ResponseWriter) *responseAnalysis {
	return &responseAnalysis{
		w: w,
	}
}
