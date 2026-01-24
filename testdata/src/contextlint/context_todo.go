package contextlint

import (
	"context"
	"database/sql"
)

// BadContextTODO demonstrates context.TODO() in production code.
func BadContextTODO(db *sql.DB) error {
	ctx := context.TODO() // want "AIL010: context.TODO\\(\\) used in non-test code"
	return db.PingContext(ctx)
}

// BadContextTODOInline demonstrates inline context.TODO() usage.
func BadContextTODOInline(db *sql.DB) error {
	return db.PingContext(context.TODO()) // want "AIL010: context.TODO\\(\\) used in non-test code"
}

// GoodContextParam demonstrates receiving context as parameter.
func GoodContextParam(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx) // OK - context passed in
}

// GoodContextBackground demonstrates context.Background() usage.
func GoodContextBackground(db *sql.DB) error {
	ctx := context.Background() // OK - Background is intentional
	return db.PingContext(ctx)
}

// GoodContextWithCancel demonstrates creating context with cancel.
func GoodContextWithCancel(db *sql.DB) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return db.PingContext(ctx) // OK - proper context usage
}
