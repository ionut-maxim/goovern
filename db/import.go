package db

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/jackc/pgx/v5"

	"github.com/ionut-maxim/goovern/ckan"
	"github.com/ionut-maxim/goovern/csv"
)

// Import requires a transactional client to work properly because we are using a `TEMP` table
func (c *DB) Import(ctx context.Context, db Tx, resource ckan.Resource, data io.Reader) error {
	logger := c.logger.With("resource_name", resource.Name, "table", resource.Name)

	config, ok := importConfigs[resource.Name]
	if !ok {
		err := fmt.Errorf("no import configuration registered for resource: %s", resource.Name)
		logger.Error("Import configuration not found", "error", err)
		return err
	}

	logger.Debug("Reading CSV headers")
	reader := csv.NewReader(data, '^')

	headers, err := reader.Read()
	if err != nil {
		logger.Error("Failed to read CSV headers", "error", err)
		return fmt.Errorf("reading CSV headers: %w", err)
	}

	stripBOM(headers)
	logger.Debug("CSV headers parsed", "column_count", len(headers))

	source := csv.NewSource(reader)

	tx, err := db.Begin(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	logger.Info("Starting data import to database")
	bytes, rowsAffected, err := importWithConfig(ctx, tx, headers, source, config, logger)
	if err != nil {
		logger.Error("Import failed", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.Info("Import completed successfully",
		"rows_inserted", rowsAffected,
		"bytes_processed", humanize.Bytes(uint64(bytes)))

	return nil
}

func importWithConfig(ctx context.Context, tx Tx, headers []string, source *csv.Source, config ImportConfig, logger *slog.Logger) (bytes int64, rows int64, err error) {
	normalizeHeaders(headers, config.ColumnMapping)

	tempTable := fmt.Sprintf("%s_%d", config.TempTableName, rand.IntN(5000))
	logger.Debug("Creating temporary table", "temp_table", tempTable, "target_table", config.TableName)

	createTableQuery := fmt.Sprintf(
		`CREATE TEMP TABLE %s (LIKE %s INCLUDING DEFAULTS EXCLUDING GENERATED) ON COMMIT DROP`,
		pgx.Identifier{tempTable}.Sanitize(),
		pgx.Identifier{config.TableName}.Sanitize(),
	)
	if _, err = tx.Exec(ctx, createTableQuery); err != nil {
		return 0, 0, fmt.Errorf("creating temp table: %w", err)
	}

	// Add progress callback to log every 10,000 rows
	source.WithProgressCallback(func(rowCount int64) {
		logger.Debug("Import progress", "rows_processed", humanize.Comma(rowCount))
	}, 10000)

	logger.Debug("Copying data to temporary table")
	bytes, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{tempTable},
		headers,
		source,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("copying to temp table: %w", err)
	}

	logger.Info("Data copied to temporary table", "bytes", humanize.Bytes(uint64(bytes)))

	logger.Debug("Inserting data into target table", "target_table", config.TableName)
	// Build column list from headers (which are the non-generated columns in temp table)
	columnList := ""
	for i, header := range headers {
		if i > 0 {
			columnList += ", "
		}
		columnList += pgx.Identifier{header}.Sanitize()
	}

	insertQuery := fmt.Sprintf(
		`INSERT INTO %s (%s) SELECT %s FROM %s ON CONFLICT DO NOTHING`,
		pgx.Identifier{config.TableName}.Sanitize(),
		columnList,
		columnList,
		pgx.Identifier{tempTable}.Sanitize(),
	)
	result, err := tx.Exec(ctx, insertQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("inserting from temp table: %w", err)
	}

	logger.Debug("Data inserted", "rows_affected", result.RowsAffected())
	return bytes, result.RowsAffected(), nil
}

// stripBOM removes the UTF-8 BOM from the first header if present
func stripBOM(headers []string) {
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\ufeff")
	}
}

// normalizeHeaders applies a column mapping and converts headers to lowercase
func normalizeHeaders(headers []string, columnMapping map[string]string) {
	for i := range headers {
		csvHeader := strings.ToLower(headers[i])
		if dbColumn, ok := columnMapping[csvHeader]; ok {
			headers[i] = dbColumn
		} else {
			headers[i] = csvHeader
		}
	}
}
