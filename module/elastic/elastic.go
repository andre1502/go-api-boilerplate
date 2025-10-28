package elastic

import (
	"context"
	"errors"
	"fmt"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/logger"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

type ElasticConnections interface {
	LogToIndex(indexName string, message string, data map[string]interface{}) error
	SearchHitsIndex(indexName string, request *search.Request) (*types.HitsMetadata, error)
	SearchAggregationIndex(indexName string, request *search.Request) (map[string]types.Aggregate, error)

	AggregateDateHistogramData(ctx context.Context, indexName, rootField string, query *types.Query,
		aggregations map[string]types.Aggregations) ([]types.DateHistogramBucket, error)

	CalculateFromPage(page, pageSize int) int
}

type ElasticConnection struct {
	ctx    context.Context
	config *config.Config
	client *elasticsearch.TypedClient
}

func NewElasticConnection(cfg *config.Config) (ElasticConnections, error) {
	ctx := context.Background()
	elasticConfig := elasticsearch.Config{
		Addresses: []string{
			cfg.ELASTIC_URL,
		},
		Username: cfg.ELASTIC_USERNAME,
		Password: cfg.ELASTIC_PASSWORD,
	}

	client, err := elasticsearch.NewTypedClient(elasticConfig)
	if err != nil {
		msg := "elastic connection failed: %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, err
	}

	// 驗證連線是否成功
	_, err = client.Info().Header(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON).Header("Accept", module.APPLICATION_JSON).Do(ctx)
	if err != nil {
		msg := "failed to get client info: %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, err
	}

	es := &ElasticConnection{
		ctx:    ctx,
		config: cfg,
		client: client,
	}

	elasticLogger, err := NewLoggerElasticHook(es)
	if err != nil {
		msg := "failed to create elastic hook: %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, err
	}

	logger.Log.AddHook(elasticLogger)

	return es, nil
}

func (ec *ElasticConnection) IsIndexBlocked(indexName string) (*Setting, error) {
	setting := &Setting{}

	res, err := ec.client.Indices.GetSettings().Header(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON).
		Header("Accept", module.APPLICATION_JSON).Index(indexName).FlatSettings(true).Do(ec.ctx)

	if err != nil {
		var elasticErr *types.ElasticsearchError
		if errors.As(err, &elasticErr) {
			if elasticErr != nil && elasticErr.Status == http.StatusNotFound {
				return setting, nil
			}
		}

		return setting, ErrGetIndexSetting
	}

	if res == nil {
		return setting, nil
	}

	settings, ok := res[indexName]
	if !ok {
		return setting, nil
	}

	if settings.Settings == nil || settings.Settings.Index == nil || settings.Settings.Index.Blocks == nil {
		return setting, nil
	}

	if settings.Settings.Index.Blocks.Metadata != nil {
		value := settings.Settings.Index.Blocks.Metadata == "true"
		setting.IndexBlocksMetadata = &value
	}

	if settings.Settings.Index.Blocks.Read != nil {
		value := settings.Settings.Index.Blocks.Read == "true"
		setting.IndexBlocksRead = &value
	}

	if settings.Settings.Index.Blocks.ReadOnly != nil {
		value := settings.Settings.Index.Blocks.ReadOnly == "true"
		setting.IndexBlocksReadOnly = &value
	}

	if settings.Settings.Index.Blocks.ReadOnlyAllowDelete != nil {
		value := settings.Settings.Index.Blocks.ReadOnlyAllowDelete == "true"
		setting.IndexBlocksReadOnlyAllowDelete = &value
	}

	if settings.Settings.Index.Blocks.Write != nil {
		value := settings.Settings.Index.Blocks.Write == "true"
		setting.IndexBlocksWrite = &value
	}

	return setting, nil
}

func (ec *ElasticConnection) LogToIndex(indexName, message string, data map[string]interface{}) error {
	logEntry := LogEntry{
		Hostname:   ec.config.HOST_NAME,
		HostIP:     ec.config.HOST_IP,
		PodID:      ec.config.POD_ID,
		PodName:    ec.config.POD_NAME,
		PodIP:      ec.config.POD_IP,
		RepoName:   ec.config.REPO_NAME,
		BranchName: ec.config.BRANCH_NAME,
		CommitHash: ec.config.COMMIT_HASH,
		BuildDate:  ec.config.BUILD_DATE,
		Version:    ec.config.VERSION,
		AppName:    ec.config.APP_NAME,
		Timestamp:  time.Now(),
		Message:    message,
	}

	level, ok := data["level"].(string)
	if ok {
		logEntry.Level = &level
		delete(data, "level")
	}

	file, ok := data["file"].(string)
	if ok {
		logEntry.File = &file
		delete(data, "file")
	}

	function, ok := data["function"].(string)
	if ok {
		logEntry.Function = &function
		delete(data, "function")
	}

	if data != nil {
		logEntry.Data = &data
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Log.WithFields(logger.GetElasticLogFields(logEntry.Hostname, logEntry.HostIP, logEntry.PodID, logEntry.PodName, logEntry.PodIP, logEntry.RepoName,
				logEntry.BranchName, logEntry.CommitHash, logEntry.BuildDate, logEntry.Version, logEntry.AppName, logEntry.Timestamp, logEntry.Message, logEntry.Data)).
				Error("Panic in LogToIndex: ", r)
		}
	}()

	setting, err := ec.IsIndexBlocked(indexName)
	if err != nil {
		logger.Log.WithFields(logger.GetElasticLogFields(logEntry.Hostname, logEntry.HostIP, logEntry.PodID, logEntry.PodName, logEntry.PodIP, logEntry.RepoName,
			logEntry.BranchName, logEntry.CommitHash, logEntry.BuildDate, logEntry.Version, logEntry.AppName, logEntry.Timestamp, logEntry.Message, logEntry.Data)).
			Error(err)

		return err
	}

	if setting.IndexBlocksWrite != nil && *setting.IndexBlocksWrite {
		logger.Log.WithFields(logger.GetElasticLogFields(logEntry.Hostname, logEntry.HostIP, logEntry.PodID, logEntry.PodName, logEntry.PodIP, logEntry.RepoName,
			logEntry.BranchName, logEntry.CommitHash, logEntry.BuildDate, logEntry.Version, logEntry.AppName, logEntry.Timestamp, logEntry.Message, logEntry.Data)).
			Warn("Index write blocked")

		return nil
	}

	_, err = ec.client.Index(indexName).Header(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON).
		Header("Accept", module.APPLICATION_JSON).Request(logEntry).Do(ec.ctx)

	if err != nil {
		logger.Log.WithFields(logger.GetElasticLogFields(logEntry.Hostname, logEntry.HostIP, logEntry.PodID, logEntry.PodName, logEntry.PodIP, logEntry.RepoName,
			logEntry.BranchName, logEntry.CommitHash, logEntry.BuildDate, logEntry.Version, logEntry.AppName, logEntry.Timestamp, logEntry.Message, logEntry.Data)).
			Error("Failed to log to Elasticsearch: ", err)

		return err
	}

	return nil
}

func (ec *ElasticConnection) SearchHitsIndex(indexName string, request *search.Request) (*types.HitsMetadata, error) {
	request.TrackTotalHits = true

	res, err := ec.client.Search().Header(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON).
		Header("Accept", module.APPLICATION_JSON).Index(indexName).Request(request).Do(ec.ctx)

	if err != nil {
		return nil, ErrSearchIndex
	}

	return &types.HitsMetadata{
		Hits:  res.Hits.Hits,
		Total: res.Hits.Total,
	}, nil
}

func (ec *ElasticConnection) SearchAggregationIndex(indexName string, request *search.Request) (map[string]types.Aggregate, error) {
	request.TrackTotalHits = true

	res, err := ec.client.Search().Header(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON).
		Header("Accept", module.APPLICATION_JSON).Index(indexName).Request(request).Do(ec.ctx)

	if err != nil {
		if he, ok := err.(*types.ElasticsearchError); ok {
			if he.Status == http.StatusNotFound {
				return nil, nil
			}
		}

		return nil, ErrSearchIndex
	}

	return res.Aggregations, nil
}

func (ec *ElasticConnection) AggregateDateHistogramData(ctx context.Context, indexName, rootField string, query *types.Query,
	aggregations map[string]types.Aggregations) ([]types.DateHistogramBucket, error) {

	pageSize := 0

	request := &search.Request{
		Size:         &pageSize,
		Query:        query,
		Aggregations: aggregations,
	}

	res, err := ec.SearchAggregationIndex(indexName, request)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	dailyAggr, ok := res[rootField].(*types.DateHistogramAggregate)
	if !ok {
		return nil, ErrCastDataType
	}

	buckets, ok := dailyAggr.Buckets.([]types.DateHistogramBucket)
	if !ok {
		return nil, ErrCastDataType
	}

	return buckets, nil
}

func (ec *ElasticConnection) CalculateFromPage(page, pageSize int) int {
	from := 0

	if page > 1 {
		from = (page - 1) * pageSize
	}

	return from
}
