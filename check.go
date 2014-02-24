package proxyfinder

import (
    "github.com/ddliu/go-httpclient"
    "fmt"
    "time"
    "sync"
    "sort"
    "log"
)

func Check(options map[int]interface{}, url string, proxy Proxy) (time.Duration, error) {
    now := time.Now()

    res, err := httpclient.NewHttpClient(options).
        WithOption(httpclient.OPT_PROXYTYPE, proxy.Type).
        WithOption(httpclient.OPT_PROXY, fmt.Sprintf("%s:%d", proxy.Host, proxy.Port)).
        Get(url, nil)

    duratioin := time.Since(now)

    if err != nil {
        return duratioin, err
    }

    if res.StatusCode != 200 {
        return duratioin, fmt.Errorf("Check proxy: %s, StatusCode: %d", proxy.Host, res.StatusCode)
    }

    return duratioin, nil
}

type CheckResult struct {
    Proxy Proxy
    Duration time.Duration
    Error error
}

type CheckResults []CheckResult

func (this CheckResults) Len() int {
    return len(this)
}

func (this CheckResults) Less(i, j int) bool {
    if this[j].Error != nil {
        return true
    }

    if this[i].Error != nil {
        return false
    }

    return this[i].Duration < this[j].Duration
}

func (this CheckResults) Swap(i, j int) {
    this[i], this[j] = this[j], this[i]
}

func CheckAll(concurrency int, options map[int]interface{}, url string, proxies []Proxy) CheckResults {
    var mu sync.Mutex

    var checkresults CheckResults
    chs := make([]chan bool, concurrency)
    for i := 0; i < concurrency; i++ {
        chs[i] = make(chan bool)
        go func(ch chan bool) {
            for {
                if len(proxies) == 0 {
                    break
                }
                var proxy Proxy
                mu.Lock()
                if len(proxies) > 0 {
                    proxy = proxies[0]
                    proxies = proxies[1:]
                }
                mu.Unlock()

                if proxy.Host == "" {
                    break
                }

                duration, err := Check(options, url, proxy)
                if err != nil {
                    log.Println(err.Error())
                } else {
                    mu.Lock()
                    checkresults = append(checkresults, CheckResult {
                        Proxy: proxy,
                        Duration: duration,
                        Error: err,
                    })
                    mu.Unlock()
                }
            }
            ch <- true
        }(chs[i])
    }
    
    for i := 0; i < len(chs); i++ {
        <- chs[i]
    }

    sort.Sort(checkresults) 

    return checkresults
}