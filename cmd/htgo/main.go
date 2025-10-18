package main

import (
	"fmt"
	"os"

	"github.com/bertilxi/htgo/cmd/htgo/commands"
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
	case "new":
		commands.NewCmd(os.Args[2:])
	case "version":
		fmt.Printf("htgo version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  HTGO - React SSR for Go
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

USAGE:
  htgo <command> [options]

COMMANDS:
  dev              Start development server with hot-reload
                   Usage: htgo dev [--port 8080] [--dir .]

  build            Build for production
                   Usage: htgo build [--dir .] [--output ./dist/app]

  new              Create a new HTGO project
                   Usage: htgo new <project-name>

  version          Print version

  help             Show this help message

OPTIONS:
  --port <number>  Port for dev server (default: 8080)
  --dir <path>     Project directory (default: current directory)
  --output <path>  Output binary path for build

EXAMPLES:
  # Create a new project
  htgo new my-app
  cd my-app

  # Start dev server
  htgo dev

  # Start on custom port
  htgo dev --port 3000

  # Build for production
  htgo build

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
}
