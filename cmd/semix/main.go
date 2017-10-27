package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"
)

const (
	statusOK int = iota
	statusInvalidCommand
	statusCommandError
)

var listCommands bool
var commands = map[string][]string{
	"httpd":  {"semix-httpd", "run semix http daemon"},
	"daemon": {"semix-restd", "run semix REST API daemon"},
}

func init() {
	flag.BoolVar(&listCommands, "list", false, "list available commands")
}

func main() {
	flag.Parse()
	if listCommands {
		printCommands()
	} else {
		runCommand(flag.Args())
	}
}

func runCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing command\n")
		os.Exit(statusInvalidCommand)
	}
	if _, ok := commands[args[0]]; !ok {
		fmt.Fprintf(os.Stderr, "invalid command\n")
		os.Exit(statusInvalidCommand)
	}
	cmd := exec.Command(commands[args[0]][0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running command: %s\n", commands[args[0]][0])
		os.Exit(statusCommandError)
	}
	fmt.Printf("args: %v", args)
}

func printCommands() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Available commands:\n")
	for command, args := range commands {
		fmt.Fprintf(w, "%s\t%s\t%s\n", command, args[0], args[1])
	}
	w.Flush()
}
