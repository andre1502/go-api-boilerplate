package module

const (
	DEV = "Dev"
	STG = "Staging"
	PRD = "Production"

	HEADER_CONTENT_TYPE  = "Content-Type"
	HEADER_AUTHORIZATION = "Authorization"
	BEARER_AUTHORIZATION = "Bearer %s"

	APPLICATION_YAML = "application/yaml"
	APPLICATION_JSON = "application/json"
	APPLICATION_FORM = "application/x-www-form-urlencoded"

	TIMEOUT_SECONDS           = 60
	MAX_RETRY                 = 3
	MAX_QUERY_MONTH           = 3
	PAGE_SIZE                 = 20
	DATETIMEMS_FORMAT         = "2006-01-02 15:04:05.000"
	DATETIMEMS_ELASTIC_FORMAT = "2006-01-02T15:04:05.000Z"
	DATETIME_ELASTIC_FORMAT   = "2006-01-02T15:04:05Z"
)
