package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time string `json:"time,omitempty"`
}

var clients = map[*websocket.Conn]bool{}
var mu sync.Mutex
var history = []Message{}

func broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	mu.Lock()
	history = append(history, msg)
	if len(history) > 50 {
		history = history[len(history)-50:]
	}
	var dead []*websocket.Conn
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			client.Close()
			dead = append(dead, client)
		}
	}
	mu.Unlock()
	for _, client := range dead {
		mu.Lock()
		delete(clients, client)
		mu.Unlock()
		client.Close()
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	mu.Lock()
	snapshot := slices.Clone(history)
	mu.Unlock()

	for _, message := range snapshot {
		data, err := json.Marshal(message)
		if err != nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil || string(message) == "" {
			log.Println(err)
			return
		}

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println(err)
			return
		}
		msg.Time = time.Now().Format("15:04:05")

		broadcast(msg)
	}
}

func main() {

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, nil)
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/ws", handleWS)

	log.Println("слушаю на http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
