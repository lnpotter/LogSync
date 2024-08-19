package main

import (
    "encoding/json"
    "log"
    "net"
    "net/http"
    "strings"
)

type LogEntry struct {
    Timestamp   string                 `json:"timestamp"`
    Level       string                 `json:"level"`
    Service     string                 `json:"service"`
    Host        string                 `json:"host"`
    Message     string                 `json:"message"`
    Environment string                 `json:"environment"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// SyslogHandler handles incoming Syslog messages
func syslogHandler(conn *net.UDPConn) {
    buf := make([]byte, 1024)

    for {
        n, addr, err := conn.ReadFromUDP(buf)
        if err != nil {
            log.Printf("Error reading from Syslog connection: %v", err)
            continue
        }

        message := strings.TrimSpace(string(buf[:n]))
        log.Printf("Received Syslog message from %v: %v", addr, message)

        // Convert the Syslog message into a LogEntry
        logEntry := LogEntry{
            Timestamp:   fmt.Sprintf("%v", addr),
            Level:       "INFO",
            Service:     "syslog",
            Host:        addr.String(),
            Message:     message,
            Environment: "production",
            Metadata:    make(map[string]interface{}),
        }

        processSyslog(logEntry)
    }
}

// StartSyslogServer initializes a Syslog listener
func startSyslogServer() {
    addr, err := net.ResolveUDPAddr("udp", ":514")
    if err != nil {
        log.Fatalf("Failed to resolve UDP address: %v", err)
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        log.Fatalf("Failed to start Syslog server: %v", err)
    }

    defer conn.Close()
    log.Println("Syslog server is running on port 514...")
    syslogHandler(conn)
}

// FluentdHandler handles incoming Fluentd messages via HTTP
func fluentdHandler(w http.ResponseWriter, r *http.Request) {
    var logEntry LogEntry
    if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    log.Printf("Received Fluentd log: %+v", logEntry)
    processFluentd(logEntry)
    w.WriteHeader(http.StatusAccepted)
    w.Write([]byte(`{"status":"success"}`))
}

func processSyslog(logEntry LogEntry) {
    log.Println("Processing Syslog log entry:", logEntry)
}

func processFluentd(logEntry LogEntry) {
    log.Println("Processing Fluentd log entry:", logEntry)
}

func main() {
    // HTTP server for Fluentd logs
    http.HandleFunc("/api/fluentd/logs", fluentdHandler)
    go func() {
        log.Println("Fluentd log collector running on port 8081...")
        log.Fatal(http.ListenAndServe(":8081", nil))
    }()

    // Start the Syslog server
    startSyslogServer()
}
