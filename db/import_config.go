package db

type ImportConfig struct {
	TableName     string            // Destination table name
	TempTableName string            // Temp table prefix (will have random number appended)
	ColumnMapping map[string]string // CSV header -> DB column mapping
}

var importConfigs = map[string]ImportConfig{
	"N_VERSIUNE_CAEN.CSV": {
		TableName:     "caen_versions",
		TempTableName: "caen_versions",
		ColumnMapping: map[string]string{
			"cod":       "code",
			"descriere": "description",
		},
	},
	"N_CAEN.CSV": {
		TableName:     "caen_codes",
		TempTableName: "caen_codes",
		ColumnMapping: map[string]string{
			"sectiunea":     "section",
			"subsectiunea":  "subsection",
			"diviziunea":    "division",
			"grupa":         "group",
			"clasa":         "class",
			"denumire":      "name",
			"versiune_caen": "caen_version",
		},
	},
	"N_STARE_FIRMA.CSV": {
		TableName:     "company_statuses",
		TempTableName: "company_statuses",
		ColumnMapping: map[string]string{
			"cod":      "code",
			"denumire": "name",
		},
	},
	"OD_FIRME.CSV": {
		TableName:     "companies",
		TempTableName: "companies",
		ColumnMapping: map[string]string{
			"denumire":           "name",
			"cui":                "tax_id",
			"cod_inmatriculare":  "registration_code",
			"data_inmatriculare": "registration_date",
			"euid":               "euid",
			"forma_juridica":     "legal_form",
			"adr_tara":           "country",
			"adr_judet":          "county",
			"adr_localitate":     "locality",
			"adr_den_strada":     "street_name",
			"adr_nr_strada":      "street_number",
			"adr_bloc":           "building",
			"adr_scara":          "staircase",
			"adr_etaj":           "floor",
			"adr_apartament":     "apartment",
			"adr_cod_postal":     "postal_code",
			"adr_sector":         "sector",
			"adr_completare":     "address_details",
			"web":                "website",
			"tara_firma_mama":    "parent_company_country",
		},
	},
	"OD_CAEN_AUTORIZAT.CSV": {
		TableName:     "authorized_activities",
		TempTableName: "authorized_activities",
		ColumnMapping: map[string]string{
			"cod_inmatriculare":  "registration_code",
			"cod_caen_autorizat": "authorized_caen_code",
			"ver_caen_autorizat": "caen_version",
		},
	},
	"OD_STARE_FIRMA.CSV": {
		TableName:     "company_status_history",
		TempTableName: "company_status_history",
		ColumnMapping: map[string]string{
			"cod_inmatriculare": "registration_code",
			"cod":               "status_code",
		},
	},
	"OD_REPREZENTANTI_LEGALI.CSV": {
		TableName:     "legal_representatives",
		TempTableName: "legal_representatives",
		ColumnMapping: map[string]string{
			"cod_inmatriculare":      "registration_code",
			"persoana_imputernicita": "authorized_person",
			"calitate":               "role",
			"data_nastere":           "birth_date",
			"localitate_nastere":     "birth_locality",
			"judet_nastere":          "birth_county",
			"tara_nastere":           "birth_country",
			"localitate":             "locality",
			"judet":                  "county",
			"tara":                   "country",
		},
	},
	"OD_REPREZENTANTI_IF.CSV": {
		TableName:     "family_business_representatives",
		TempTableName: "family_business_representatives",
		ColumnMapping: map[string]string{
			"cod_inmatriculare":  "registration_code",
			"nume":               "name",
			"data_nastere":       "birth_date",
			"localitate_nastere": "birth_locality",
			"judet_nastere":      "birth_county",
			"tara_nastere":       "birth_country",
			"calitate":           "role",
		},
	},
	"OD_SUCURSALE_ALTE_STATE_MEMBRE.CSV": {
		TableName:     "foreign_branches",
		TempTableName: "foreign_branches",
		ColumnMapping: map[string]string{
			"cod_inmatriculare":  "registration_code",
			"tip_unitate":        "unit_type",
			"denumire_sucursala": "branch_name",
			"euid":               "euid",
			"cod_fiscal":         "tax_code",
			"tara":               "country",
		},
	},
}
