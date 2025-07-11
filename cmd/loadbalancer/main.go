package main

import (
	"flag"
	"log"
	"strings"

	"loadbalancer/internal/strategies/roundrobin"
)

func main() {
	var servers string
	var port int

	flag.StringVar(&servers, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.Parse()

	serverList := strings.Split(servers, ",")
	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	roundrobin.StartRR(serverList, port)
}
