// Package csvreport renders a slice of flats into a CSV file (results.csv).
package csvreport

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"github.com/asmisnik/reports-builder/internal/model"
)

var header = []string{
	"id", "link", "parsed_at", "price", "flat_score",
	"underground_score", "underground_place", "underground_distance_info",
	"room_number", "total_area", "living_area", "kitchen_area",
	"floor", "max_floor",
	"deposit", "deposit_months", "comission",
	"renovation", "is_apartments", "loggia_count", "balcony_count", "windows_view",
	"separated_bathroom_count", "combined_bathroom_count",
	"has_dishwasher", "has_conditioner", "children_allowed", "pets_allowed",
	"last_updated", "ceiling_height",
	"building_entrances_number", "building_apartments_number", "building_elevators_number",
	"region", "deal_type",
	"sale_type", "mortgage_allowed", "is_new_building", "new_building_name",
	"is_by_homeowner", "demolished_in_moscow_program",
}

// Build renders flats as a CSV file (with a header row) and returns its bytes.
func Build(flats []model.FlatRecord) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(header); err != nil {
		return nil, err
	}
	for _, f := range flats {
		row := []string{
			strconv.Itoa(f.ID), f.Link, f.ParsedAt.Format("2006-01-02 15:04:05"),
			strconv.Itoa(f.Price), strconv.Itoa(f.FlatScore),
			formatFloat(f.UndergroundScore), strconv.Itoa(f.UndergroundPlace), f.UndergroundDistanceInfo,
			strconv.Itoa(f.RoomNumber), formatFloat(f.TotalArea), formatFloat(f.LivingArea), formatFloat(f.KitchenArea),
			strconv.Itoa(f.Floor), strconv.Itoa(f.MaxFloor),
			strconv.Itoa(f.Deposit), strconv.Itoa(f.DepositMonths), strconv.Itoa(f.Comission),
			f.Renovation, formatBool(f.IsApartments), strconv.Itoa(f.LoggiaCount), strconv.Itoa(f.BalconyCount), f.WindowsView,
			strconv.Itoa(f.SeparatedBathroomCount), strconv.Itoa(f.CombinedBathroomCount),
			formatBool(f.HasDishwasher), formatBool(f.HasConditioner), formatBool(f.ChildrenAllowed), formatBool(f.PetsAllowed),
			f.LastUpdated, formatFloat(f.CeilingHeight),
			strconv.Itoa(f.BuildingEntrancesNumber), strconv.Itoa(f.BuildingApartmentsNumber), strconv.Itoa(f.BuildingElevatorsNumber),
			strconv.Itoa(f.Region), f.DealType,
			f.SaleType, formatBool(f.MortgageAllowed), formatBool(f.IsNewBuilding), f.NewBuildingName,
			formatBool(f.IsByHomeowner), formatBool(f.DemolishedInMoscowProgram),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatBool(v bool) string {
	if v {
		return "да"
	}
	return "нет"
}
