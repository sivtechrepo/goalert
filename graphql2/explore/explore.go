package explore

import (
	"embed"
	_ "embed"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
)

//go:embed explore.html
var htmlStr string

//go:embed build
var fs embed.FS

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(htmlStr))

func Handler(w http.ResponseWriter, req *http.Request) {
	var data struct {
		ApplicationName string
		PlayJS          template.JS
		PlayCSS         template.CSS
	}

	ctx := req.Context()
	err := permission.LimitCheckAny(ctx)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	jspath, err := filepath.Abs(filepath.Join("graphql2", "explore", "build", "explore.js"))
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "graphql explore js path"))
		return
	}
	csspath, err := filepath.Abs(filepath.Join("graphql2", "explore", "build", "explore.css"))
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "graphql explore css path"))
		return
	}

	jsData, err := fs.ReadFile(jspath)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "read graphql explore js"))
		return
	}
	cssData, err := fs.ReadFile(csspath)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "read graphql explore css"))
		return
	}

	cfg := config.FromContext(ctx)
	data.ApplicationName = cfg.ApplicationName()
	data.PlayJS = template.JS(string(jsData))
	data.PlayCSS = template.CSS(string(string(cssData)))

	err = playTmpl.Execute(w, data)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
