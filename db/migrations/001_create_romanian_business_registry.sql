-- +goose Up
-- +goose StatementBegin
-- Migration: Create Romanian Business Registry tables
-- Description: Tables for Romanian ONRC (Trade Registry Office) data

-- ============================================================================
-- Nomenclature Tables (lookup/reference data)
-- ============================================================================

-- CAEN (Romanian economic activities classification) versions
CREATE TABLE IF NOT EXISTS caen_versions (
    code        INT PRIMARY KEY,
    description TEXT NOT NULL
);

-- CAEN codes (economic activities)
CREATE TABLE IF NOT EXISTS caen_codes (
    section      TEXT NOT NULL,
    subsection   TEXT NOT NULL,
    division     TEXT NOT NULL,
    "group"      TEXT NOT NULL,
    class        TEXT NOT NULL,
    name         TEXT NOT NULL,
    caen_version INT  NOT NULL,
    PRIMARY KEY (section, subsection, division, "group", class, caen_version),
    FOREIGN KEY (caen_version) REFERENCES caen_versions(code)
);

-- Company status types (nomenclature)
CREATE TABLE IF NOT EXISTS company_statuses (
    code INT  PRIMARY KEY,
    name TEXT NOT NULL
);

-- ============================================================================
-- Main Companies Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS companies (
    registration_code      TEXT PRIMARY KEY,
    name                   TEXT NOT NULL,
    tax_id                 TEXT,
    registration_date      TEXT,
    euid                   TEXT,
    legal_form             TEXT,
    country                TEXT,
    county                 TEXT,
    locality               TEXT,
    street_name            TEXT,
    street_number          TEXT,
    building               TEXT,
    staircase              TEXT,
    floor                  TEXT,
    apartment              TEXT,
    postal_code            TEXT,
    sector                 TEXT,
    address_details        TEXT,
    website                TEXT,
    parent_company_country TEXT
);

-- Add generated tsvector column for full-text search on company names
ALTER TABLE companies
ADD COLUMN name_tsvector tsvector
GENERATED ALWAYS AS (to_tsvector('romanian', name)) STORED;

-- Indexes on companies table
CREATE INDEX IF NOT EXISTS idx_companies_tax_id
    ON companies(tax_id);

CREATE INDEX IF NOT EXISTS idx_companies_euid
    ON companies(euid);

CREATE INDEX IF NOT EXISTS idx_companies_name_tsvector
    ON companies USING GIN (name_tsvector);

-- ============================================================================
-- Company-Related Tables (depend on companies)
-- ============================================================================

-- Authorized economic activities for each company
CREATE TABLE IF NOT EXISTS authorized_activities (
    registration_code    TEXT NOT NULL,
    authorized_caen_code TEXT NOT NULL,
    caen_version         INT  NOT NULL,
    PRIMARY KEY (registration_code, authorized_caen_code, caen_version),
    FOREIGN KEY (registration_code) REFERENCES companies(registration_code) ON DELETE CASCADE,
    FOREIGN KEY (caen_version) REFERENCES caen_versions(code)
);

CREATE INDEX IF NOT EXISTS idx_authorized_activities_registration_code
    ON authorized_activities(registration_code);

-- Company status history
CREATE TABLE IF NOT EXISTS company_status_history (
    registration_code TEXT NOT NULL,
    status_code       INT  NOT NULL,
    PRIMARY KEY (registration_code, status_code),
    FOREIGN KEY (registration_code) REFERENCES companies(registration_code) ON DELETE CASCADE,
    FOREIGN KEY (status_code) REFERENCES company_statuses(code)
);

CREATE INDEX IF NOT EXISTS idx_company_status_history_registration_code
    ON company_status_history(registration_code);

-- Legal representatives
CREATE TABLE IF NOT EXISTS legal_representatives (
    id                SERIAL PRIMARY KEY,
    registration_code TEXT NOT NULL,
    authorized_person TEXT NOT NULL,
    role              TEXT,
    birth_date        TEXT,
    birth_locality    TEXT,
    birth_county      TEXT,
    birth_country     TEXT,
    locality          TEXT,
    county            TEXT,
    country           TEXT,
    FOREIGN KEY (registration_code) REFERENCES companies(registration_code) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_legal_representatives_registration_code
    ON legal_representatives(registration_code);

-- Family business representatives (întreprindere familială)
CREATE TABLE IF NOT EXISTS family_business_representatives (
    id                SERIAL PRIMARY KEY,
    registration_code TEXT NOT NULL,
    name              TEXT NOT NULL,
    birth_date        TEXT,
    birth_locality    TEXT,
    birth_county      TEXT,
    birth_country     TEXT,
    role              TEXT,
    FOREIGN KEY (registration_code) REFERENCES companies(registration_code) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_family_business_representatives_registration_code
    ON family_business_representatives(registration_code);

-- Foreign branches (sucursale in other EU member states)
CREATE TABLE IF NOT EXISTS foreign_branches (
    id                SERIAL PRIMARY KEY,
    registration_code TEXT NOT NULL,
    unit_type         TEXT,
    branch_name       TEXT,
    euid              TEXT,
    tax_code          TEXT,
    country           TEXT,
    FOREIGN KEY (registration_code) REFERENCES companies(registration_code) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_foreign_branches_registration_code
    ON foreign_branches(registration_code);

CREATE INDEX IF NOT EXISTS idx_foreign_branches_euid
    ON foreign_branches(euid);

-- ============================================================================
-- CKAN Resources Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS resources (
    id                     UUID PRIMARY KEY,
    package_id             UUID        NOT NULL,
    name                   TEXT        NOT NULL,
    description            TEXT,
    url                    TEXT        NOT NULL,
    url_type               TEXT,
    format                 TEXT,
    mimetype               TEXT,
    mimetype_inner         TEXT,
    size                   BIGINT,
    hash                   TEXT,
    state                  TEXT,
    position               INTEGER,
    resource_type          TEXT,

    -- Timestamps
    created                TIMESTAMPTZ,
    last_modified          TIMESTAMPTZ,
    cache_last_updated     TIMESTAMPTZ,

    -- Cache/archiver info
    cache_url              TEXT,
    datagovro_download_url TEXT,

    -- JSON fields for complex data
    qa                     JSONB,
    archiver               JSONB,

    -- Flags
    datastore_active       BOOLEAN     DEFAULT FALSE,

    -- Metadata
    revision_id            UUID,

    -- Audit timestamps
    created_at             TIMESTAMPTZ DEFAULT NOW(),
    updated_at             TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes on resources table
CREATE INDEX IF NOT EXISTS idx_resources_package_id
    ON resources(package_id);

CREATE INDEX IF NOT EXISTS idx_resources_format
    ON resources(format);

CREATE INDEX IF NOT EXISTS idx_resources_state
    ON resources(state);

-- GIN indexes on JSONB fields for efficient querying
CREATE INDEX IF NOT EXISTS idx_resources_qa
    ON resources USING GIN (qa);

CREATE INDEX IF NOT EXISTS idx_resources_archiver
    ON resources USING GIN (archiver);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS foreign_branches CASCADE;
DROP TABLE IF EXISTS family_business_representatives CASCADE;
DROP TABLE IF EXISTS legal_representatives CASCADE;
DROP TABLE IF EXISTS company_status_history CASCADE;
DROP TABLE IF EXISTS authorized_activities CASCADE;
DROP TABLE IF EXISTS companies CASCADE;
DROP TABLE IF EXISTS company_statuses CASCADE;
DROP TABLE IF EXISTS caen_codes CASCADE;
DROP TABLE IF EXISTS caen_versions CASCADE;
DROP TABLE IF EXISTS resources CASCADE;

-- +goose StatementEnd
