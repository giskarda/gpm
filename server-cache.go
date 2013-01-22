package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import _ "net/http/pprof"

import "github.com/pmylund/go-cache"

type Request struct {
	ReqCh chan *request
	ResCh chan map[string][]string
}

func (req *Request) Run(c *cache.Cache) {
	for {
		select {
		case r := <-req.ReqCh:
			key := filepath.Base(r.root)
			res, _ := c.Get(key)
			r.ch <- res
			close(r.ch)
		}
	}
}

type request struct {
	root string
	pkg  string
	ch   chan interface{}
}

func (req *Request) GetPkgList(root string, pkg string) []string {
	r := &request{
		root: root,
		pkg:  pkg,
		ch:   make(chan interface{}, 1),
	}
	req.ReqCh <- r
	result := <-r.ch

	return result.([]string)
}

func GenPkgList(root string, pkg string, c *cache.Cache) {
	for {
		list := []string{}

		pf := func(path string, f os.FileInfo, err error) error {
			if strings.HasSuffix(path, pkg) {
				list = append(list, path)
			}
			return err
		}

		err := filepath.Walk(root, pf)
		if err != nil {
			log.Printf("Filepath.Walk() returned %v\n", err)
		}

		key := filepath.Base(root)
		err = c.Replace(key, list, 0)
		if err != nil {
			fmt.Printf("Tried to replace %s, fallback to Set\n", err)
			c.Set(key, list, 0)
		} else {
			fmt.Printf("Replaced %s\n", key)
		}
		time.Sleep(10*time.Second)
	}
}

func FilterRpmList(rpmlist []string, filter string) []string {
	filteredList := []string{}
	for f := range rpmlist {
		if strings.Contains(rpmlist[f], filter) {
			filteredList = append(filteredList, rpmlist[f])
		}
	}
	return filteredList
}

func NewRequest() *Request {
	return &Request{
		ReqCh: make(chan *request),
		ResCh: make(chan map[string][]string),
	}
}

func main() {
	DefaultRequest := NewRequest()
	c := cache.New(time.Minute, 30*time.Second)

	go DefaultRequest.Run(c)

	go GenPkgList("/web/futuretrain/gpm/test/devel", "rpm", c)

	getRpmList := func(w http.ResponseWriter, req *http.Request) {
		root := req.URL.Query().Get("distro")
		req.Body.Close()
		if root == "" {
			fmt.Fprintf(w, "getall: empty list\n")
		}

		list := DefaultRequest.GetPkgList(root, "rpm")

		filter := ""
		if req.URL.Query().Get("filter") != "" {
			filter = req.URL.Query().Get("filter")
			fl := FilterRpmList(list, filter)
			list = fl
		}

		b, err := json.Marshal(list)
		if err != nil {
			fmt.Fprintf(w, "getall: empty list")
		}

		w.Write(b)
	}

	http.HandleFunc("/list", getRpmList)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
