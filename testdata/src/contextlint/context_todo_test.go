package contextlint

import (
	"context"
	"database/sql"
	"testing"
)

// TestContextTODO demonstrates that context.TODO() is OK in test files.
func TestContextTODO(t *testing.T) {
	var db *sql.DB
	ctx := context.TODO() // OK - in test file
	_ = db.PingContext(ctx)
}
