package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "config":
		if err := runConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		printUsage()
	default:
		if err := runSend(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runConfig() error {
	fmt.Print("Enter Slack token: ")
	tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	fmt.Print("Enter default recipient (username or leave blank to skip): ")
	reader := bufio.NewReader(os.Stdin)
	defaultRecipient, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading recipient: %w", err)
	}
	defaultRecipient = strings.TrimSpace(defaultRecipient)

	cfg := &Config{
		Token:            token,
		DefaultRecipient: defaultRecipient,
	}
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	path, _ := configPath()
	fmt.Printf("Config saved to %s\n", path)
	return nil
}

func runSend(args []string) error {
	if len(args) > 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var recipient, message string
	switch len(args) {
	case 1:
		message = args[0]
		recipient = cfg.DefaultRecipient
		if recipient == "" {
			return fmt.Errorf("no default recipient set — run: notify config\n       or specify one: notify <username> \"<message>\"")
		}
	case 2:
		recipient = args[0]
		message = args[1]
	}

	client := newSlackClient(cfg.Token)
	if err := client.SendDM(recipient, message); err != nil {
		return err
	}

	fmt.Printf("Sent to @%s\n", strings.TrimPrefix(recipient, "@"))
	return nil
}

func printUsage() {
	fmt.Print(`notify — send a Slack DM from the terminal

Usage:
  notify config                     Interactive setup (token + default recipient)
  notify "<message>"                Send to default recipient
  notify <username> "<message>"     Send to a specific user

Examples:
  notify config
  notify "Build complete"
  notify sveinungkb "Build complete"
  make build && notify "Build done" || notify "Build FAILED"
`)
}
