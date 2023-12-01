package main

import (
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
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

	// interval for generating a tick (ms)
	tickInt = 100

	// Setup a counter with this period
	countPeriod = tickInt * time.Millisecond

	// How often to write to the client
	writePeriod = 1 * time.Second

	connLimit = 15
	indexTmpl = "index.html"
)

var (
	addr         = flag.String("addr", "localhost:8710", "http service address")
	fname        = flag.String("fname", "", "file to persist counter (opt)")
	allowedHosts = []string{
		"jarv.org",
		"like.jarv.org",
		"cmdchallenge.com",
		"12days.cmdchallenge.com",
		"oops.cmdchallenge.com",
		"localhost",
		"127.0.0.1",
	}

	homeTempl = template.Must(template.New(indexTmpl).ParseFiles(indexTmpl))
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return isAllowedHost(origin)
		},
	}
	logger      = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	connLimiter = make(chan struct{}, connLimit)
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
		<-connLimiter
		logger.Info("Finished connection", "remoteAddr", ws.RemoteAddr().String())
	}()

	select {
	case connLimiter <- struct{}{}:
		logger.Info("Starting connection", "remoteAddr", ws.RemoteAddr().String())
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
	default:
		logger.Warn("Too many connections!", "remoteAddr", ws.RemoteAddr().String())
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
				logger.Error("Persisting count to a file failed!", "err", err)
				os.Exit(1)
			}
		}
	}
}

func serveWs(cnt *Counter, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logger.Error("websocket upgrade error", "err", err)
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

	d := struct {
		Host string
		Cnt  string
	}{
		r.Host,
		cnt.Disp(),
	}
	homeTempl.Execute(w, d)
}

func resetCnt(cnt *Counter, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	origin := r.Header.Get("Origin")
	if isAllowedHost(origin) {
		allowOrigin(w, r)
	}

	logger.Info("Counter reset!", "RemoteAddr", r.RemoteAddr)
	cnt.Reset()
	w.WriteHeader(http.StatusOK)
}

func isAllowedHost(origin string) bool {
	for _, h := range allowedHosts {
		if strings.HasPrefix(origin, "https://"+h) ||
			strings.HasPrefix(origin, "http://"+h) {
			return true
		}
	}

	return false
}

func allowOrigin(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "PUT")
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
		logger.Error("Unable to read from storage", "err", err)
		os.Exit(1)
	}

	cnt := NewCounter(initialValue)

	go persistCount(cnt, store)
	go incCount(cnt)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveHome(cnt, w, r)
	})

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(cnt, w, r)
	})

	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			origin := r.Header.Get("Origin")
			if isAllowedHost(origin) {
				allowOrigin(w, r)
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		resetCnt(cnt, w, r)
	})

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}

	logger.Info("Server started", "addr", *addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Unable to setup listener", "err", err)
		os.Exit(1)
	}
}
