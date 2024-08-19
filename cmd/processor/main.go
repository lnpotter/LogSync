package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
    logProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "log_processed_total",
            Help: "Total number of processed logs",
        },
        []string{"service"},
    )
    logErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "log_errors_total",
            Help: "Total number of log errors",
        },
        []string{"service"},
    )
)

func init() {
    prometheus.MustRegister(logProcessed)
    prometheus.MustRegister(logErrors)
}

type LogEntry struct {
    Timestamp   string                 `json:"timestamp"`
    Level       string                 `json:"level"`
    Service     string                 `json:"service"`
    Host        string                 `json:"host"`
    Message     string                 `json:"message"`
    Environment string                 `json:"environment"`
    Metadata    map[string]interface{} `json:"metadata"`
}

func sendLogToElasticsearch(logEntry LogEntry) error {
    url := "http://localhost:9200/logs/_doc/"
    logData, err := json.Marshal(logEntry)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(logData))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to send log: %s", resp.Status)
    }

    return nil
}

func processLogs(logEntry LogEntry) {
    if logEntry.Environment == "production" && logEntry.Level == "DEBUG" {
        log.Println("Filtered DEBUG log in production environment")
        return
    }

    logProcessed.WithLabelValues(logEntry.Service).Inc()
    if logEntry.Level == "ERROR" {
        logErrors.WithLabelValues(logEntry.Service).Inc()
    }

    logEntry = enrichLog(logEntry)
    err := sendLogToElasticsearch(logEntry)
    if err != nil {
        log.Printf("Error sending log to Elasticsearch: %s", err)
    } else {
        log.Println("Log processed and sent to Elasticsearch!")
    }
}

func enrichLog(logEntry LogEntry) LogEntry {
    if _, ok := logEntry.Metadata["correlation_id"]; !ok {
        logEntry.Metadata["correlation_id"] = generateCorrelationID()
    }
    logEntry.Metadata["geo_data"] = "Dummy Geo Data"
    return logEntry
}

func generateCorrelationID() string {
    return fmt.Sprintf("corr-id-%d", time.Now().UnixNano())
}

func consumeMessages() {
    c, err := kafka.NewConsumer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
        "group.id":          "logsync-group",
        "auto.offset.reset": "earliest",
    })

    if err != nil {
        log.Fatalf("Failed to create Kafka consumer: %s", err)
    }

    c.SubscribeTopics([]string{"logs"}, nil)

    for {
        msg, err := c.ReadMessage(-1)
        if err == nil {
            var logEntry LogEntry
            err := json.Unmarshal(msg.Value, &logEntry)
            if err == nil {
                processLogs(logEntry)
            }
        } else {
            log.Printf("Consumer error: %v (%v)\n", err, msg)
        }
    }

    c.Close()
}

func main() {
    http.Handle("/metrics", promhttp.Handler())
    go func() {
        log.Println("Prometheus metrics exposed on port 9092...")
        log.Fatal(http.ListenAndServe(":9092", nil))
    }()

    log.Println("Log Processor is running...")
    consumeMessages()
}
