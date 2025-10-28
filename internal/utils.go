package internal

import (
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/elastic"
	"strings"
)

func GetLogIndex(requestUri string) string {
	if module.IsEmptyString(requestUri) {
		return elastic.ELASTIC_INTERNAL_LOG_INDEX
	}

	if strings.HasPrefix(requestUri, "/backoffice") {
		return elastic.ELASTIC_BACKOFFICE_LOG_INDEX
	}

	return elastic.ELASTIC_PLATFORM_LOG_INDEX
}
