package cli

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/core"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type hotReload struct {
	mutex       sync.RWMutex
	connections map[*websocket.Conn]*sync.Mutex
	upgrader    websocket.Upgrader
}

func newHotReload() *hotReload {
	return &hotReload{
		mutex:       sync.RWMutex{},
		connections: make(map[*websocket.Conn]*sync.Mutex),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (hr *hotReload) reload() {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	if len(hr.connections) == 0 {
		return
	}

	alloy.ClearBundleCache()

	for conn, writeMutex := range hr.connections {
		go func(c *websocket.Conn, m *sync.Mutex) {
			m.Lock()
			defer m.Unlock()
			c.WriteMessage(1, []byte("reload"))
		}(conn, writeMutex)
	}
}

func (hr *hotReload) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = filepath.Walk(core.CacheDir, func(path string, fi os.FileInfo, err error) error {
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

	writeMutex := &sync.Mutex{}

	hr.mutex.Lock()
	hr.connections[ws] = writeMutex
	hr.mutex.Unlock()

	go func() {
		defer func() {
			hr.mutex.Lock()
			delete(hr.connections, ws)
			hr.mutex.Unlock()
			ws.Close()
		}()

		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
