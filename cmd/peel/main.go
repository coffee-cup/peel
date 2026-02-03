package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/coffee-cup/peel/internal/image"
	"github.com/coffee-cup/peel/internal/server"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	port := flag.Int("port", 8080, "port to listen on")
	dev := flag.Bool("dev", false, "development mode")
	noOpen := flag.Bool("no-open", false, "don't auto-open browser")
	platform := flag.String("platform", "", "target platform os/arch (default: host)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	ref := flag.Arg(0)
	if ref == "" {
		fmt.Fprintf(os.Stderr, "usage: peel <image-reference> [flags]\n")
		os.Exit(1)
	}

	plat, err := image.ParsePlatform(*platform)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("loading %s (%s/%s)", ref, plat.OS, plat.Architecture)
	img, err := image.LoadImage(ref, plat)
	if err != nil {
		log.Fatal(err)
	}

	analyzed, err := image.Analyze(img, ref)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("analyzed %d layers", analyzed.Info.LayerCount)

	srv := server.New(*dev, analyzed)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal(err)
		}
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)

	if !*noOpen && !*dev {
		go openBrowser(url)
	}

	log.Printf("listening on %s", url)
	if err := http.Serve(ln, srv); err != nil {
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
