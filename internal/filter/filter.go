// Package filter applies the report-subscription filters that aren't easily
// expressed as SQL predicates in db.GetFlatsForReport.
package filter

import (
	"github.com/asmisnik/reports-builder/internal/db"
	"github.com/asmisnik/reports-builder/internal/model"
)

// Apply returns the subset of flats that satisfy sub's remaining filters:
// minimum underground place (0 = no data, always excluded from a real
// threshold), minimum renovation rank, balcony-or-loggia, and bathroom type.
func Apply(flats []model.FlatRecord, sub db.ReportSubscription) []model.FlatRecord {
	result := make([]model.FlatRecord, 0, len(flats))
	for _, f := range flats {
		if sub.MinUndergroundPlace > 0 && (f.UndergroundPlace == 0 || f.UndergroundPlace > sub.MinUndergroundPlace) {
			continue
		}
		if sub.MinRenovation != "" && renovationRank(f.Renovation) < renovationRank(sub.MinRenovation) {
			continue
		}
		if sub.BalconyRequired && f.BalconyCount == 0 && f.LoggiaCount == 0 {
			continue
		}
		switch sub.BathroomType {
		case "separated":
			if f.SeparatedBathroomCount == 0 {
				continue
			}
		case "combined":
			if f.CombinedBathroomCount == 0 {
				continue
			}
		}
		result = append(result, f)
	}
	return result
}

// renovationRank ranks renovation levels design > euro > cosmetic > (any
// other value, including no renovation info), matching subscription-handler's
// session.Renovation* ordering and flats-analyzer's consumer.renovationRank.
func renovationRank(level string) int {
	switch level {
	case "design":
		return 3
	case "euro":
		return 2
	case "cosmetic":
		return 1
	default:
		return 0
	}
}
