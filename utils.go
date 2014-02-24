package proxyfinder

import (
    "strings"
    "fmt"
    "regexp"
    "log"
    "strconv"
)

func StringArrayUnique(arr []string) []string {
    index := make(map[string]bool)
    var result []string
    for _, v := range arr {
        if _, ok := index[v]; !ok {
            index[v] = true
            result = append(result, v)
        }
    }

    return result
}

func ConvertProxyType(t string) (int, error) {
    t = strings.ToUpper(t)
    switch {
        case t == "HTTP": return PROXY_HTTP, nil
        case t == "SOCKS4": return PROXY_SOCKS4, nil
        case t == "SOCKS5": return PROXY_SOCKS5, nil
        case t == "SOCKS4A": return PROXY_SOCKS4A, nil
    }

    return 0, fmt.Errorf(`Proxy type "%s" is not recognized`, t)
}

func DumpProxyList(proxies []Proxy) string {
    var s string
    for _, p := range proxies {
        s += fmt.Sprintf("%s:%d\t%s\n", p.Host, p.Port, p.GetDisplayType())
    }

    return s
}

func ParseProxyList(s string) []Proxy {
    ss := strings.Split(s, "\n")
    var proxies []Proxy

    re := regexp.MustCompile(`(?i)([a-z0-9\._-]+):(\d+)(\t([a-z0-9]+))`)
    for _, s := range ss {
        s := strings.TrimSpace(s)
        if s != "" {
            matches := re.FindStringSubmatch(s)
            if len(matches) == 0 {
                log.Printf("Parse proxy line failed: %s\n", s)
            } else {
                port, _ := strconv.Atoi(matches[2])
                t, _ := ConvertProxyType(matches[4])
                proxies = append(proxies, Proxy {
                    Host: matches[1],
                    Port: port,
                    Type: t,
                })
            }
        }
    }

    return proxies
}