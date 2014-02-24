package proxyfinder

import (
    "sync"
)

const (
    PROXY_HTTP = 0
    PROXY_SOCKS4 = 4
    PROXY_SOCKS5 = 5
    PROXY_SOCKS4A = 6
)

type Proxy struct {
    Host string
    Port int
    Type int
}

func (p Proxy) GetDisplayType() string {
    switch p.Type {
    case PROXY_HTTP:
        return "HTTP"
    case PROXY_SOCKS4:
        return "SOCKS4"
    case PROXY_SOCKS5:
        return "SOCKS5"
    case PROXY_SOCKS4A:
        return "SOCKS4A"
    }

    return ""
}

func NewProxyContainer() *ProxyContainer {
    return &ProxyContainer{
    }
}

type ProxyContainer struct {
    Proxies []Proxy
    mu sync.Mutex
}

func (this *ProxyContainer) Add(proxies ...Proxy) {
    this.mu.Lock()
    this.Proxies = append(this.Proxies, proxies...)
    this.mu.Unlock()
}