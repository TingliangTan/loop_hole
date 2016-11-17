package web

import (
  "loop_hole/ifs"
  "github.com/mijia/sweb/render"
  "net/http"
  "golang.org/x/net/context"
)

type IndexController struct {
  ifs.BaseMux
}

func (ic *IndexController) MuxHandlers(s ifs.Muxer) {
  s.Get("/index", "GetIndex", ic.getIndex)
}


func (c *IndexController) GetTemplates() []*render.TemplateSet {
  return []*render.TemplateSet{
    render.NewTemplateSet("index", "index.html", "index.html"),
  }
}
func (c *IndexController) getIndex(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
  c.RenderHtmlOr500(w, http.StatusOK, "index", nil)
  return ctx
}

