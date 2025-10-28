package logger

import (
	"fmt"
	"go-api-boilerplate/module"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func NewLogger(fileName string) *logrus.Logger {
	Log = logrus.New()

	// 设置日志格式
	Log.SetFormatter(&logrus.JSONFormatter{})

	// 设置日志级别
	Log.SetLevel(logrus.InfoLevel)

	Log.SetReportCaller(true)

	// 创建日志文件目录，如果不存在
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	// 根据当前日期创建日志文件
	currentDate := time.Now().Format(time.DateOnly)
	logFileName := fmt.Sprintf("%s/%s_%s.log", logDir, fileName, currentDate)
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
	}

	// 设置日志输出到文件
	Log.SetOutput(logFile)

	return Log
}

func GetLogFields(method, requestUri, requestId string, ipAddress string, err error, httpStatus int, code int, message string) logrus.Fields {
	return map[string]interface{}{
		"method":     method,
		"uri":        requestUri,
		"requestId":  requestId,
		"ip_address": ipAddress,
		"error":      err,
		"httpStatus": httpStatus,
		"code":       code,
		"message":    message,
	}
}

func GetElasticLogFields(hostname, hostIP, podID, podName, podIP, repoName, branchName, commitHash, buildDate, version, appName string, timestamp time.Time,
	message string, data *map[string]interface{}) logrus.Fields {

	logFields := map[string]interface{}{
		"host_name":    hostname,
		"host_ip":      hostIP,
		"repo_name":    repoName,
		"branch_name":  branchName,
		"commit_hash":  commitHash,
		"build_date":   buildDate,
		"version":      version,
		"app_name":     appName,
		"timestamp":    timestamp,
		"message":      message,
		"from_elastic": true,
		"data":         data,
	}

	if !module.IsEmptyString(podID) {
		logFields["pod_id"] = podID
	}

	if !module.IsEmptyString(podName) {
		logFields["pod_name"] = podName
	}

	if !module.IsEmptyString(podIP) {
		logFields["pod_ip"] = podIP
	}

	return logFields
}
