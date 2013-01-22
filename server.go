package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
)



type Request struct {
	ReqCh chan *request
	ResCh chan map[string][]string
}

func (req *Request) Run() {
	for {
		select {
		case r := <- req.ReqCh:
			r.ch <- GenPkgList(r.root, r.pkg)
			close(r.ch)
		}
	}
}

type request struct {
	root    string
	pkg     string
	ch      chan interface{}
}

func (req *Request) GetPkgList(root string, pkg string) map[string][]string{
	r := &request{
		root: root,
		pkg: pkg,
		ch: make(chan interface{}, 1),
	}
	req.ReqCh <- r
	result := <-r.ch
	return result.(map[string][]string)
}

func GenPkgList(root string, pkg string) map[string][]string {
	pkgMap := make(map[string][]string)
	if root != "" {
		list := []string{}

		pf := func (path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, pkg) {
				list = append(list,path)
			}
			return err
		}

		err := filepath.Walk(root, pf)
		if err != nil {
			log.Printf("Filepath.Walk() returned %v\n", err)
		}
		pkgMap[root] = list
	}
	return pkgMap
}

func FilterRpmList(rpmlist []string, filter string) []string{
	filteredList := []string{}
	for f := range rpmlist {
		if strings.Contains(rpmlist[f], filter) {
			filteredList = append(filteredList, rpmlist[f])
		}
	}
	return filteredList
}

func NewRequest() *Request {
	return  &Request{
		ReqCh: make(chan *request),
		ResCh: make(chan map[string][]string),
	}
}


func main() {
	DefaultRequest := NewRequest()
	go DefaultRequest.Run()

	getRpmList := func(w http.ResponseWriter, req *http.Request) {
		root := req.URL.Query().Get("distro")
		if root == "" {
			fmt.Fprintf(w,"getall: empty list\n")
		}
		list := DefaultRequest.GetPkgList(root, "rpm")

		filter := ""
		if req.URL.Query().Get("filter") != ""  {
		 	filter = req.URL.Query().Get("filter")
		 	fl := FilterRpmList(list[root], filter)
		 	list[root] = fl
		}

		b, err := json.Marshal(list)
		if err != nil {
		  	fmt.Fprintf(w,"getall: empty list")
		}
		fmt.Fprintf(w, "%s", b)
	}

	http.HandleFunc("/list", getRpmList)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
