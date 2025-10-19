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
		"pages/api",
		".htgo",
		"cmd/dev",
		"cmd/build",
		"cmd/app",
	}

	for _, dir := range dirs {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	fmt.Printf("📝 Creating files...\n")

	files := map[string]string{
		filepath.Join(projectDir, ".htgo/keep"):                  "",
		filepath.Join(projectDir, ".htgo/favicon.svg"):           faviconTemplate,
		filepath.Join(projectDir, "app.go"):                     appGoTemplate,
		filepath.Join(projectDir, "pages/generate.go"):          pagesGenerateTemplate,
		filepath.Join(projectDir, "pages/index.tsx"):            indexPageTemplate,
		filepath.Join(projectDir, "pages/index.go"):             indexLoaderTemplate,
		filepath.Join(projectDir, "pages/loaders_generated.go"): pagesLoadersGeneratedTemplate,
		filepath.Join(projectDir, "pages/api/hello.go"):         apiHelloTemplate,
		filepath.Join(projectDir, "styles.css"):                 stylesCssTemplate,
		filepath.Join(projectDir, "go.mod"):                     goModTemplate,
		filepath.Join(projectDir, "tsconfig.json"):              tsconfigTemplate,
		filepath.Join(projectDir, "package.json"):               packageJsonTemplate,
		filepath.Join(projectDir, ".gitignore"):                 gitignoreTemplate,
		filepath.Join(projectDir, "cmd/dev/main.go"):            devMainTemplate,
		filepath.Join(projectDir, "cmd/build/main.go"):          buildMainTemplate,
		filepath.Join(projectDir, "cmd/app/main.go"):            appMainTemplate,
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
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  htgo install     # Install dependencies\n")
	fmt.Printf("  htgo dev         # Start development\n")
	fmt.Printf("  htgo build       # Build for production\n")
	fmt.Printf("  ./dist/app       # Run production binary\n\n")
	fmt.Printf("Open your browser at http://localhost:8080\n\n")
	fmt.Printf("Happy coding! 🎉\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return nil
}

const appGoTemplate = `package app

import (
	"embed"

	"github.com/bertilxi/htgo"
	"my-app/pages"
)

//go:embed .htgo
var EmbedFS embed.FS

var Options = htgo.Options{
	EmbedFS:  &EmbedFS,
	PagesDir: "./pages",
	Title:    "My HTGO App",
	Loaders:  pages.LoaderRegistry,
	Handlers: pages.HandlerRegistry,
}
`

const indexLoaderTemplate = `package pages

import (
	"github.com/gin-gonic/gin"
)

// LoadIndex provides server-side data to pages/index.tsx
// This function is auto-registered via HandlerRegistry in pages/loaders_generated.go
func LoadIndex(c *gin.Context) (any, error) {
	return map[string]any{
		"message": "Hello from HTGO! 🚀",
	}, nil
}
`

const indexPageTemplate = `import "../styles.css";

interface Props {
  message: string;
}

export default function Home(props: Props) {
  return (
    <main>
      <div className="flex flex-col items-center justify-center min-h-screen bg-gradient-to-b from-white to-gray-50">
        <h1 className="text-5xl font-bold text-gray-900 mb-4">
          Welcome to HTGO
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          {props.message}
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

const devMainTemplate = `package main

import (
	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
	app "my-app"
)

func main() {
	cli.Dev(htgo.New(app.Options))
}
`

const buildMainTemplate = `package main

import (
	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
	app "my-app"
)

func main() {
	cli.Build(htgo.New(app.Options))
}
`

const appMainTemplate = `package main

import (
	"github.com/bertilxi/htgo"
	app "my-app"
)

func main() {
	engine := htgo.New(app.Options)
	engine.Start()
}
`

const pagesGenerateTemplate = `package pages

//go:generate go run github.com/bertilxi/htgo/cmd/htgo-gen-loaders .
`

const pagesLoadersGeneratedTemplate = `// Code generated by htgo-gen-loaders. DO NOT EDIT.
// This file is auto-generated by running: go generate ./pages

package pages

import (
	"github.com/bertilxi/htgo"
	api "my-app/pages/api"
)

// LoaderRegistry maps page routes to their corresponding loader functions.
// Loaders return (any, error) and their data is used as props for SSR.
var LoaderRegistry = map[string]htgo.PageLoader{
	"/": LoadIndex,
}

// HandlerRegistry maps API routes to their corresponding handler functions.
// Handlers have full Gin API control - use c.JSON(), c.File(), etc. directly.
var HandlerRegistry = map[string]htgo.Handler{
	"/api/hello": api.Hello,
}
`

const apiHelloTemplate = `package api

import (
	"github.com/gin-gonic/gin"
)

// Hello is a sample API handler that responds to /api/hello
// API handlers have full control over the Gin context and response
// Try it: curl http://localhost:8080/api/hello?name=Alice
func Hello(c *gin.Context) error {
	name := c.DefaultQuery("name", "World")
	c.JSON(200, gin.H{
		"message": "Hello, " + name + "!",
		"status":  "ok",
	})
	return nil
}
`
