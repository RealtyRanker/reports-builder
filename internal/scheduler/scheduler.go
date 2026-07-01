package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/asmisnik/reports-builder/internal/csvreport"
	"github.com/asmisnik/reports-builder/internal/db"
	"github.com/asmisnik/reports-builder/internal/filter"
	"github.com/asmisnik/reports-builder/internal/metrics"
	"github.com/asmisnik/reports-builder/internal/notifier"
	"github.com/asmisnik/reports-builder/internal/report"
)

const topN = 5

type Scheduler struct {
	db           *db.DB
	notifier     *notifier.Client
	logger       *zap.Logger
	pollInterval time.Duration
}

func New(database *db.DB, notifierClient *notifier.Client, logger *zap.Logger, pollInterval time.Duration) *Scheduler {
	return &Scheduler{
		db:           database,
		notifier:     notifierClient,
		logger:       logger,
		pollInterval: pollInterval,
	}
}

// Run polls for due report subscriptions every pollInterval until ctx is
// cancelled. It executes immediately on start.
func (s *Scheduler) Run(ctx context.Context) {
	s.tick(ctx)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	subs, err := s.db.GetDueReportSubscriptions(ctx)
	if err != nil {
		s.logger.Error("fetching due report subscriptions failed", zap.Error(err))
		return
	}
	if len(subs) == 0 {
		return
	}
	s.logger.Info("processing due report subscriptions", zap.Int("count", len(subs)))

	for _, sub := range subs {
		s.processSubscription(ctx, sub)
	}
}

func (s *Scheduler) processSubscription(ctx context.Context, sub db.ReportSubscription) {
	since := sub.LastReportSentAt
	until := time.Now()

	flats, err := s.db.GetFlatsForReport(ctx, sub, since, until)
	if err != nil {
		s.logger.Error("fetching flats for report failed",
			zap.Int("subscription_id", sub.ID), zap.Error(err))
		metrics.ReportsFailed.Inc()
		return
	}
	flats = filter.Apply(flats, sub)

	if len(flats) == 0 {
		s.logger.Info("no flats matched report subscription, skipping send",
			zap.Int("subscription_id", sub.ID), zap.Int64("chat_id", sub.ChatID))
		if err := s.db.UpdateLastReportSentAt(ctx, sub.ID, until); err != nil {
			s.logger.Warn("updating last_report_sent_at failed", zap.Int("subscription_id", sub.ID), zap.Error(err))
		}
		return
	}

	csvBytes, err := csvreport.Build(flats)
	if err != nil {
		s.logger.Error("building CSV report failed", zap.Int("subscription_id", sub.ID), zap.Error(err))
		metrics.ReportsFailed.Inc()
		return
	}

	caption := report.BuildCaption(since, until, report.TopN(flats, topN))

	if err := s.notifier.SendDocument(ctx, sub.ChatID, "results.csv", csvBytes, caption); err != nil {
		s.logger.Warn("sending report failed",
			zap.Int("subscription_id", sub.ID), zap.Int64("chat_id", sub.ChatID), zap.Error(err))
		metrics.ReportsFailed.Inc()
		return
	}

	if err := s.db.UpdateLastReportSentAt(ctx, sub.ID, until); err != nil {
		s.logger.Warn("updating last_report_sent_at failed", zap.Int("subscription_id", sub.ID), zap.Error(err))
	}

	metrics.ReportsSent.Inc()
	metrics.FlatsInReport.Observe(float64(len(flats)))
	s.logger.Info("report sent",
		zap.Int("subscription_id", sub.ID),
		zap.Int64("chat_id", sub.ChatID),
		zap.Int("flats_count", len(flats)),
	)
}
