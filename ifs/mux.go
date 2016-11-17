package ifs

type BaseMux struct {
	ResponseRenderer
	UrlReverser
}

func (b *BaseMux) SetResponseRenderer(rr ResponseRenderer) {
	b.ResponseRenderer = rr
}

func (b *BaseMux) SetUrlReverser(ur UrlReverser) {
	b.UrlReverser = ur
}
