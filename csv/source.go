package csv

import "io"

type ProgressCallback func(rowCount int64)

func NewSource(data *GoovernReader) *Source {
	return &Source{
		reader: data,
	}
}

// WithProgressCallback sets a callback that will be invoked every 'interval' rows
func (c *Source) WithProgressCallback(callback ProgressCallback, interval int64) *Source {
	c.progressCallback = callback
	c.progressInterval = interval
	return c
}

// Source implements pgx.CopyFromSource for streaming CSV data
type Source struct {
	reader           Reader
	currentRow       []string
	readErr          error
	rowCount         int64
	progressCallback ProgressCallback
	progressInterval int64
}

func (c *Source) Next() bool {
	c.currentRow, c.readErr = c.reader.Read()
	if c.readErr == io.EOF {
		return false
	}
	if c.readErr == nil {
		c.rowCount++
		// Call progress callback at intervals
		if c.progressCallback != nil && c.progressInterval > 0 {
			if c.rowCount%c.progressInterval == 0 {
				c.progressCallback(c.rowCount)
			}
		}
	}
	return c.readErr == nil
}

func (c *Source) Values() ([]any, error) {
	if c.readErr != nil {
		return nil, c.readErr
	}

	values := make([]any, len(c.currentRow))
	for i, v := range c.currentRow {
		values[i] = v
	}
	return values, nil
}

func (c *Source) Err() error {
	if c.readErr == io.EOF {
		return nil
	}
	return c.readErr
}
