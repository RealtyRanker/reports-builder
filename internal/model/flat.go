package model

import "time"

// FlatRecord mirrors a full flats_history row (db tags match column names,
// used with pgx.RowToStructByName).
type FlatRecord struct {
	ID                        int       `db:"id"`
	Link                      string    `db:"link"`
	ParsedAt                  time.Time `db:"parsed_at"`
	Price                     int       `db:"price"`
	FlatScore                 int       `db:"flat_score"`
	UndergroundScore          float64   `db:"underground_score"`
	UndergroundPlace          int       `db:"underground_place"`
	UndergroundDistanceInfo   string    `db:"underground_distance_info"`
	UndergroundStations       []string  `db:"underground_stations"`
	RoomNumber                int       `db:"room_number"`
	TotalArea                 float64   `db:"total_area"`
	LivingArea                float64   `db:"living_area"`
	KitchenArea               float64   `db:"kitchen_area"`
	Floor                     int       `db:"floor"`
	MaxFloor                  int       `db:"max_floor"`
	Deposit                   int       `db:"deposit"`
	DepositMonths             int       `db:"deposit_months"`
	Comission                 int       `db:"comission"`
	Renovation                string    `db:"renovation"`
	IsApartments              bool      `db:"is_apartments"`
	LoggiaCount               int       `db:"loggia_count"`
	BalconyCount              int       `db:"balcony_count"`
	WindowsView               string    `db:"windows_view"`
	SeparatedBathroomCount    int       `db:"separated_bathroom_count"`
	CombinedBathroomCount     int       `db:"combined_bathroom_count"`
	HasDishwasher             bool      `db:"has_dishwasher"`
	HasConditioner            bool      `db:"has_conditioner"`
	ChildrenAllowed           bool      `db:"children_allowed"`
	PetsAllowed               bool      `db:"pets_allowed"`
	LastUpdated               string    `db:"last_updated"`
	CeilingHeight             float64   `db:"ceiling_height"`
	BuildingEntrancesNumber   int       `db:"building_entrances_number"`
	BuildingApartmentsNumber  int       `db:"building_apartments_number"`
	BuildingElevatorsNumber   int       `db:"building_elevators_number"`
	Region                    int       `db:"region"`
	DealType                  string    `db:"deal_type"`
	SaleType                  string    `db:"sale_type"`
	MortgageAllowed           bool      `db:"mortgage_allowed"`
	IsNewBuilding             bool      `db:"is_new_building"`
	NewBuildingName           string    `db:"new_building_name"`
	IsByHomeowner             bool      `db:"is_by_homeowner"`
	DemolishedInMoscowProgram bool      `db:"demolished_in_moscow_program"`
}
