package cli

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bertilxi/htgo"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type hotReload struct {
	mutex    sync.Mutex
	ws       *websocket.Conn
	upgrader websocket.Upgrader
}

func newHotReload() *hotReload {
	return &hotReload{
		mutex: sync.Mutex{},
		ws:    nil,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (hr *hotReload) reload() {
	if hr.ws == nil {
		return
	}

	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	hr.ws.WriteMessage(1, []byte("reload"))
}

func (hr *hotReload) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = filepath.Walk(htgo.CacheDir, func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op.String() != "CHMOD" && !strings.Contains(event.Name, ".tmp.") {
				hr.reload()
			}

		case err := <-watcher.Errors:
			return err
		}
	}
}

func (hr *hotReload) websocket(c *gin.Context) {
	ws, err := hr.upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	hr.ws = ws
}
