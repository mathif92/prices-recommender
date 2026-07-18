package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/mathif92/prices-recommender/pkg/repositories"
)

type CollectorRunner interface {
	Run(ctx context.Context, userID int64) error
}

type Scheduler struct {
	log       logrus.FieldLogger
	repo      repositories.DataRepository
	collector CollectorRunner
	cron      *cron.Cron
	jobs      map[int64]cron.EntryID
}

func NewScheduler(log logrus.FieldLogger, repo repositories.DataRepository, collector CollectorRunner) *Scheduler {
	return &Scheduler{
		log:       log,
		repo:      repo,
		collector: collector,
		cron:      cron.New(cron.WithLocation(time.UTC)),
		jobs:      make(map[int64]cron.EntryID),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	schedules, err := s.repo.ListActiveSchedules(ctx)
	if err != nil {
		return fmt.Errorf("failed to load collection schedules: %w", err)
	}

	for _, schedule := range schedules {
		s.addJob(schedule)
	}

	s.cron.Start()
	s.log.Infof("scheduler started with %d active schedule(s)", len(schedules))

	for _, schedule := range schedules {
		entryID, ok := s.jobs[schedule.ID]
		if !ok {
			continue
		}
		entry := s.cron.Entry(entryID)
		s.log.Infof("schedule %d next run at %s (cron: %s, user: %d)",
			schedule.ID, entry.Next.Format("2006-01-02 15:04:05 UTC"), schedule.CronExpression, schedule.UserID)
	}

	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.log.Info("scheduler stopped")
}

func (s *Scheduler) AddSchedule(schedule repositories.CollectionSchedule) {
	if !schedule.IsActive {
		return
	}
	s.addJob(schedule)

	entryID, ok := s.jobs[schedule.ID]
	if ok {
		entry := s.cron.Entry(entryID)
		s.log.Infof("schedule %d added — next run at %s (cron: %s, user: %d)",
			schedule.ID, entry.Next.Format("2006-01-02 15:04:05 UTC"), schedule.CronExpression, schedule.UserID)
	}
}

func (s *Scheduler) RemoveSchedule(scheduleID int64) {
	if entryID, ok := s.jobs[scheduleID]; ok {
		s.cron.Remove(entryID)
		delete(s.jobs, scheduleID)
		s.log.Infof("removed schedule %d", scheduleID)
	}
}

func (s *Scheduler) addJob(schedule repositories.CollectionSchedule) {
	entryID, err := s.cron.AddFunc(schedule.CronExpression, func() {
		s.log.Infof("running scheduled collection for schedule %d (user %d)", schedule.ID, schedule.UserID)
		if err := s.collector.Run(context.Background(), schedule.UserID); err != nil {
			s.log.Errorf("scheduled collection failed for schedule %d: %v", schedule.ID, err)
		}
	})
	if err != nil {
		s.log.Errorf("failed to add schedule %d (invalid cron expression %q): %v", schedule.ID, schedule.CronExpression, err)
		return
	}
	s.jobs[schedule.ID] = entryID
}
