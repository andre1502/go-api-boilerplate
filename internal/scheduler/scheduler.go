package scheduler

import (
	"context"
	"errors"
	"fmt"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/db"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/logger"
	"go-api-boilerplate/module/redis"
	"sync"
	"time"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Scheduler struct {
	context        context.Context
	Wg             sync.WaitGroup
	Config         *config.Config
	DB             *db.DBConnection
	Redis          *redis.RedisConnection
	Elastic        elastic.ElasticConnections
	Cron           gocron.Scheduler
	jobs           map[uuid.UUID]*Job
	jobRedisPrefix string
	jobsRedisKey   []string
}

type Job struct {
	id         uuid.UUID
	name       string
	fn         func()
	definition gocron.JobDefinition
	error      error
}

func NewScheduler(
	cfg *config.Config,
	dbc *db.DBConnection,
	rds *redis.RedisConnection,
	elastic elastic.ElasticConnections,
) (*Scheduler, error) {
	locker, err := redislock.NewRedisLocker(rds.Client, redislock.WithTries(3), redislock.WithExpiry(30*time.Minute))
	if err != nil {
		return nil, err
	}

	cron, err := gocron.NewScheduler(gocron.WithStopTimeout(60*time.Second), gocron.WithDistributedLocker(locker))
	if err != nil {
		return nil, err
	}

	scheduler := &Scheduler{
		context:        context.Background(),
		Config:         cfg,
		DB:             dbc,
		Redis:          rds,
		Elastic:        elastic,
		Cron:           cron,
		jobs:           make(map[uuid.UUID]*Job),
		jobRedisPrefix: cfg.SCHEDULE_REDIS_PREFIX + ":%s",
		jobsRedisKey:   []string{},
	}

	scheduler.RegisterJob()
	err = scheduler.Schedule()
	if err != nil {
		msg := "error when schedule the job. %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, err
	}

	return scheduler, nil
}

func (s *Scheduler) Schedule() error {
	for uid, job := range s.jobs {
		if job.definition == nil {
			logger.Log.Warnf("[%s] Job definition is nil, this job will be skipped.", job.name)
			continue
		}

		cronjob, err := s.Cron.NewJob(job.definition, gocron.NewTask(job.fn),
			gocron.WithEventListeners(
				gocron.AfterJobRuns(s.jobDone()),
				gocron.AfterJobRunsWithError(s.jobError()),
				gocron.AfterJobRunsWithPanic(s.jobPanic()),
				gocron.AfterLockError(s.lockError()),
			), gocron.WithName(s.Redis.WrapKey(fmt.Sprintf(s.jobRedisPrefix, job.name))), gocron.WithIdentifier(uid))

		if err != nil {
			job.error = err
			return err
		}

		job.id = cronjob.ID()
	}

	return nil
}

func (s *Scheduler) CleanJobRedisKey() {
	if len(s.jobsRedisKey) == 0 {
		return
	}

	if err := s.Redis.DelCache(s.context, s.jobsRedisKey...); err != nil {
		logger.Log.Errorf("[%s] error on clean job redis key %v", s.jobsRedisKey, err)
	}
}

func (s *Scheduler) jobDone() func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		if err := s.Redis.DelCache(s.context, jobName); err != nil {
			logger.Log.Errorf("[%s_%s] error on delete redis key for job done, %v", jobID, jobName, err)
		}

		logger.Log.Infof("[%s_%s] - job done.", jobID, jobName)
	}
}

func (s *Scheduler) jobError() func(jobID uuid.UUID, jobName string, err error) {
	return func(jobID uuid.UUID, jobName string, err error) {
		if err := s.Redis.DelCache(s.context, jobName); err != nil {
			logger.Log.Errorf("[%s_%s] error on delete redis key for job error, %v", jobID, jobName, err)
		}

		logger.Log.Error(fmt.Sprintf("[%s_%s] job error", jobID, jobName), err)
	}
}

func (s *Scheduler) jobPanic() func(jobID uuid.UUID, jobName string, recoverData any) {
	return func(jobID uuid.UUID, jobName string, recoverData any) {
		if err := s.Redis.DelCache(s.context, jobName); err != nil {
			logger.Log.Errorf("[%s_%s] error on delete redis key for job panic, %v", jobID, jobName, err)
		}

		logger.Log.Errorf("[%s_%s] job panic %v", jobID, jobName, recoverData)
	}
}

func (s *Scheduler) lockError() func(jobID uuid.UUID, jobName string, err error) {
	return func(jobID uuid.UUID, jobName string, err error) {
		if errors.Is(err, redislock.ErrFailedToObtainLock) {
			logger.Log.Infof("[%s_%s] - ongoing.", jobID, jobName)
			return
		}

		logger.Log.Errorf("[%s_%s] lock error %v", jobID, jobName, err)
	}
}

func (s *Scheduler) RegisterJob() {
	// s.jobs[uuid.New()] = &Job{
	// 	name:       "DistributeScheduledPromotionReward",
	// 	fn:         s.distributeScheduledPromotionReward(),
	// 	definition: gocron.DurationJob(1 * time.Minute),
	// }

	for _, job := range s.jobs {
		s.jobsRedisKey = append(s.jobsRedisKey, fmt.Sprintf(s.jobRedisPrefix, job.name))
	}
}
