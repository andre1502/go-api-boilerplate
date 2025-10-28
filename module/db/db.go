package db

import (
	"context"
	"fmt"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/logger"
	"strings"
	"time"

	gormLoggerLogrus "github.com/nekomeowww/gorm-logger-logrus"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type dbKey struct{}
type DBConnection struct {
	Client *gorm.DB
}

func NewDB(cfg *config.Config) (*DBConnection, error) {
	if cfg.DB == nil {
		fmt.Println(ErrEmptyConfig)
		logger.Log.Fatal(ErrEmptyConfig)

		return nil, ErrEmptyConfig
	}

	var err error

	gormLog := gormLoggerLogrus.New(gormLoggerLogrus.Options{
		Logger:                    logrus.NewEntry(logger.Log).WithField("elastic_index", elastic.ELASTIC_GORM_LOG_INDEX),
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: false,
		SlowThreshold:             time.Second * 5,
		FileWithLineNumField:      "repository.file",
	})

	client, err := gorm.Open(mysql.Open(cfg.DB.DB_URL), &gorm.Config{
		PrepareStmt:    true,
		TranslateError: true,
		Logger:         gormLog,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		msg := "error when open database connection. %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, ErrOpenConnection
	}

	clientDb, err := client.DB()
	if err != nil {
		return nil, err
	}

	clientDb.SetMaxOpenConns(module.StrToIntDefault(cfg.DB.DB_MAX_OPEN_CONNS, 100))
	clientDb.SetMaxIdleConns(module.StrToIntDefault(cfg.DB.DB_MAX_IDLE_CONNS, 20))
	clientDb.SetConnMaxIdleTime(time.Duration(module.StrToIntDefault(cfg.DB.DB_CONN_MAX_IDLE_TIME_SECOND, 600)) * time.Second)
	clientDb.SetConnMaxLifetime(time.Duration(module.StrToIntDefault(cfg.DB.DB_CONN_MAX_LIFE_TIME_SECOND, 1800)) * time.Second)

	return &DBConnection{
		Client: client,
	}, nil
}

func (dbc *DBConnection) GetTotalRecord(result *gorm.DB, countFieldName string) (uint64, error) {
	var dummyRes []map[string]interface{}
	var sql strings.Builder
	additionalVars := 0

	stmt := result.Session(&gorm.Session{DryRun: true}).Find(&dummyRes).Statement
	stmtSql := stmt.SQL.String()
	stmtVars := stmt.Vars
	totalVars := len(stmt.Vars)

	indexFrom := strings.Index(stmtSql, "FROM")
	indexGroupBy := strings.Index(stmtSql, "GROUP BY")
	indexOrderBy := strings.Index(stmtSql, "ORDER BY")
	indexLimit := strings.Index(stmtSql, "LIMIT")
	indexOffset := strings.Index(stmtSql, "OFFSET")

	if indexLimit > -1 {
		additionalVars++
		stmtSql = stmtSql[:indexLimit]
	}

	if indexOffset > -1 {
		additionalVars++
	}

	if additionalVars > 0 {
		stmtVars = stmtVars[0:(totalVars - additionalVars)]
	}

	if indexOrderBy > -1 {
		stmtSql = stmtSql[:indexOrderBy]
	}

	if indexGroupBy > -1 {
		sql.WriteString(fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM (%s) AS sq", stmtSql))
	} else if indexFrom > -1 {
		stmtSql = stmtSql[indexFrom:]

		sql.WriteString(fmt.Sprintf("SELECT COUNT(DISTINCT %s) AS total_rows ", countFieldName))
		sql.WriteString(stmtSql)
	}

	var rows *TableTotalRows
	output := result.Session(&gorm.Session{}).Raw(sql.String(), stmtVars...).Scan(&rows)

	if output.Error != nil {
		logger.Log.Errorf("error when get total record: %v", output.Error)
		return 0, ErrScanTotalRecord
	}

	return rows.TotalRows, nil
}

func (dbc *DBConnection) CtxWithSession(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, tx)
}

func (dbc *DBConnection) GetSession(ctx context.Context, currentDb *gorm.DB) *gorm.DB {
	session := ctx.Value(dbKey{})

	if session == nil {
		session = currentDb
	}

	return session.(*gorm.DB)
}
