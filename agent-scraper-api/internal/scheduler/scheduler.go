package scheduler

import (
	"context"
	"log"

	"github.com/ArminDashti/agent-scraper-api/internal/pipeline"
	"github.com/robfig/cron/v3"
)

type Service struct {
	cron     *cron.Cron
	schedule string
}

func Start(schedule string, runner *pipeline.Runner) (*Service, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	c := cron.New(cron.WithParser(parser), cron.WithChain(cron.Recover(cron.DefaultLogger)))
	_, err := c.AddFunc(schedule, func() {
		if runErr := runner.Run(context.Background()); runErr != nil {
			log.Printf("scheduled scrape failed: %v", runErr)
		}
	})
	if err != nil {
		return nil, err
	}
	c.Start()
	log.Printf("scrape schedule started: %s", schedule)
	return &Service{cron: c, schedule: schedule}, nil
}

func (s *Service) Stop() {
	if s != nil && s.cron != nil {
		s.cron.Stop()
	}
}

func (s *Service) Schedule() string {
	if s == nil {
		return ""
	}
	return s.schedule
}
