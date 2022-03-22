package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/util/errutil"
)

//go:embed src/build
var bundleFS embed.FS

//go:embed live.js
var liveJS string

// NewHandler creates a new http.Handler that will serve UI files
// using bundled assets or locally if uiDir if set.
func NewHandler(uiDir, prefix string) (http.Handler, error) {
	mux := http.NewServeMux()

	var extraJS string
	if uiDir != "" {
		extraJS = liveJS
		mux.Handle("/static/", NoCache(NewEtagFileServer(http.Dir(uiDir), false)))
	} else {
		sub, err := fs.Sub(bundleFS, "src/build")
		if err != nil {
			return nil, err
		}
		mux.Handle("/static/", NewEtagFileServer(http.FS(sub), true))
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		cfg := config.FromContext(req.Context())

		var buf bytes.Buffer
		err := indexTmpl.Execute(&buf, renderData{
			ApplicationName: cfg.ApplicationName(),
			Prefix:          prefix,
			ExtraJS:         template.JS(extraJS),
		})
		if errutil.HTTPError(req.Context(), w, err) {
			return
		}

		h := sha256.New()
		h.Write(buf.Bytes())
		indexETag := fmt.Sprintf(`W/"sha256-%s"`, hex.EncodeToString(h.Sum(nil)))
		w.Header().Set("ETag", indexETag)

		if uiDir == "" {
			w.Header().Set("Cache-Control", "private; max-age=60, stale-while-revalidate=600, stale-if-error=259200")
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}

		http.ServeContent(w, req, "/", time.Time{}, bytes.NewReader(buf.Bytes()))
	})

	return mux, nil
}
