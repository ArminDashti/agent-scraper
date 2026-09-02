package pipeline

import (
	"context"
	"log"

	"github.com/ArminDashti/agent-scraper-api/internal/forwarder"
	"github.com/ArminDashti/agent-scraper-api/internal/scraper"
	"github.com/ArminDashti/agent-scraper-api/internal/store"
)

type Runner struct {
	store     *store.Store
	extractor *scraper.Extractor
	forwarder *forwarder.Client
}

func NewRunner(st *store.Store, extractor *scraper.Extractor, fwd *forwarder.Client) *Runner {
	return &Runner{store: st, extractor: extractor, forwarder: fwd}
}

func (r *Runner) Run(ctx context.Context) error {
	job, err := r.store.CreateJobRun(ctx)
	if err != nil {
		return err
	}

	sources, err := r.store.ListEnabledSources(ctx)
	if err != nil {
		_ = r.store.FinishJobRun(ctx, job.ID, "failed", 0, err.Error())
		return err
	}
	if len(sources) == 0 {
		return r.store.FinishJobRun(ctx, job.ID, "skipped", 0, "no enabled sources")
	}

	drafts, err := r.extractor.Extract(ctx, sources)
	if err != nil {
		_ = r.store.FinishJobRun(ctx, job.ID, "failed", 0, err.Error())
		return err
	}

	jobRunID := job.ID
	for _, draft := range drafts {
		sourceID := draft.SourceID
		row, insertErr := r.store.InsertExpense(ctx, draft.Shop, draft.Item, draft.Expense, &sourceID, &jobRunID)
		if insertErr != nil {
			_ = r.store.FinishJobRun(ctx, job.ID, "failed", 0, insertErr.Error())
			return insertErr
		}
		if deliverErr := r.forwarder.Deliver(ctx, *row); deliverErr != nil {
			log.Printf("outbound post failed for expense %d: %v", row.ID, deliverErr)
		}
	}

	return r.store.FinishJobRun(ctx, job.ID, "succeeded", len(drafts), "")
}
