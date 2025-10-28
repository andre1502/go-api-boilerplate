package elastic

import (
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

type Setting struct {
	IndexBlocksMetadata            *bool `json:"index.blocks.metadata"`
	IndexBlocksRead                *bool `json:"index.blocks.read"`
	IndexBlocksReadOnly            *bool `json:"index.blocks.read_only"`
	IndexBlocksReadOnlyAllowDelete *bool `json:"index.blocks.read_only_allow_delete"`
	IndexBlocksWrite               *bool `json:"index.blocks.write"`
}

type LogEntry struct {
	Hostname   string                  `json:"host_name"`
	HostIP     string                  `json:"host_ip"`
	PodID      string                  `json:"pod_id"`
	PodName    string                  `json:"pod_name"`
	PodIP      string                  `json:"pod_ip"`
	RepoName   string                  `json:"repo_name"`
	BranchName string                  `json:"branch_name"`
	CommitHash string                  `json:"commit_hash"`
	BuildDate  string                  `json:"build_date"`
	Version    string                  `json:"version"`
	AppName    string                  `json:"app_name"`
	Timestamp  time.Time               `json:"timestamp"`
	Level      *string                 `json:"level,omitempty"`
	File       *string                 `json:"file,omitempty"`
	Function   *string                 `json:"function,omitempty"`
	Message    string                  `json:"message"`
	Data       *map[string]interface{} `json:"data,omitempty"`
}

type QueryAggregation struct {
	Index          string
	RootField      string
	PageSize       int
	Query          *types.Query
	Aggregations   map[string]types.Aggregations
	BucketName     string
	AggregateField string
}
