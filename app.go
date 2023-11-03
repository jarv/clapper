package main

import (
	"flag"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// How often to persist the counter value to disk
	persistPeriod = 10 * time.Second

	// Send a counter with this period
	countPeriod = 1 * time.Millisecond

	// How often to write to the client
	writePeriod = 100 * time.Millisecond
)

var (
	addr  = flag.String("addr", ":8710", "http service address")
	fname = flag.String("fname", "", "file to persist counter (opt)")

	homeTempl = template.Must(template.New("").Parse(homeHTML))
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// CheckOrigin: func(r *http.Request) bool {
		// 	origin := r.Header.Get("Origin")
		// 	return origin == "http://localhost:1313"
		// },
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(cnt *Counter, ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	writeTicker := time.NewTicker(writePeriod)

	defer func() {
		pingTicker.Stop()
		writeTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-writeTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(cnt.Disp())); err != nil {
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func incCount(c *Counter) {
	cntTicker := time.NewTicker(countPeriod)
	defer cntTicker.Stop()

	for {
		select {
		case <-cntTicker.C:
			c.Inc()
		}
	}
}

func persistCount(c *Counter, s Storer) {
	persistTicker := time.NewTicker(persistPeriod)
	defer persistTicker.Stop()

	for {
		select {
		case <-persistTicker.C:
			if *fname == "" {
				return
			}

			if err := s.Write(c.Load()); err != nil {
				slog.Error("Persisting count to a file failed!", "err", err)
			}
		}
	}
}

func serveWs(cnt *Counter, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writer(cnt, ws)
	reader(ws)
}

func serveHome(cnt *Counter, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var v = struct {
		Host string
		Data string
	}{
		r.Host,
		cnt.Disp(),
	}
	homeTempl.Execute(w, &v)
}

func resetCnt(cnt *Counter, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cnt.Reset()
	w.WriteHeader(http.StatusOK)
}

func main() {
	flag.Parse()
	var store Storer

	if *fname != "" {
		store = NewFileStore(*fname)
	} else {
		store = NewMemStore()
	}

	initialValue, err := store.Read()
	if err != nil {
		log.Fatalf("Unable to read from storage: %v", err)
	}

	cnt := NewCounter(initialValue)

	go persistCount(cnt, store)
	go incCount(cnt)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveHome(cnt, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(cnt, w, r)
	})

	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		resetCnt(cnt, w, r)
	})

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	slog.Info("Server started", "addr", *addr)
	log.Fatal(server.ListenAndServe())
}

const homeHTML = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>WebSocket Example</title>
    </head>
    <body>
		<button onclick="resetCnt()">👇 <span style="font-family: monospace;" id="cnt"></span></button>
        <script type="text/javascript">
            function resetCnt() {
              fetch("//{{.Host}}/reset", {
                method: "PUT",
              })
            }
            (function() {
                var data = document.getElementById("cnt");
				var wss = (window.location.protocol == "https:") ? "wss:" : "ws:"
				var conn = new WebSocket(wss + "//{{.Host}}/ws")
                conn.onclose = function(evt) {
                    data.textContent = 'Connection closed';
                }
                conn.onmessage = function(evt) {
                    data.textContent = evt.data;
                }
            })();
        </script>
    </body>
</html>
`
