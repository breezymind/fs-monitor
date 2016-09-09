package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	PATH string
)

func init() {
	flag.StringVar(&PATH, "path", "", "monitor target path")
	flag.Parse()
}

func watcher(watchwg *sync.WaitGroup, watchdir chan string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	for {
		select {
		case req := <-watchdir:
			err := watcher.Add(req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(req)
		}
	}

	watchwg.Done()
}

func finddir(wg *sync.WaitGroup, reqdir, resdir chan string) {
	for {
		select {
		case dir := <-reqdir:
			go func() {
				nodes, err := ioutil.ReadDir(dir)
				if err != nil {
					log.Fatal(err)
				}
				for _, node := range nodes {
					nodename := node.Name()
					nodepath := dir + "/" + nodename
					if node.IsDir() {
						resdir <- nodepath
					}
				}
				wg.Done()
			}()
		}
	}
}

// https://kldp.org/node/92783
// https://github.com/fsnotify/fsnotify/wiki/FAQ

func main() {

	if PATH == "" {
		fmt.Println("usage: fs-monitor -path=/Users/breezymind/Cert")
		os.Exit(0)
	}
	var (
		maxcpu   = runtime.NumCPU()
		dirwg    = sync.WaitGroup{}
		watchwg  = sync.WaitGroup{}
		reqdir   = make(chan string)
		resdir   = make(chan string)
		watchdir = make(chan string)
	)

	runtime.GOMAXPROCS(maxcpu)

	for i := 0; i < maxcpu; i++ {
		fmt.Println(fmt.Sprintf("init finder #%d", i))
		go finddir(&dirwg, reqdir, resdir)
	}
	go func() {
		for {
			select {
			case res := <-resdir:
				watchdir <- res
				dirwg.Add(1)
				reqdir <- res
			}
		}
	}()
	go func() {
		dirwg.Add(1)
		reqdir <- PATH
		watchdir <- PATH
		dirwg.Wait()
	}()

	watchwg.Add(1)
	for i := 0; i < maxcpu; i++ {
		fmt.Println(fmt.Sprintf("init watcher #%d", i))
		go watcher(&watchwg, watchdir)
	}
	watchwg.Wait()
}
