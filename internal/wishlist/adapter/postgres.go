package adapter

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fanuelson/wishlist-api/internal/wishlist/domain"
)

type PostgresAdapter struct {
	pool *pgxpool.Pool
}

func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{pool: pool}
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (a *PostgresAdapter) querier(ctx context.Context) querier {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return a.pool
}

func (a *PostgresAdapter) WithinCustomerLock(ctx context.Context, customerID domain.CustomerID, fn func(ctx context.Context) error) error {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`SELECT pg_advisory_xact_lock(hashtext($1))`,
		string(customerID),
	); err != nil {
		return fmt.Errorf("acquire customer lock: %w", err)
	}

	if err := fn(withTx(ctx, tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (a *PostgresAdapter) CountByCustomer(ctx context.Context, customerID domain.CustomerID) (int, error) {
	var count int
	if err := a.querier(ctx).QueryRow(ctx,
		`SELECT count(*) FROM wishlist_items WHERE customer_id = $1`,
		string(customerID),
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return count, nil
}

func (a *PostgresAdapter) Insert(ctx context.Context, item domain.Item) (bool, error) {
	tag, err := a.querier(ctx).Exec(ctx,
		`INSERT INTO wishlist_items (customer_id, product_id, added_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (customer_id, product_id) DO NOTHING`,
		string(item.CustomerID), string(item.ProductID), item.AddedAt,
	)
	if err != nil {
		return false, fmt.Errorf("insert item: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (a *PostgresAdapter) Delete(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) (bool, error) {
	tag, err := a.querier(ctx).Exec(ctx,
		`DELETE FROM wishlist_items WHERE customer_id = $1 AND product_id = $2`,
		string(customerID), string(productID),
	)
	if err != nil {
		return false, fmt.Errorf("delete item: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (a *PostgresAdapter) Exists(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) (bool, error) {
	var exists bool
	if err := a.querier(ctx).QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM wishlist_items
			WHERE customer_id = $1 AND product_id = $2
		)`,
		string(customerID), string(productID),
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check existence: %w", err)
	}
	return exists, nil
}

func (a *PostgresAdapter) FindAllByCustomer(ctx context.Context, customerID domain.CustomerID) ([]domain.Item, error) {
	rows, err := a.querier(ctx).Query(ctx,
		`SELECT customer_id, product_id, added_at
		 FROM wishlist_items
		 WHERE customer_id = $1
		 ORDER BY added_at DESC`,
		string(customerID),
	)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.CustomerID, &item.ProductID, &item.AddedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return items, nil
}
