package main

import (
	"flag"
	"log"
	"runtime"
	"loop_hole/web"
)

func main() {
	var webAddr string
	var isDebug bool
	flag.StringVar(&webAddr, "web", ":9000", "Web address server listening on")
	flag.BoolVar(&isDebug, "debug", false, "Debug mode")
	flag.Parse()

	web.ServerInstance = web.NewServer(isDebug)
	log.Fatal(web.ServerInstance.ListenAndServe(webAddr))
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
