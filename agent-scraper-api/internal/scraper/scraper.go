package scraper

import (
	"context"

	"github.com/ArminDashti/agent-scraper-api/internal/store"
)

type Draft struct {
	Shop     string
	Item     string
	Expense  string
	SourceID int64
}

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(_ context.Context, sources []store.Source) ([]Draft, error) {
	if len(sources) == 0 {
		return nil, nil
	}
	return nil, nil
}
