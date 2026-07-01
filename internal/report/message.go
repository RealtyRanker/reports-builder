// Package report builds the Telegram caption text accompanying a CSV report.
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/asmisnik/reports-builder/internal/model"
)

var ruMonths = [...]string{
	"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря",
}

func formatRuDateTime(t time.Time) string {
	return fmt.Sprintf("%d %s %02d:%02d", t.Day(), ruMonths[t.Month()-1], t.Hour(), t.Minute())
}

// TopN returns at most the first n flats (flats is expected to already be
// sorted by score descending).
func TopN(flats []model.FlatRecord, n int) []model.FlatRecord {
	if len(flats) <= n {
		return flats
	}
	return flats[:n]
}

// BuildCaption renders the message accompanying results.csv: the reporting
// window and a top-5 (or fewer) list of the best-scoring flats in it.
func BuildCaption(since, until time.Time, top []model.FlatRecord) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Отчёт о квартирах %s - %s\n", formatRuDateTime(since), formatRuDateTime(until))

	if len(top) == 0 {
		sb.WriteString("\nЗа это время подходящих квартир не найдено.")
		return sb.String()
	}

	fmt.Fprintf(&sb, "\nТоп-%d квартир за это время:\n", len(top))
	for i, f := range top {
		fmt.Fprintf(&sb, "%d. %s, score %d\n", i+1, f.Link, f.FlatScore)
	}
	return sb.String()
}
