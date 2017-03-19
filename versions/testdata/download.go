package main

import (
	"net/http"
	"os"
	"io"
	"log"
	"github.com/PhilipBorgesen/minecraft/versions/internal"
)

func main() {
	resp, err := http.Get(internal.VersionsURL)
	if err != nil {
		log.Fatalf("download failed: %s", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
}
