package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed templates/*
var content embed.FS

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan ScanMessage)
	mutex     sync.Mutex
)

type ScanMessage struct {
	Type    string      `json:"type"` // "progress", "result", "complete"
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ScanProgress struct {
	Current    int     `json:"current"`
	Total      int     `json:"total"`
	Percentage float64 `json:"percentage"`
	AliveHosts int     `json:"alive_hosts"`
	OpenPorts  int     `json:"open_ports"`
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/results", handleResultsPage)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/results", handleResults)
	http.HandleFunc("/api/result/", handleResultDetail)
	
	go handleMessages()

	port := ":8888"
	fmt.Printf("🚀 网络空间扫描Web服务已启动\n")
	fmt.Printf("📡 扫描控制台: http://localhost%s\n", port)
	fmt.Printf("📊 结果展示页: http://localhost%s/results\n", port)
	fmt.Printf("⏰ 启动时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	log.Fatal(http.ListenAndServe(port, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(content, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleResultsPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(content, "templates/results.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket错误: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go runDemoScan()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "扫描已开始",
	})
}

func handleResults(w http.ResponseWriter, r *http.Request) {
	resultsDir := "../results"
	files, err := os.ReadDir(resultsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			info, _ := file.Info()
			results = append(results, map[string]interface{}{
				"name": file.Name(),
				"size": info.Size(),
				"time": info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handleResultDetail(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/api/result/"):]
	filePath := filepath.Join("../results", filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func runDemoScan() {
	broadcast <- ScanMessage{
		Type:    "progress",
		Message: "正在初始化扫描...",
		Data: ScanProgress{
			Current:    0,
			Total:      95,
			Percentage: 0,
		},
	}

	time.Sleep(1 * time.Second)

	for i := 1; i <= 95; i++ {
		time.Sleep(50 * time.Millisecond)
		broadcast <- ScanMessage{
			Type:    "progress",
			Message: fmt.Sprintf("正在扫描目标 %d/95", i),
			Data: ScanProgress{
				Current:    i,
				Total:      95,
				Percentage: float64(i) * 100 / 95,
				AliveHosts: i / 2,
				OpenPorts:  i,
			},
		}
	}

	broadcast <- ScanMessage{
		Type:    "complete",
		Message: "扫描完成！",
		Data: map[string]interface{}{
			"total":       95,
			"alive":       51,
			"open_ports":  56,
			"duration":    "5.2s",
			"result_file": "scan_result_" + time.Now().Format("20060102_150405") + ".json",
		},
	}
}
