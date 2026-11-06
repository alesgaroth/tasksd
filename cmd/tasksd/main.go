package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alesgaroth/tasksd/internal/store"
	"github.com/alesgaroth/tasksd/internal/web"
)

func runValidate(args []string) int {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}

	errs, count, err := store.ValidatePath(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", target, err)
		return 1
	}

	for _, e := range errs {
		fmt.Fprintln(os.Stderr, e)
	}

	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation failed: %d error(s) found in %d item(s).\n", len(errs), count)
		return 1
	}

	fmt.Printf("Validated %d item(s). All items valid.\n", count)
	return 0
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		os.Exit(runValidate(os.Args[2:]))
	}

	repoDir := flag.String("repo", ".", "path to task data repository")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	handler := web.NewServerWithRepo(*repoDir)
	log.Printf("Starting tasksd on %s (repo: %s)", *addr, *repoDir)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
