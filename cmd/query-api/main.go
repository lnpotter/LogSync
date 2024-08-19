package main

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
)

type LogQueryResponse struct {
    Hits struct {
        Hits []struct {
            Source map[string]interface{} `json:"_source"`
        } `json:"hits"`
    } `json:"hits"`
}

func searchLogsHandler(w http.ResponseWriter, r *http.Request) {
    elasticSearchURL := "http://localhost:9200/logs/_search"

    query := `{
        "query": {
            "match": {
                "service": "auth-service"
            }
        }
    }`

    req, err := http.NewRequest("GET", elasticSearchURL, bytes.NewBuffer([]byte(query)))
    if err != nil {
        http.Error(w, "Error creating request", http.StatusInternalServerError)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Error querying Elasticsearch", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    var esResponse LogQueryResponse
    if err := json.NewDecoder(resp.Body).Decode(&esResponse); err != nil {
        http.Error(w, "Error parsing response", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(esResponse.Hits.Hits)
}

func main() {
    http.HandleFunc("/api/logs/search", searchLogsHandler)
    log.Println("Log Query API is running on port 8081...")
    log.Fatal(http.ListenAndServe(":8081", nil))
}
