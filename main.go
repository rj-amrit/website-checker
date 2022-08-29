package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	fmt.Println("starting server")
	http.HandleFunc("/", DefaultHandler)
	http.HandleFunc("/websites", WebsiteHandler)
	go StatusCheckerPerMin()
	http.ListenAndServe("127.0.0.1:8080", nil)
}

var sites map[string]string

type StatusChecker interface {
	Check(name string) (status bool)
}

type HttpChecker struct {
}

func (h HttpChecker) Check(name string) (status bool) {
	resp, err := http.Get("https://" + name)
	if err == nil && resp.StatusCode == 200 {
		status = true
		return
	}
	status = false
	return
}

var checker = HttpChecker{}

type SiteStruct struct {
	Website []string `json:"website"`
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is the default page")
}

func WebsiteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		// GET requests:
		// curl 127.0.0.1:8080/websites
		// curl 127.0.0.1:8080/websites?name=www.google.com
		GetHandler(w, r)
	case "POST":
		// POST request:
		// curl -d '{"""website""": ["""www.google.com""", """www.facebook.com""", """www.fakewebsite1.com"""]}' -X POST 127.0.0.1:8080/websites
		PostHandler(w, r)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	sites = map[string]string{}
	site := SiteStruct{}
	err := json.NewDecoder(r.Body).Decode(&site)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, j := range site.Website {
		sites[j] = "UNCHECKED"
	}
	fmt.Fprint(w, "200 OK")
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if len(params) == 0 {
		fmt.Fprint(w, sites)
	} else {
		website := r.URL.Query().Get("name")
		if val, ok := sites[website]; ok {
			fmt.Fprint(w, val)
		}
	}
}

func StatusCheckerPerMin() {
	wg := sync.WaitGroup{}
	for {
		for website := range sites {
			wg.Add(1)
			go func(website string) {
				defer wg.Done()
				if checker.Check(website) {
					sites[website] = "UP"
				} else {
					sites[website] = "DOWN"
				}
			}(website)
		}
		wg.Wait()
		fmt.Printf("checked %v websites\n", len(sites))
		time.Sleep(time.Minute)
	}
}
