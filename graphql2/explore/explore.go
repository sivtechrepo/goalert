package explore

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

//go:embed build/playground.css
var playCSS string

//go:embed build/playground.js
var playJS string

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui"
    />
    <title>{{ .ApplicationName }} - GraphQL API</title>
    <style type="text/css">
       {{ .PlayCSS }}
    </style>
  </head>
  <body>
    <div id="root" />
    <script type="text/javascript">
      {{ .PlayJS }}
    </script>
  </body>
</html>
`))

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
	data.PlayJS = template.JS(playJS)
	data.PlayCSS = template.CSS(playCSS)

	err = playTmpl.Execute(w, data)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
