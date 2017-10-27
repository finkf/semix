package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/tabwriter"
)

var (
	list     bool
	help     bool
	commands = map[string][]string{
		"httpd": {"semix-httpd", "run semix http daemon"},
		"restd": {"semix-restd", "run semix REST API daemon"},
	}
)

func init() {
	flag.BoolVar(&list, "list", false, "list available commands")
	flag.BoolVar(&help, "help", false, "prints this help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	if list {
		printCommands()
	} else {
		runCommand(flag.Args())
	}
}

func runCommand(args []string) {
	if len(args) == 0 {
		log.Fatalf("missing command")
	}
	if _, ok := commands[args[0]]; !ok {
		log.Fatalf("invalid command: %s", args[0])
	}
	cmd := exec.Command(commands[args[0]][0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("error running command %s: %v", commands[args[0]][0], err)
	}
}

func printCommands() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Available commands:\n")
	for command, args := range commands {
		fmt.Fprintf(w, "%s\t%s\t%s\n", command, args[0], args[1])
	}
	w.Flush()
}
