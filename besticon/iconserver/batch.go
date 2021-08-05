package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mat/besticon/besticon"
)

type iconResp struct {
	Url     string `json:"url"`
	Favicon string `json:"favicon"`
	Title   string `json:"title"`
}

func iconSingleHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	var withTitle bool
	if wt := r.FormValue("with_title"); wt == "true" {
		withTitle = true
	}

	result, err := parseUrlsInBatch([]string{url}, withTitle)
	if err != nil || len(result) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	bs, _ := json.Marshal(result[0])
	w.Write(bs)

	return
}

func iconBatchHandler(w http.ResponseWriter, r *http.Request) {
	var urls []string
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&urls); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var withTitle bool
	if wt := r.FormValue("with_title"); wt == "true" {
		withTitle = true
	}

	result, err := parseUrlsInBatch(urls, withTitle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		bs, _ := json.Marshal(result)
		w.Write(bs)
	}

	return
}

func parseUrlsInBatch(urls []string, withTitle bool) (resp []*iconResp, err error) {
	job := newJob(urls, withTitle)
	job.process()
	resp = job.result
	return
}

type job struct {
	withTitle  bool
	urls       []string
	urlChan    chan string
	resultChan chan *iconResp
	workerNum  int
	close      chan bool
	result     []*iconResp
}

func newJob(urls []string, withTitle bool) *job {
	job := &job{
		withTitle:  withTitle,
		urls:       urls,
		urlChan:    make(chan string),
		resultChan: make(chan *iconResp),
		close:      make(chan bool),
	}

	if len(urls) < 8 {
		job.workerNum = len(urls)
	} else if len(urls) < 16 {
		job.workerNum = len(urls) / 2
	} else {
		job.workerNum = 8
	}

	return job
}

func (job *job) process() {
	for i := 0; i < job.workerNum; i++ {
		go job.spawnWorker(i)
	}

	go job.dispatch()

	for resp := range job.resultChan {
		job.result = append(job.result, resp)

		if len(job.result) == len(job.urls) {
			break
		}
	}

	for i := 0; i < job.workerNum; i++ {
		job.close <- true
	}
}

func (job *job) dispatch() {
	for _, url := range job.urls {
		url = strings.TrimSpace(url)
		if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
			url = "http://" + url
		}

		job.urlChan <- url
	}
}

func (job *job) spawnWorker(num int) {
	for {
		select {
		case url := <-job.urlChan:
			favicon, title := parseUrl(url, job.withTitle)
			job.resultChan <- &iconResp{
				Url:     url,
				Favicon: favicon,
				Title:   title,
			}
		case <-job.close:
			return
		}
	}
}

func parseUrl(url string, withTitle bool) (favicon, title string) {
	var cached bool
	if withTitle {
		if title, favicon, cached = readTitle(url); cached {
			return
		}
	} else {
		if favicon, cached = readFavicon(getHost(url)); cached {
			return
		}
	}

	return getFaviconAndTitle(url)
}

func getFaviconAndTitle(url string) (favicon, title string) {
	finder := newIconFinder()

	var err error
	var icon *besticon.Icon
	if title, icon, err = finder.FetchIconsWithTitle(url); err != nil {
		log.Println(err)
		return
	}

	if icon != nil {
		favicon = icon.URL
	}

	go update(url, favicon, title)
	return
}

func getHost(rawurl string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Println(err)
		return ""
	}

	return u.Hostname()
}
