package graphqlapp

import (
	_ "embed"
	"html/template"
)

const playVersion = "1.7.33"
const playPackageName = "@apollographql/graphql-playground-react"

//go:embed playground.html
var playHTML string

//go:embed playground.js
var playJS string

//go:embed playground.css
var playCSS string

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(playHTML))
