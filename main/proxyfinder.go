package main

import (
    "github.com/ddliu/go-proxyfinder"
    "github.com/ddliu/go-httpclient"
    "github.com/spf13/cobra"
    "fmt"
    "os"
    "log"
    "strings"
    "io/ioutil"
)

// cli args
var (
    timeout int
    testurl string
    logfile string
    output string
    source string
    concurrency int
)

func getFinder(source string) *proxyfinder.Finder {
    f := &proxyfinder.Finder{}

    if source == "" {
        for _, h := range proxyfinder.FINDER_HANDLERS {
            f.AddHandler(nil, h)
        }
    } else {
        sources := strings.Split(source, ",")
        for _, s := range sources {
            s = strings.TrimSpace(s)

            if h, ok := proxyfinder.FINDER_HANDLERS[s]; ok {
                f.AddHandler(nil, h)
            } else {
                log.Printf("Invalid handler: %s\n", s)
            }
        }
    }

    return f
}

func main() {
    var finderCmd = &cobra.Command {
        Use: "proxyfinder",
        Short: "Find proxy servers and check for the best",
        Long: `Find proxy servers and check for the best`,
        Run: func (cmd *cobra.Command, args []string) {
            cmd.Help()
        },
    }

    finderCmd.PersistentFlags().StringVarP(&logfile, "log", "l", "./proxyfinder.log", "Specify log file")

    var runCmd = &cobra.Command {
        Use: "run",
        Short: "Run",
        Long: "Run",
        Run: func (cmd *cobra.Command, args []string) {
            finder := &proxyfinder.Finder{}

            finder.AddHandler(nil, proxyfinder.FinderCNProxy)

            proxies := finder.Find()

            total := len(proxies)

            if concurrency < 1 {
                fmt.Println("Invalid concurrency")
                return
            }

            checkresults := proxyfinder.CheckAll(concurrency, map[int]interface{} {
                httpclient.OPT_TIMEOUT: timeout,
            }, testurl, proxies)

            proxies = make([]proxyfinder.Proxy, len(checkresults))
            for k, v := range checkresults {
                proxies[k] = v.Proxy
            }

            s := proxyfinder.DumpProxyList(proxies)

            if output == "" {
                fmt.Print(s)
            } else {
                err := ioutil.WriteFile(output, []byte(s), 0666)
                if err != nil {
                    fmt.Println(err.Error())
                    log.Println(err.Error())
                } else {
                    fmt.Printf("%d of %d proxy servers available, showing top 10 results:\n", len(checkresults), total)
                    proxies := proxies[0:10]
                    fmt.Print(proxyfinder.DumpProxyList(proxies))
                }
            }
        },
    }

    runCmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Maximum execution time when check each proxy server")
    runCmd.Flags().StringVarP(&testurl, "url", "u", "http://httpbin.org/get", "Target url to test proxy servers")
    runCmd.Flags().StringVarP(&output, "output", "o", "", "Output file name, if not specified, result will be printed out")
    runCmd.Flags().StringVarP(&source, "source", "s", "", "Sources to find proxy servers")
    runCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 50, "Number of tests to run at the same time")

    var findCmd = &cobra.Command {
        Use: "find",
        Short: "Find proxy servers",
        Long: "Find proxy servers from given sources.",
        Run: func (cmd *cobra.Command, args[]string) {
            finder := getFinder(source)

            proxies := finder.Find()

            s := proxyfinder.DumpProxyList(proxies)

            if output == "" {
                fmt.Print(s)
            } else {
                err := ioutil.WriteFile(output, []byte(s), 0666)
                if err != nil {
                    fmt.Println(err.Error())
                    log.Println(err.Error())
                } else {
                    fmt.Printf("Found %d results, saved to %s\n", len(proxies), output)
                }
            }
        },
    }

    findCmd.Flags().StringVarP(&source, "source", "s", "", "Sources to find proxy servers")
    findCmd.Flags().StringVarP(&output, "output", "o", "", "Output file name, if not specified, result will be printed out")


    var checkCmd = &cobra.Command {
        Use: "check",
        Short: "Check proxy status",
        Long: `Given a list of proxy servers, find out which ones are available, and witch ones are faster.`,
        Run: func (cmd *cobra.Command, args []string) {
            if source == "" {
                cmd.Help()
                return
            }
            bs, err := ioutil.ReadFile(source)
            if err != nil {
                fmt.Println(err.Error())
                log.Println(err)
                return
            }

            proxies := proxyfinder.ParseProxyList(string(bs))
            total := len(proxies)

            if total == 0 {
                fmt.Println("No proxy to check")
                return
            }

            if concurrency < 1 {
                fmt.Println("Invalid concurrency")
                return
            }

            checkresults := proxyfinder.CheckAll(concurrency, map[int]interface{} {
                httpclient.OPT_TIMEOUT: timeout,
            }, testurl, proxies)

            proxies = make([]proxyfinder.Proxy, len(checkresults))
            for k, v := range checkresults {
                proxies[k] = v.Proxy
            }

            s := proxyfinder.DumpProxyList(proxies)

            if output == "" {
                fmt.Print(s)
            } else {
                err := ioutil.WriteFile(output, []byte(s), 0666)
                if err != nil {
                    fmt.Println(err.Error())
                    log.Println(err.Error())
                } else {
                    fmt.Printf("%d of %d proxy servers available, showing top 10 results:\n", len(checkresults), total)
                    proxies := proxies[0:10]
                    fmt.Print(proxyfinder.DumpProxyList(proxies))
                }
            }
        },
    }

    checkCmd.Flags().StringVarP(&source, "input", "i", "", "File that contains proxy list for checking")
    checkCmd.Flags().StringVarP(&output, "output", "o", "", "Output file name, if not specified, result will be printed out")
    checkCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 50, "Number of tests to run at the same time")


    finderCmd.AddCommand(runCmd, findCmd, checkCmd)


    if logfile != "" {
        f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
        if err == nil {
            defer f.Close()
            log.SetOutput(f)
        } else {
            log.Println(err.Error())
        }
    }

    finderCmd.Execute()
}