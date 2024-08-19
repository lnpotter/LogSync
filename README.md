
# LogSync - Distributed Logging System

## Overview
LogSync is a comprehensive distributed logging system built in Go. It is designed to handle log collection, processing, storage, and visualization across multiple services and environments. LogSync enables easy monitoring, filtering, alerting, and retention of logs while supporting multiple log formats. Its modular architecture ensures scalability and high availability.

## Key Features

1. **Log Filtering and Enrichment**:
   - **Log Filtering**: Logs are filtered based on criteria such as environment and log level. For example, DEBUG logs can be ignored in production environments.
   - **Log Enrichment**: Logs are enriched with metadata like correlation IDs, service information, and geographic data.

2. **Alerts & Monitoring**:
   - **Prometheus Metrics**: The system tracks custom metrics such as `log_processed_total` and `log_errors_total`, which are exposed via a `/metrics` endpoint.
   - **Threshold-Based Alerts**: Alerts are triggered when specific thresholds are exceeded (e.g., high error rates). Prometheus rules are defined for alerting based on log processing metrics.
   - **Alert Integrations**: Integrations with external notification services like Slack and PagerDuty via Prometheus Alertmanager.

3. **Log Retention Policy**:
   - **Configurable Retention Periods**: Retention periods can be configured for each service (e.g., 30 days for critical logs, 7 days for others). After the retention period, logs are either deleted or archived.
   - **Archiving to AWS S3**: Logs can be archived to AWS S3 for long-term storage.

4. **Distributed Processing & High Availability**:
   - **Kafka Integration**: Distributed log processing is supported via Kafka. The system can scale horizontally by adding more processing nodes.
   - **Clustered Elasticsearch**: Logs are stored in a clustered Elasticsearch setup, ensuring data redundancy and high availability.

5. **Support for Multiple Log Formats**:
   - **Syslog Integration**: The system listens on UDP port 514 for Syslog messages and processes them accordingly.
   - **Fluentd Integration**: Logs are collected from Fluentd via an HTTP endpoint.

6. **Custom Dashboards**:
   - **Grafana Dashboards**: Pre-configured dashboards visualize metrics such as log volume, error rates, and performance metrics.
   - **Kibana Dashboards**: Logs stored in Elasticsearch can be visualized and queried using Kibana.

## System Architecture

### Components
- **Log Collector**: Collects logs via Syslog, Fluentd, and HTTP sources.
- **Log Processor**: Processes logs by filtering and enriching them before storing them in Elasticsearch.
- **Query API**: Provides an HTTP API for querying logs from Elasticsearch.
- **Retention Archiver**: Manages the retention policy by archiving logs to S3 or deleting them from Elasticsearch after a specified period.
- **Prometheus & Alertmanager**: Handles metrics monitoring, alerting, and external notifications.

### Data Flow
1. Logs are ingested by the **Log Collector** via HTTP, Syslog, or Fluentd.
2. Logs are sent to **Kafka** for distributed processing.
3. The **Log Processor** consumes the logs, applies filters, enriches them with metadata, and stores them in **Elasticsearch**.
4. **Prometheus** scrapes metrics from the Log Processor, tracking log processing rates and errors.
5. Alerts are triggered based on metrics and routed through **Alertmanager** to external services like Slack and PagerDuty.
6. Logs are queried through the **Query API** and visualized via **Grafana** and **Kibana** dashboards.
7. The **Retention Archiver** enforces log retention policies by archiving or deleting logs based on the configured rules.

## Tools & Versions

- **Go**: 1.20+
- **Prometheus**: 2.30+
- **Grafana**: 8.2+
- **Elasticsearch**: 7.10+
- **Kafka**: 2.8+
- **AWS SDK for Go**: v1.38+
- **Fluentd**: 1.13+
- **Kibana**: 7.10+
- **Slack & PagerDuty**: For external alert notifications

## Installation and Setup

### Prerequisites
1. **Install Go**: Ensure that Go 1.20 or higher is installed on your system.
2. **Kafka & Zookeeper**: Install and start Kafka and Zookeeper for log processing.
3. **Elasticsearch**: Install and configure Elasticsearch.
4. **Prometheus & Grafana**: Install Prometheus for metrics collection and Grafana for visualizing metrics.
5. **Alertmanager**: Install Alertmanager and configure it for notifications (Slack, PagerDuty, etc.).

### Steps to Run LogSync

1. **Clone the Repository**:
   ```bash
   git clone github.com/lnpotter/LogSync
   cd LogSync
   ```

2. **Set up Kafka**:
   Start Zookeeper and Kafka:
   ```bash
   ./bin/zookeeper-server-start.sh config/zookeeper.properties
   ./bin/kafka-server-start.sh config/server.properties
   ```

3. **Set up Elasticsearch**:
   Start the Elasticsearch cluster:
   ```bash
   ./bin/elasticsearch
   ```

4. **Set up Prometheus and Grafana**:
   Configure Prometheus to scrape metrics from LogSync and set up Grafana to visualize those metrics.

5. **Run the Log Collector**:
   ```bash
   go run cmd/collector/main.go
   ```

6. **Run the Log Processor**:
   ```bash
   go run cmd/processor/main.go
   ```

7. **Run the Query API**:
   ```bash
   go run cmd/query-api/main.go
   ```

8. **Set up Retention Policies**:
   Edit the `internal/retention/config.json` file to configure the retention period and archiving rules for each service.

9. **Monitor Logs and Metrics**:
   Use Grafana and Kibana to visualize logs and monitor system performance. Set up Prometheus alerts for critical metrics.

## Contributing

Contributions to LogSync are welcome! Feel free to submit issues or pull requests to improve the project.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
