package web

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"strings"
	"time"
	"loop_hole/ifs"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/render"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"
	"io"
)

type Server struct {
	*server.Server
	render         *render.Render
	proxyAddr      string
	isDebug        bool
	muxControllers []ifs.MuxController
}

var ServerInstance *Server;
func WebServer() *Server{
	return ServerInstance
}

func (s *Server) AddMuxController(mcs ...ifs.MuxController) {
	s.muxControllers = append(s.muxControllers, mcs...)
}

func (s *Server) ListenAndServe(addr string) error {
	s.AddMuxController(&IndexController{})

	s.render = s.initRender()

	ignoredUrls := []string{"/javascripts/", "/images/", "/stylesheets/", "/fonts/", "/debug/vars", "/favicon", "/robots"}
	s.Middleware(NewRecoveryWare(s.isDebug))
	s.Middleware(server.NewStatWare(ignoredUrls...))
	s.Middleware(server.NewRuntimeWare(ignoredUrls, true, 15*time.Minute))
	s.EnableExtraAssetsJson("assets_map.json")

	//初始化购买加红包
	for _, mc := range s.muxControllers {
		mc.SetResponseRenderer(s)
		mc.SetUrlReverser(s)
		mc.MuxHandlers(s)
	}

	return s.Run(addr)
}

func (s *Server) initRender() *render.Render {
	tSets := []*render.TemplateSet{
	}
	for _, mc := range s.muxControllers {
		mcTSets := mc.GetTemplates()
		tSets = append(tSets, mcTSets...)
	}
	r := render.New(render.Options{
		Directory:     "templates",
		Funcs:         s.renderFuncMaps(),
		Delims:        render.Delims{"{{", "}}"},
		IndentJson:    true,
		UseBufPool:    true,
		IsDevelopment: s.isDebug,
	}, tSets)
	log.Info("Templates loaded ...")
	return r
}

func (s *Server) getRuntimeStat(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	http.DefaultServeMux.ServeHTTP(w, r)
	return ctx
}

func formatTime(tm time.Time, layout string) string {
	return tm.Format(layout)
}

func (s *Server) renderFuncMaps() []template.FuncMap {
	funcs := make([]template.FuncMap, 0)
	funcs = append(funcs, s.DefaultRouteFuncs())
	funcs = append(funcs, template.FuncMap{
		"add": func(input interface{}, toAdd int) float64 {
			switch t := input.(type) {
			case int:
				return float64(t) + float64(toAdd)
			case int64:
				return float64(t) + float64(toAdd)
			case int32:
				return float64(t) + float64(toAdd)
			case float32:
				return float64(t) + float64(toAdd)
			case float64:
				return t + float64(toAdd)
			default:
				return float64(toAdd)
			}
		},
		"formatTime": formatTime,
	})
	return funcs
}

func (s *Server) RenderJsonOr500(w http.ResponseWriter, status int, v interface{}) {
	s.renderJsonOr500(w, status, v)
}

func (s *Server) renderJsonOr500(w http.ResponseWriter, status int, v interface{}) {
	if err := s.render.Json(w, status, v); err != nil {
		log.Errorf("Server got a json rendering error, %q", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) RenderHtmlOr500(w http.ResponseWriter, status int, name string, binding interface{}) {
	s.renderHtmlOr500(w, status, name, binding)
}

func (s *Server) RenderRawHtml(w http.ResponseWriter, status int, htmlString string) {
	s.renderString(w, status, htmlString)
}

func (s *Server) renderHtmlOr500(w http.ResponseWriter, status int, name string, binding interface{}) {
	w.Header().Set("Cache-Control", "no-store, no-cache")
	if err := s.render.Html(w, status, name, binding); err != nil {
		log.Errorf("Server got a rendering error, %q", err)
		if s.isDebug {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			//渲染一个500错误页面
			s.RenderError500(w, err)
		}
	}
}

func (s *Server) Get(path, name string, handle server.Handler) {
	newHandle := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (newCtx context.Context) {
		newCtx = ctx
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, s.isDebug)]
				msg := fmt.Sprintf("Request: %s \r\n PANIC: %s\n%s", r.URL.String(), err, stack)
				log.Error(msg)
				s.RenderError500(w, errors.New(msg))
			}
		}()
		newCtx = handle(ctx, w, r)
		return
	}
	s.Server.Get(path, name, newHandle)
}

func (s *Server) RenderError404(w http.ResponseWriter) {
	s.render.Html(w, http.StatusNotFound, "page_not_found_error", nil)
}

func (s *Server) RenderError500(w http.ResponseWriter, err error) {
	params := make(map[string]interface{})
		params["Code"] = strings.ToUpper("error")
		params["HasCode"] = true
	s.render.Html(w, http.StatusInternalServerError, "internal_server_error", params)
}

func (s *Server) GetJson(path string, name string, handle ifs.JsonHandler) {
	s.Get(path, name, s.makeJsonHandler(handle))
}

func (s *Server) PostJson(path string, name string, handle ifs.JsonHandler) {
	s.Post(path, name, s.makeJsonHandler(handle))
}
func (s *Server) makeJsonHandler(handle ifs.JsonHandler) server.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		resp := handle(ctx, w, r)
		s.renderJsonOr500(w, 200, resp)
		return ctx
	}
}

func NewServer(isDebug bool) *Server {
	if isDebug {
		log.EnableDebug()
	}
	srv := &Server{
		isDebug:        isDebug,
		muxControllers: []ifs.MuxController{},
	}

	ctx := context.Background()
	srv.Server = server.New(ctx, isDebug)

	return srv
}

type BaseHandler struct {
	rr ifs.ResponseRenderer
	s  ifs.UrlReverser
}

func (s *Server) renderString(w http.ResponseWriter, status int, data string) {
	out := new(bytes.Buffer)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(status)
	out.Write([]byte(data))
	io.Copy(w, out)
}

func (s *Server) renderStringWithContentType(w http.ResponseWriter, status int, data, ct string) {
	out := new(bytes.Buffer)
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(status)
	out.Write([]byte(data))
	io.Copy(w, out)
}

func (d *BaseHandler) SetResponseRenderer(rr ifs.ResponseRenderer) {
	d.rr = rr
}

func (d *BaseHandler) SetUrlReverser(s ifs.UrlReverser) {
	d.s = s
}
