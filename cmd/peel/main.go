package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/coffee-cup/peel/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	dev := flag.Bool("dev", false, "development mode")
	noOpen := flag.Bool("no-open", false, "don't auto-open browser")
	flag.Parse()

	srv := server.New(*dev)

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	if !*noOpen && !*dev {
		go openBrowser(url)
	}

	log.Printf("listening on %s", url)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}
