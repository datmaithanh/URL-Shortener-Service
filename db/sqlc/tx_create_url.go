package db

import (
	"context"
	"database/sql"
)

type CreateUrlTxParams struct {
	CreateUrlParams
	AfterCreate func(q *Queries, url *Url) (Url, error)
}

type CreateUrlTxResult struct {
	Url Url
}

func (store *SQLStore) CreateUrlTx(ctx context.Context, arg CreateUrlTxParams) (CreateUrlTxResult, error) {
	var result CreateUrlTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		var err error
		existingUrl, err := q.GetUrlByOriginalUrl(ctx, arg.OriginalUrl)
		if err == nil {
			result.Url = existingUrl
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
		
		result.Url, err = q.CreateUrl(ctx, arg.CreateUrlParams)
		if err != nil {
			return err
		}

		result.Url, err = arg.AfterCreate(q, &result.Url)
		if err != nil {
			return err
		}
		return nil
	})

	return result, err
}
