package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func NewCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "❌ Usage: htgo new <project-name>")
		os.Exit(1)
	}

	projectName := args[0]
	if err := createProject(projectName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create project: %v\n", err)
		os.Exit(1)
	}
}

func createProject(name string) error {
	projectDir := name

	if err := os.Mkdir(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("📁 Creating project structure...\n")

	dirs := []string{
		"pages",
		"cmd/dev",
		"cmd/build",
		"cmd/app",
		".htgo",
	}

	for _, dir := range dirs {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	fmt.Printf("📝 Creating files...\n")

	files := map[string]string{
		filepath.Join(projectDir, ".htgo/keep"):           "",
		filepath.Join(projectDir, ".htgo/favicon.svg"):    faviconTemplate,
		filepath.Join(projectDir, "app.go"):              appGoTemplate,
		filepath.Join(projectDir, "cmd/dev/main.go"):     devCmdTemplate,
		filepath.Join(projectDir, "cmd/build/main.go"):   buildCmdTemplate,
		filepath.Join(projectDir, "cmd/app/main.go"):     appCmdTemplate,
		filepath.Join(projectDir, "pages/index.tsx"):     indexPageTemplate,
		filepath.Join(projectDir, "styles.css"):          stylesCssTemplate,
		filepath.Join(projectDir, "Makefile"):            makefileTemplate,
		filepath.Join(projectDir, "go.mod"):              goModTemplate,
		filepath.Join(projectDir, "tsconfig.json"):       tsconfigTemplate,
		filepath.Join(projectDir, "package.json"):        packageJsonTemplate,
		filepath.Join(projectDir, ".gitignore"):          gitignoreTemplate,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("✓ Project '%s' created successfully!\n", name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\n")
	fmt.Printf("🚀 Next steps:\n\n")
	fmt.Printf("  1. Navigate to the project:\n")
	fmt.Printf("     cd %s\n\n", name)
	fmt.Printf("  2. Install dependencies:\n")
	fmt.Printf("     make install\n\n")
	fmt.Printf("  3. Start development:\n")
	fmt.Printf("     make dev\n\n")
	fmt.Printf("  4. Open your browser:\n")
	fmt.Printf("     http://localhost:8080\n\n")
	fmt.Printf("Happy coding! 🎉\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return nil
}

const appGoTemplate = `package app

import (
	"embed"

	"github.com/bertilxi/htgo"
)

//go:embed .htgo
var EmbedFS embed.FS

var Options = htgo.Options{
	EmbedFS: &EmbedFS,
	Title:   "My HTGO App",
	Pages: []htgo.Page{
		{
			Route:       "/",
			File:        "pages/index.tsx",
			Interactive: true,
		},
	},
}
`

const devCmdTemplate = `package main

import (
	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
	app "my-app"
)

func main() {
	cli.Dev(htgo.New(app.Options))
}
`

const buildCmdTemplate = `package main

import (
	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
	app "my-app"
)

func main() {
	cli.Build(htgo.New(app.Options))
}
`

const appCmdTemplate = `package main

import (
	"embed"

	"github.com/bertilxi/htgo"
	app "my-app"
)

//go:embed .htgo/*
var embedFS embed.FS

func main() {
	options := app.Options
	options.EmbedFS = &embedFS
	engine := htgo.New(options)
	engine.Start()
}
`

const indexPageTemplate = `import "../styles.css";

export default function Home() {
  return (
    <main>
      <div className="flex flex-col items-center justify-center min-h-screen bg-gradient-to-b from-white to-gray-50">
        <h1 className="text-5xl font-bold text-gray-900 mb-4">
          Welcome to HTGO
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          React SSR for Go 🚀
        </p>
        <div className="space-x-4">
          <a
            href="https://github.com/bertilxi/htgo"
            className="inline-block px-6 py-3 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition"
          >
            View on GitHub
          </a>
          <button
            onClick={() => alert("Interactive! 🎉")}
            className="inline-block px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
          >
            Click Me
          </button>
        </div>
      </div>
    </main>
  );
}
`

const makefileTemplate = `install:
	go mod tidy
	npm install
	mkdir -p .htgo
	touch .htgo/keep

build:
	go run cmd/build/main.go
	HTGO_ENV=production go build -ldflags='-s -w' -o dist/app cmd/app/main.go

start:
	HTGO_ENV=production GIN_MODE=release ./dist/app

dev:
	go run cmd/dev/main.go

.PHONY: install build start dev
`

const packageJsonTemplate = `{
  "name": "my-htgo-app",
  "version": "0.1.0",
  "dependencies": {
    "react": "^19",
    "react-dom": "^19"
  },
	"devDependencies": {
    "@types/react": "^19",
    "@types/react-dom": "^19"
  }
}
`

const goModTemplate = `module my-app

go 1.23

require github.com/bertilxi/htgo v0.1.0

// For local development, uncomment and update the path:
// replace github.com/bertilxi/htgo => ../htgo
`

const stylesCssTemplate = `@import "tailwindcss";
`

const tsconfigTemplate = `{
  "$schema": "https://json.schemastore.org/tsconfig",
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "verbatimModuleSyntax": true,
    "isolatedModules": true,
    "noEmit": true,
    "forceConsistentCasingInFileNames": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "allowJs": true,
    "jsx": "react-jsx",
    "jsxImportSource": "react",
    "strict": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./*"]
    }
  },
  "exclude": ["dist"],
  "include": ["./**/*"]
}
`

const gitignoreTemplate = `.htgo/
.htgo-cache/
dist/
tmp/
node_modules/
*.ssr.js
*.o
*.exe
.DS_Store
go.sum
`

const faviconTemplate = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><path fill="#553986" d="M26 31h4v4h-4zM6 31h4v4H6zm24-21h-2V8h-2V6h-3V2h-2v4h-6V2h-2v4h-3v2H8v2H6v7H2v2h4v7h4v5h5v-5h6v5h5v-5h4v-7h4v-2h-4v-7zM16 21h-4v-8h4v8zm4 0v-8h4v8h-4zM34 6h2v11h-2zM0 6h2v11H0z"/></svg>`
