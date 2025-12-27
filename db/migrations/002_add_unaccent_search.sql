-- +goose Up
-- +goose StatementBegin

-- Enable the unaccent extension for diacritic-insensitive search
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Create an IMMUTABLE wrapper function for unaccent
-- This allows us to use it in a GENERATED column
CREATE OR REPLACE FUNCTION immutable_unaccent(text)
RETURNS text
LANGUAGE sql
IMMUTABLE PARALLEL SAFE STRICT
AS $$
    SELECT unaccent($1);
$$;

-- Drop the old generated column and index
ALTER TABLE companies DROP COLUMN IF EXISTS name_tsvector;
DROP INDEX IF EXISTS idx_companies_name_tsvector;

-- Recreate the tsvector column with unaccent for diacritic-insensitive search
ALTER TABLE companies
ADD COLUMN name_tsvector tsvector
GENERATED ALWAYS AS (to_tsvector('romanian', immutable_unaccent(name))) STORED;

-- Recreate the GIN index
CREATE INDEX idx_companies_name_tsvector
    ON companies USING GIN (name_tsvector);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to the original tsvector without unaccent
ALTER TABLE companies DROP COLUMN IF EXISTS name_tsvector;
DROP INDEX IF EXISTS idx_companies_name_tsvector;

ALTER TABLE companies
ADD COLUMN name_tsvector tsvector
GENERATED ALWAYS AS (to_tsvector('romanian', name)) STORED;

CREATE INDEX idx_companies_name_tsvector
    ON companies USING GIN (name_tsvector);

DROP FUNCTION IF EXISTS immutable_unaccent(text);
DROP EXTENSION IF EXISTS unaccent;

-- +goose StatementEnd
