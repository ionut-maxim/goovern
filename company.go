package goovern

type Company struct {
	TaxID                string  `json:"tax_id" db:"tax_id"`
	Name                 string  `json:"name" db:"name"`
	RegistrationCode     string  `json:"registration_code" db:"registration_code"`
	RegistrationDate     string  `json:"registration_date" db:"registration_date"`
	EUID                 string  `json:"euid" db:"euid"`
	LegalForm            string  `json:"legal_form" db:"legal_form"`
	Country              string  `json:"country" db:"country"`
	County               string  `json:"county" db:"county"`
	Locality             string  `json:"locality" db:"locality"`
	StreetName           string  `json:"street_name" db:"street_name"`
	StreetNumber         string  `json:"street_number" db:"street_number"`
	Building             string  `json:"building" db:"building"`
	Staircase            string  `json:"staircase" db:"staircase"`
	Floor                string  `json:"floor" db:"floor"`
	Apartment            string  `json:"apartment" db:"apartment"`
	PostalCode           string  `json:"postal_code" db:"postal_code"`
	Sector               string  `json:"sector" db:"sector"`
	AddressDetails       string  `json:"address_details" db:"address_details"`
	Website              string  `json:"website" db:"website"`
	ParentCompanyCountry string  `json:"parent_company_country" db:"parent_company_country"`
	Rank                 float32 `json:"rank" db:"rank"`
}
