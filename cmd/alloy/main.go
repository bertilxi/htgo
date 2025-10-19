package main

import (
	"fmt"
	"os"

	"github.com/bertilxi/alloy/cmd/alloy/commands"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "dev":
		commands.DevCmd(os.Args[2:])
	case "build":
		commands.BuildCmd(os.Args[2:])
	case "install":
		commands.InstallCmd(os.Args[2:])
	case "new":
		commands.NewCmd(os.Args[2:])
	case "version":
		fmt.Printf("alloy version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Alloy - React SSR for Go
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

USAGE:
  alloy <command> [options]

COMMANDS:
  install          Install project dependencies
                   Usage: alloy install [--dir .]

  dev              Start development server with hot-reload
                   Usage: alloy dev [--port 8080] [--dir .]

  build            Build for production
                   Usage: alloy build [--dir .] [--output ./dist/app]

  new              Create a new Alloy project
                   Usage: alloy new <project-name>

  version          Print version

  help             Show this help message

OPTIONS:
  --port <number>  Port for server (default: 8080)
  --dir <path>     Project directory (default: current directory)
  --output <path>  Output binary path for build

EXAMPLES:
  # Create a new project
  alloy new my-app
  cd my-app

  # Install dependencies
  alloy install

  # Start dev server
  alloy dev

  # Start on custom port
  alloy dev --port 3000

  # Build for production
  alloy build

  # Run production binary
  ./dist/app

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
}
