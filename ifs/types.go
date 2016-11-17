package ifs

import (
	"net/http"

	"github.com/mijia/sweb/render"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"
)

type JsonHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) interface{}

type ResponseRenderer interface {
	RenderJsonOr500(w http.ResponseWriter, status int, v interface{})
	RenderHtmlOr500(w http.ResponseWriter, status int, name string, binding interface{})
	RenderRawHtml(w http.ResponseWriter, status int, htmlString string)
	RenderError500(w http.ResponseWriter, err error)
	RenderError404(w http.ResponseWriter)
}

type Muxer interface {
	server.Muxer
	GetJson(path string, name string, handle JsonHandler)
	PostJson(path string, name string, handle JsonHandler)
}

type UrlReverser interface {
	Reverse(name string, params ...interface{}) string
	Assets(path string) string
}

type MuxController interface {
	MuxHandlers(m Muxer)
	SetResponseRenderer(r ResponseRenderer)
	SetUrlReverser(r UrlReverser)
	GetTemplates() []*render.TemplateSet
}
