package proxyfinder

import (
    "io/ioutil"
    "regexp"
    "fmt"
    "strings"
    "strconv"
)

func FinderCNProxy(finder *Finder, o map[string]interface{}) []Proxy {
    url := "http://www.cnproxy.com/"

    c := finder.HttpClient()
    res, err := c.Get(url, nil)
    finder.MustPass(err)

    list := NewProxyContainer()

    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)

    finder.MustPass(err)

    re := regexp.MustCompile(`"(proxy[a-z0-9]+\.html)"`)

    links := re.FindAllString(string(body), -1)
    for i := 0; i < len(links); i++ {
        links[i] = strings.Replace(links[i], `"`, "", -1)
    }

    links = StringArrayUnique(links)

    chs := make([]chan error, len(links))

    for i := 0; i < len(links); i++ {
        chs[i] = make(chan error)
        go func(url string, ch chan error) {
            res, err := c.Get(url, nil)
            if err != nil {
                ch <- err
                return
            }

            defer res.Body.Close()
            body, err := ioutil.ReadAll(res.Body)

            if err != nil {
                ch <- err
                return
            }

            proxies, err := parseList(string(body))

            if err != nil {
                ch <- err
                return
            }

            list.Add(proxies...)

            ch <- nil
        }(url + links[i], chs[i])
    }

    for i := 0; i < len(links); i++ {
        <- chs[i]
    }

    return list.Proxies
}

func parseMap(content string) (map[string]string, error) {
    re := regexp.MustCompile(`(?i)<SCRIPT\s+type="text/javascript">\s*(([a-z]="\d";)+)\s*</SCRIPT>`)
    ss := re.FindAllStringSubmatch(content, -1)
    if len(ss) == 0 {
        return nil, fmt.Errorf("parseMap error")
    }

    s := ss[0][1]

    re = regexp.MustCompile(`(?i)([a-z])="(\d)";`)
    ss = re.FindAllStringSubmatch(s, -1)

    m := make(map[string]string)
    for _, v := range(ss) {
        m[v[1]] = v[2]
    }
    
    return m, nil
}

func parseList(content string) ([]Proxy, error) {
    m, err := parseMap(content)
    if err != nil {
        return nil, err
    }

    re := regexp.MustCompile(`(?i)<tr><td>((\d{1,3}\.){3}\d{1,3})<SCRIPT type=text/javascript>document\.write\(":"((\+[a-z])+)\)</SCRIPT></td><td>([a-z0-9]+)</td>`)
    matches := re.FindAllStringSubmatch(content, -1)
    if len(matches) ==  0 {
        return nil, fmt.Errorf("No match")
    }

    var proxies []Proxy

    for _, match := range(matches) {
        host := match[1]
        port := formatPort(match[3], m)
        t, err := ConvertProxyType(match[5])
        if err != nil {
            return nil, err
        }

        p := Proxy{}
        p.Host = host
        p.Port = 0
        p.Type = 0

        proxies = append(proxies, Proxy{
            Host: host,
            Port: port,
            Type: t,
        })
    }

    return proxies, nil
}

func formatPort(s string, m map[string]string) int {
    s = strings.Replace(s, "+", "", -1)
    for k, v := range m {
        s = strings.Replace(s, k, v, -1)
    }

    port, _ := strconv.Atoi(s)

    return port
}