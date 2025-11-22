package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/tabwriter"

	"go.aimuz.me/mynt/zfs"
)

const defaultAddr = "http://localhost:8080"

func main() {
	addr := flag.String("addr", defaultAddr, "Address of myntd")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	switch args[0] {
	case "pool":
		handlePool(args[1:], *addr)
	case "dataset":
		handleDataset(args[1:], *addr)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: mynt [flags] <command> [subcommand]")
	fmt.Println("Commands:")
	fmt.Println("  pool list")
	fmt.Println("  dataset list")
}

func handlePool(args []string, addr string) {
	if len(args) < 1 || args[0] != "list" {
		fmt.Println("Usage: mynt pool list")
		return
	}

	resp, err := http.Get(addr + "/api/v1/pools")
	if err != nil {
		log.Fatalf("Failed to connect to myntd: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Error: %s", body)
	}

	var pools []zfs.Pool
	if err := json.NewDecoder(resp.Body).Decode(&pools); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSIZE\tALLOC\tFREE\tHEALTH")
	for _, p := range pools {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\n", p.Name, p.Size, p.Allocated, p.Free, p.Health)
	}
	w.Flush()
}

func handleDataset(args []string, addr string) {
	if len(args) < 1 || args[0] != "list" {
		fmt.Println("Usage: mynt dataset list")
		return
	}

	resp, err := http.Get(addr + "/api/v1/datasets")
	if err != nil {
		log.Fatalf("Failed to connect to myntd: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Error: %s", body)
	}

	var datasets []zfs.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&datasets); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tUSED\tAVAIL\tMOUNTPOINT")
	for _, d := range datasets {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n", d.Name, d.Type, d.Used, d.Available, d.Mountpoint)
	}
	w.Flush()
}
