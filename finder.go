package proxyfinder

import (
    "github.com/ddliu/go-httpclient"
    "github.com/ddliu/goption"
    "fmt"
)

var FINDER_HANDLERS = map[string]func(*Finder, map[string]interface{})[]Proxy {
    "cnproxy.com": FinderCNProxy,
}

type FinderHandler struct {
    Handler func(*Finder, map[string]interface{}) []Proxy
    Options map[string]interface{}
}

type Finder struct {
    Handlers []FinderHandler
}

func (this *Finder) HttpClient() *httpclient.HttpClient {
    return httpclient.NewHttpClient(map[int]interface{} {
        httpclient.OPT_USERAGENT: "Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/532.5 (KHTML, like Gecko) Chrome/4.0.249.0 Safari/532.5",
    })
}

func (this *Finder) Error(err error) {
    panic(err)
}

func (this *Finder) MustOk(o bool) {
    if !o {
        this.Error(fmt.Errorf("Must OK"))
    }
}

func (this *Finder) MustPass(err error) {
    if err != nil {
        this.Error(err)
    }
}

// func (this *Finder) CacheGet(url)

func (this *Finder) Option(m map[string]interface{}) *goption.Option {
    return goption.NewOption(m)
}

func (this *Finder) AddHandler(options map[string]interface{}, handler func(*Finder, map[string]interface{}) []Proxy) {
    this.Handlers = append(this.Handlers, FinderHandler{
        Handler: handler,
        Options: options,
    })
}

func (this *Finder) Find() []Proxy {
    var result []Proxy
    for _, h := range this.Handlers {
        result = append(result, h.Handler(this, h.Options)...)
    }

    return result
}