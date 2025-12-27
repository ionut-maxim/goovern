package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/ionut-maxim/goovern"
)

//SELECT name, registration_code, ts_rank(name_tsvector, query) AS rank
//FROM companies,
//     to_tsquery('romanian', immutable_unaccent('maxim & ionut & persoana & fizica & autorizata')) query
//WHERE name_tsvector @@ query
//ORDER BY rank DESC
//LIMIT 20;

func (c *DB) Search(ctx context.Context, db Querier, searchTerm string, limit int) ([]goovern.Company, error) {
	// Split search term into words and join with & for AND search
	words := strings.Fields(searchTerm)
	if len(words) == 0 {
		return nil, fmt.Errorf("search term cannot be empty")
	}
	tsQuery := strings.Join(words, " | ")

	// Check if search term looks like a CUI (tax ID) - typically numeric with optional RO prefix
	// If it's purely numeric or starts with RO, prioritize exact CUI match
	// Otherwise, use full-text search on name and also check if CUI contains the search term
	query := `
		SELECT
			registration_code,
			name,
			COALESCE(tax_id, '') as tax_id,
			COALESCE(registration_date, '') as registration_date,
			COALESCE(euid, '') as euid,
			COALESCE(legal_form, '') as legal_form,
			COALESCE(country, '') as country,
			COALESCE(county, '') as county,
			COALESCE(locality, '') as locality,
			COALESCE(street_name, '') as street_name,
			COALESCE(street_number, '') as street_number,
			COALESCE(building, '') as building,
			COALESCE(staircase, '') as staircase,
			COALESCE(floor, '') as floor,
			COALESCE(apartment, '') as apartment,
			COALESCE(postal_code, '') as postal_code,
			COALESCE(sector, '') as sector,
			COALESCE(address_details, '') as address_details,
			COALESCE(website, '') as website,
			COALESCE(parent_company_country, '') as parent_company_country,
			CASE
				WHEN tax_id ILIKE $1 || '%' THEN 1.0
				ELSE ts_rank(name_tsvector, query)
			END AS rank
		FROM companies
		LEFT JOIN LATERAL (
			SELECT to_tsquery('romanian', immutable_unaccent($2)) AS query
		) q ON true
		WHERE
			tax_id ILIKE '%' || $1 || '%'
			OR name_tsvector @@ query
		ORDER BY rank DESC, name
		LIMIT $3
	`

	rows, err := db.Query(ctx, query, searchTerm, tsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []goovern.Company
	for rows.Next() {
		var comp goovern.Company
		err = rows.Scan(
			&comp.RegistrationCode,
			&comp.Name,
			&comp.TaxID,
			&comp.RegistrationDate,
			&comp.EUID,
			&comp.LegalForm,
			&comp.Country,
			&comp.County,
			&comp.Locality,
			&comp.StreetName,
			&comp.StreetNumber,
			&comp.Building,
			&comp.Staircase,
			&comp.Floor,
			&comp.Apartment,
			&comp.PostalCode,
			&comp.Sector,
			&comp.AddressDetails,
			&comp.Website,
			&comp.ParentCompanyCountry,
			&comp.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, comp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}
