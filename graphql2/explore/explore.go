package explore

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

//go:embed explore.html
var htmlStr string

//go:embed build/explore.css
var cssStr string

//go:embed build/explore.js
var jsStr string

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

	cfg := config.FromContext(ctx)
	data.ApplicationName = cfg.ApplicationName()
	data.PlayJS = template.JS(jsStr)
	data.PlayCSS = template.CSS(cssStr)

	err = playTmpl.Execute(w, data)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
