package explore

import (
	"embed"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

//go:embed explore.html
var htmlStr string

//go:embed build
var build embed.FS

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

	jsData, err := build.ReadFile("build/explore/explore.js")
	if err != nil {
		return
	}
	cssData, err := build.ReadFile("build/explore/explore.css")
	if err != nil {
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
