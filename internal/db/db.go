package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/asmisnik/reports-builder/internal/model"
)

// ReportSubscription mirrors a report_user_subscriptions row: the same
// filters as a regular user_subscriptions row, plus a send period and the
// timestamp of the last report sent.
type ReportSubscription struct {
	ID     int
	ChatID int64

	DealType      string
	Region        int
	MetroStations []string

	MinPrice int
	MaxPrice int
	MinArea  float64
	MaxArea  float64
	Rooms    []int64
	MinScore int

	MinUndergroundPlace int
	MinKitchenArea      float64
	MinFloor            int
	MaxFloor            int
	MinCeilingHeight    float64
	ChildrenRequired    bool
	PetsRequired        bool
	DishwasherRequired  bool
	ConditionerRequired bool
	MinRenovation       string
	BalconyRequired     bool
	BathroomType        string

	PeriodSeconds    int
	LastReportSentAt time.Time
}

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

// GetDueReportSubscriptions returns active report subscriptions whose period
// has elapsed since their last report.
func (db *DB) GetDueReportSubscriptions(ctx context.Context) ([]ReportSubscription, error) {
	rows, err := db.pool.Query(ctx,
		`SELECT id, chat_id, deal_type, region, min_price, max_price, min_area, max_area, rooms, min_score,
		        min_underground_place, min_kitchen_area, min_floor, max_floor, min_ceiling_height,
		        children_required, pets_required, dishwasher_required, conditioner_required,
		        min_renovation, balcony_required, bathroom_type, metro_stations,
		        period_seconds, last_report_sent_at
		 FROM report_user_subscriptions
		 WHERE is_active = TRUE
		   AND last_report_sent_at + (period_seconds * INTERVAL '1 second') <= NOW()`)
	if err != nil {
		return nil, fmt.Errorf("querying due report subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []ReportSubscription
	for rows.Next() {
		var s ReportSubscription
		if err := rows.Scan(&s.ID, &s.ChatID, &s.DealType, &s.Region, &s.MinPrice, &s.MaxPrice,
			&s.MinArea, &s.MaxArea, &s.Rooms, &s.MinScore,
			&s.MinUndergroundPlace, &s.MinKitchenArea, &s.MinFloor, &s.MaxFloor, &s.MinCeilingHeight,
			&s.ChildrenRequired, &s.PetsRequired, &s.DishwasherRequired, &s.ConditionerRequired,
			&s.MinRenovation, &s.BalconyRequired, &s.BathroomType, &s.MetroStations,
			&s.PeriodSeconds, &s.LastReportSentAt); err != nil {
			return nil, fmt.Errorf("scanning report subscription: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// UpdateLastReportSentAt advances the "last sent" timestamp for a report
// subscription after a report has been (attempted to be) delivered.
func (db *DB) UpdateLastReportSentAt(ctx context.Context, subscriptionID int, at time.Time) error {
	_, err := db.pool.Exec(ctx,
		`UPDATE report_user_subscriptions SET last_report_sent_at = $1 WHERE id = $2`,
		at, subscriptionID)
	if err != nil {
		return fmt.Errorf("updating last_report_sent_at: %w", err)
	}
	return nil
}

// DeactivateReportSubscriptionByID sets is_active = FALSE for the given
// report subscription, scoped to chatID.
func (db *DB) DeactivateReportSubscriptionByID(ctx context.Context, chatID int64, subscriptionID int) (bool, error) {
	tag, err := db.pool.Exec(ctx,
		`UPDATE report_user_subscriptions SET is_active = FALSE
		 WHERE id = $1 AND chat_id = $2 AND is_active = TRUE`,
		subscriptionID, chatID)
	if err != nil {
		return false, fmt.Errorf("deactivating report subscription: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// GetFlatsForReport returns flats matching the subscription's straightforward
// (SQL-expressible) filters, parsed within (since, until]. Filters that
// aren't easily expressed in SQL (renovation rank, balcony-or-loggia,
// bathroom type, underground place with "0 = no data") are applied
// afterwards in Go — see filter.Apply.
func (db *DB) GetFlatsForReport(ctx context.Context, sub ReportSubscription, since, until time.Time) ([]model.FlatRecord, error) {
	rooms := sub.Rooms
	if rooms == nil {
		rooms = []int64{}
	}

	rows, err := db.pool.Query(ctx,
		`SELECT id, link, parsed_at, price, flat_score, underground_score, underground_place, underground_distance_info,
		        underground_stations,
		        room_number, total_area, living_area, kitchen_area, floor, max_floor, deposit, deposit_months, comission,
		        renovation, is_apartments, loggia_count, balcony_count, windows_view,
		        separated_bathroom_count, combined_bathroom_count,
		        has_dishwasher, has_conditioner, children_allowed, pets_allowed, last_updated, ceiling_height,
		        building_entrances_number, building_apartments_number, building_elevators_number,
		        region, deal_type, sale_type, mortgage_allowed, is_new_building, new_building_name,
		        is_by_homeowner, demolished_in_moscow_program
		 FROM flats_history
		 WHERE deal_type = $1
		   AND region = $2
		   AND parsed_at > $3 AND parsed_at <= $4
		   AND ($5 = 0 OR price >= $5)
		   AND ($6 = 0 OR price <= $6)
		   AND ($7 = 0 OR total_area >= $7)
		   AND ($8 = 0 OR total_area <= $8)
		   AND ($9 = 0 OR flat_score >= $9)
		   AND (cardinality($10::bigint[]) = 0 OR room_number = ANY($10::bigint[]))
		   AND ($11 = 0 OR kitchen_area >= $11)
		   AND ($12 = 0 OR floor >= $12)
		   AND ($13 = 0 OR floor <= $13)
		   AND ($14 = 0 OR ceiling_height >= $14)
		   AND (NOT $15 OR children_allowed)
		   AND (NOT $16 OR pets_allowed)
		   AND (NOT $17 OR has_dishwasher)
		   AND (NOT $18 OR has_conditioner)
		 ORDER BY flat_score DESC`,
		sub.DealType, sub.Region, since, until,
		sub.MinPrice, sub.MaxPrice,
		sub.MinArea, sub.MaxArea,
		sub.MinScore,
		rooms,
		sub.MinKitchenArea,
		sub.MinFloor, sub.MaxFloor,
		sub.MinCeilingHeight,
		sub.ChildrenRequired, sub.PetsRequired, sub.DishwasherRequired, sub.ConditionerRequired,
	)
	if err != nil {
		return nil, fmt.Errorf("querying flats for report: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[model.FlatRecord])
}
