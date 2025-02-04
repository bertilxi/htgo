package htgo

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var mutex sync.Mutex
var hotReloadWs *websocket.Conn
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func broadcastFileUpdateToClients() {
	if hotReloadWs == nil {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	hotReloadWs.WriteMessage(1, []byte("reload"))
}

func startWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to start watcher", err)
		return
	}
	defer watcher.Close()

	if err = filepath.Walk(cacheDir, func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		log.Fatal("Failed to add files in directory to watcher", err)
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op.String() != "CHMOD" && !strings.Contains(event.Name, ".tmp.") {
				broadcastFileUpdateToClients()
			}

		case err := <-watcher.Errors:
			log.Fatal("Error watching files", err)
		}
	}
}

func websocketHandler(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		log.Fatal("Failed to upgrade websocket", err)
		return
	}

	hotReloadWs = ws
}

func watchServer(page Page) {
	ctx, err := esbuild.Context(backendOptions(page.File))
	if err != nil {
		log.Fatal(err)
	}

	err2 := ctx.Watch(esbuild.WatchOptions{})
	if err2 != nil {
		log.Fatal(err)
	}
}

func watchClient(page Page) {
	ctx, err := esbuild.Context(clientOptions(page.File))
	if err != nil {
		log.Fatal(err)
	}

	err2 := ctx.Watch(esbuild.WatchOptions{})
	if err2 != nil {
		log.Fatal(err)
	}
}
