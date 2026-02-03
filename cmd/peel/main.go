package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/coffee-cup/peel/internal/image"
	"github.com/coffee-cup/peel/internal/server"
	flag "github.com/spf13/pflag"
	"golang.org/x/term"
)

var version = "dev"

func isTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func main() {
	flag.Usage = usage

	showVersion := flag.BoolP("version", "v", false, "print version and exit")
	port := flag.IntP("port", "p", 0, "port to listen on")
	noOpen := flag.Bool("no-open", false, "don't auto-open browser")
	platform := flag.String("platform", "", "target platform os/arch")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	ref := flag.Arg(0)
	if ref == "" {
		usage()
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

	srv := server.New(analyzed)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)

	if !*noOpen {
		go openBrowser(url)
	}

	log.Printf("listening on %s", url)
	if err := http.Serve(ln, srv); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	tty := isTTY()

	bold := func(s string) string {
		if tty {
			return "\033[1m" + s + "\033[0m"
		}
		return s
	}
	cyan := func(s string) string {
		if tty {
			return "\033[36m" + s + "\033[0m"
		}
		return s
	}
	dim := func(s string) string {
		if tty {
			return "\033[2m" + s + "\033[0m"
		}
		return s
	}

	fmt.Fprintf(os.Stderr, "%s\n\n", bold("peel")+" — container image inspector")
	fmt.Fprintf(os.Stderr, "%s\n", bold("Usage:"))
	fmt.Fprintf(os.Stderr, "  peel <image> [flags]\n\n")
	fmt.Fprintf(os.Stderr, "%s\n", bold("Flags:"))
	fmt.Fprintf(os.Stderr, "  %s, %s         %s\n", cyan("-p"), cyan("--port"), "port to listen on "+dim("(int, default random)"))
	fmt.Fprintf(os.Stderr, "      %s     %s\n", cyan("--platform"), "target platform os/arch "+dim("(e.g. linux/amd64)"))
	fmt.Fprintf(os.Stderr, "      %s      %s\n", cyan("--no-open"), "don't auto-open browser")
	fmt.Fprintf(os.Stderr, "  %s, %s      %s\n", cyan("-v"), cyan("--version"), "print version and exit")
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
