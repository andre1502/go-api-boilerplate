package elastic

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// LoggerElasticHook is a Logrus hook for sending logs to Elasticsearch.
type LoggerElasticHook struct {
	es        ElasticConnections
	indexName string // e.g., "mylogs" or a pattern like "mylogs-2024-01-01"
}

// NewLoggerElasticHook creates a new hook.
func NewLoggerElasticHook(es ElasticConnections) (*LoggerElasticHook, error) {
	return &LoggerElasticHook{
		es:        es,
		indexName: ELASTIC_INTERNAL_LOG_INDEX,
	}, nil
}

// Levels defines the log levels for which this hook will be triggered.
func (h *LoggerElasticHook) Levels() []logrus.Level {
	return logrus.AllLevels // Trigger for all log levels
	// Or specify particular levels:
	// return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

// Fire is called when a log event occurs.
func (h *LoggerElasticHook) Fire(entry *logrus.Entry) error {
	// Prepare the log document for Elasticsearch
	indexName := h.indexName
	fromElastic := false
	doc := make(map[string]interface{})
	doc["level"] = entry.Level.String()

	// Add custom fields from Logrus entry.Data
	for k, v := range entry.Data {
		// Sanitize keys if necessary (e.g., replace dots, ensure ES compatibility)
		safeKey := strings.ReplaceAll(k, ".", "_")

		switch safeKey {
		case "elastic_index":
			indexName = v.(string)
			continue
		case "from_elastic":
			fromElastic = v.(bool)
			continue
		}

		doc[safeKey] = v
	}

	// Add caller info if available (from entry.Caller)
	if entry.HasCaller() {
		doc["file"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		doc["function"] = entry.Caller.Function
	}

	if fromElastic {
		return nil
	}

	h.es.LogToIndex(indexName, entry.Message, doc)

	return nil
}
