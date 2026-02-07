package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Input struct {
	Domain     string `json:"domain"`
	RecordType string `json:"record_type"`
	DNSServer  string `json:"dns_server"`
}

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding JSON input: %v\n", err)
		os.Exit(1)
	}

	args := []string{"doggo", "--json"}

	if input.DNSServer != "" {
		args = append(args, "@"+input.DNSServer)
	}

	args = append(args, input.Domain)

	if input.RecordType != "" {
		args = append(args, input.RecordType)
	}

	cmd := exec.Command(args[0], args[1:]...)

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running doggo: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(out))
}
