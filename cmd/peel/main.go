package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/coffee-cup/peel/internal/image"
	"github.com/coffee-cup/peel/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	dev := flag.Bool("dev", false, "development mode")
	noOpen := flag.Bool("no-open", false, "don't auto-open browser")
	platform := flag.String("platform", "", "target platform os/arch (default: host)")
	flag.Parse()

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
