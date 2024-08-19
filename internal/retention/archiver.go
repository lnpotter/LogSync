package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "time"
    "bytes"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/olivere/elastic/v7"
)

type RetentionPolicy struct {
    RetentionDays int    `json:"retention_days"`
    ArchiveToS3   bool   `json:"archive_to_s3"`
    S3Bucket      string `json:"s3_bucket,omitempty"`
}

type RetentionConfig struct {
    RetentionPolicies map[string]RetentionPolicy `json:"retention_policies"`
}

// Load retention configuration from a JSON file
func loadRetentionConfig(filePath string) (RetentionConfig, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return RetentionConfig{}, fmt.Errorf("failed to read retention config: %s", err)
    }

    var config RetentionConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return RetentionConfig{}, fmt.Errorf("failed to parse retention config: %s", err)
    }

    return config, nil
}

// Archive logs to S3
func archiveLogToS3(logID string, logData string, bucket string) error {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
    })
    if err != nil {
        return fmt.Errorf("failed to create AWS session: %s", err)
    }

    svc := s3.New(sess)

    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(fmt.Sprintf("logs/%s.json", logID)),
        Body:   bytes.NewReader([]byte(logData)),
    })
    if err != nil {
        return fmt.Errorf("failed to upload log to S3: %s", err)
    }

    return nil
}

// Delete logs from Elasticsearch that are older than the retention period
func deleteOldLogs(client *elastic.Client, service string, retentionDays int) error {
    cutoff := time.Now().AddDate(0, 0, -retentionDays).Format(time.RFC3339)

    query := elastic.NewRangeQuery("timestamp").Lt(cutoff)
    searchResult, err := client.Search().
        Index("logs").
        Query(query).
        Size(1000).
        Do(context.Background())
    if err != nil {
        return fmt.Errorf("failed to search logs: %s", err)
    }

    for _, hit := range searchResult.Hits.Hits {
        logID := hit.Id
        logData := string(hit.Source)

        // Check if we need to archive the log
        if retentionConfig.RetentionPolicies[service].ArchiveToS3 {
            bucket := retentionConfig.RetentionPolicies[service].S3Bucket
            err = archiveLogToS3(logID, logData, bucket)
            if err != nil {
                log.Printf("Failed to archive log %s: %s", logID, err)
                continue
            }
        }

        // Delete the log from Elasticsearch
        _, err := client.Delete().
            Index("logs").
            Id(logID).
            Do(context.Background())
        if err != nil {
            log.Printf("Failed to delete log %s: %s", logID, err)
        } else {
            log.Printf("Log %s deleted successfully", logID)
        }
    }

    return nil
}

// Process retention policies for all services
func processRetention(client *elastic.Client, config RetentionConfig) {
    for service, policy := range config.RetentionPolicies {
        log.Printf("Processing retention for service: %s", service)
        err := deleteOldLogs(client, service, policy.RetentionDays)
        if err != nil {
            log.Printf("Failed to process retention for service %s: %s", service, err)
        }
    }
}

func main() {
    client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
    if err != nil {
        log.Fatalf("Failed to create Elasticsearch client: %s", err)
    }

    config, err := loadRetentionConfig("config.json")
    if err != nil {
        log.Fatalf("Failed to load retention configuration: %s", err)
    }

    // Run retention processing
    processRetention(client, config)
}
