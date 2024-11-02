package util

import (
	"context"
	"log"
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type cronJob func()

type Scheduler struct {
	schedule gocron.Scheduler
}

func NewScheduler() (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{schedule: s}, nil
}

func (rs *Scheduler) AddCronJob(cron string, job cronJob) (uuid.UUID, error) {
	j, err := rs.schedule.NewJob(
		gocron.CronJob(
			cron,
			false,
		),
		gocron.NewTask(
			job,
		),
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	return j.ID(), nil
}

func (rs *Scheduler) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	rs.schedule.Start()
	<-ctx.Done()
	log.Println("shutting down report service")
	err := rs.schedule.Shutdown()
	if err != nil {
		log.Println("failed to shut down report service")
	}
}
