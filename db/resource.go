package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ionut-maxim/goovern/ckan"
)

func (c *DB) SaveResource(ctx context.Context, tx Tx, resource ckan.Resource) error {
	q := `
	INSERT INTO resources (
		id, package_id, name, description, url, url_type, format, mimetype, mimetype_inner,
		size, hash, state, position, resource_type, created, last_modified, cache_last_updated,
		cache_url, datagovro_download_url, datastore_active, revision_id
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
	)
	ON CONFLICT (id) DO UPDATE SET
		package_id = EXCLUDED.package_id,
		name = EXCLUDED.name,
		description = EXCLUDED.description,
		url = EXCLUDED.url,
		url_type = EXCLUDED.url_type,
		format = EXCLUDED.format,
		mimetype = EXCLUDED.mimetype,
		mimetype_inner = EXCLUDED.mimetype_inner,
		size = EXCLUDED.size,
		hash = EXCLUDED.hash,
		state = EXCLUDED.state,
		position = EXCLUDED.position,
		resource_type = EXCLUDED.resource_type,
		created = EXCLUDED.created,
		last_modified = EXCLUDED.last_modified,
		cache_last_updated = EXCLUDED.cache_last_updated,
		cache_url = EXCLUDED.cache_url,
		datagovro_download_url = EXCLUDED.datagovro_download_url,
		datastore_active = EXCLUDED.datastore_active,
		revision_id = EXCLUDED.revision_id
	`

	_, err := tx.Exec(ctx, q,
		resource.Id,
		resource.PackageId,
		resource.Name,
		resource.Description,
		resource.Url,
		resource.UrlType,
		resource.Format,
		resource.Mimetype,
		resource.MimetypeInner,
		resource.Size,
		resource.Hash,
		resource.State,
		resource.Position,
		resource.ResourceType,
		resource.Created,
		resource.LastModified,
		resource.CacheLastUpdated,
		resource.CacheUrl,
		resource.DatagovroDownloadUrl,
		resource.DatastoreActive,
		resource.RevisionId,
	)

	return err
}

func (c *DB) Resource(ctx context.Context, tx Tx, id uuid.UUID) (ckan.Resource, bool, error) {
	var res ckan.Resource
	q := `
	SELECT
		id, package_id, name, description, url, url_type, format, mimetype, mimetype_inner,
		size, hash, state, position, resource_type, created, last_modified, cache_last_updated,
		cache_url, datagovro_download_url, datastore_active, revision_id
	FROM resources
	WHERE id = $1
	`

	rows, err := tx.Query(ctx, q, id)
	if err != nil {
		return res, false, err
	}
	defer rows.Close()

	res, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[ckan.Resource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, false, nil
		}
		return res, false, err
	}

	return res, true, nil
}
