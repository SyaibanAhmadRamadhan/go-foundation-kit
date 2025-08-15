package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	sqlx "github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/appdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := sqlx.NewRDBMS(
		db,
		sqlx.WithStmtShardCount(16),
		sqlx.WithStmtJanitorInterval(15*time.Second),
		sqlx.WithStmtIdleTTL(20*time.Minute),
		sqlx.UseDebugNql(true),
	)

	ctx := context.Background()

	// Pastikan ada unique index di name, aman untuk banyak kali run.
	if _, err := r.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS users_name_uniq ON users(name)`); err != nil {
		log.Fatal("create unique index:", err)
	}
	// panggil dari main():
	// 1) pakai prepared statements (stmt cache)
	if err := testTxMode(ctx, r, r, true); err != nil {
		log.Fatal(err)
	}
	if err := testTxMode(ctx, r, r, true); err != nil {
		log.Fatal(err)
	}
	// 2) tanpa prepared statements (langsung)
	if err := testTxMode(ctx, r, r, false); err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}
func testTxMode(ctx context.Context, r sqlx.RDBMS, tx sqlx.Tx, useStmt bool) error {
	mode := "NON-STMT"
	query := r.QueryContext
	if useStmt {
		mode = "STMT"
		query = r.QueryStmtContext
	}

	insertSQL := `INSERT INTO users(name) VALUES($1)`
	countSQL := `SELECT COUNT(*) FROM users WHERE name = $1`

	fmt.Printf("\n==================== %s: success tx ====================\n", mode)
	nameOK := fmt.Sprintf("ok-%s-%d", strings.ToLower(mode), time.Now().UnixNano())

	// TX sukses: insert -> select (1) -> commit -> select luar (1)
	if err := tx.DoTxContext(ctx, nil, func(ctx context.Context, t sqlx.RDBMS) error {
		log.Println("[TX-OK] begin", mode)
		exec := t.ExecContext
		query := t.QueryContext
		if useStmt {
			mode = "STMT"
			exec = t.ExecStmtContext
			query = t.QueryStmtContext
		}
		if _, err := exec(ctx, insertSQL, nameOK); err != nil {
			return fmt.Errorf("[TX-OK] insert: %w", err)
		}
		var c int
		rows, err := query(ctx, countSQL, nameOK)
		if err != nil {
			return fmt.Errorf("[TX-OK] select: %w", err)
		}
		for rows.Next() {
			if err := rows.Scan(&c); err != nil {
				rows.Close()
				return err
			}
		}
		rows.Close()
		log.Printf("[TX-OK] %s count in-tx = %d\n", mode, c)
		if c != 1 {
			return fmt.Errorf("[TX-OK] expected count=1 in tx, got %d", c)
		}
		log.Println("[TX-OK] commit", mode)
		return nil
	}); err != nil {
		return err
	}

	var c1 int
	rows1, err := query(ctx, countSQL, nameOK)
	if err != nil {
		return err
	}
	for rows1.Next() {
		if err := rows1.Scan(&c1); err != nil {
			rows1.Close()
			return err
		}
	}
	rows1.Close()
	log.Printf("[TX-OK] %s count after commit (outside) = %d\n", mode, c1)

	fmt.Printf("==================== %s: failed tx ====================\n", mode)
	// TX gagal (duplikat) : insert -> insert (dupe) -> rollback -> select luar (0)
	nameFail := fmt.Sprintf("fail-%s-%d", strings.ToLower(mode), time.Now().UnixNano())

	err = tx.DoTxContext(ctx, nil, func(ctx context.Context, t sqlx.RDBMS) error {
		log.Println("[TX-FAIL] begin", mode)
		exec := t.ExecContext
		if useStmt {
			mode = "STMT"
			exec = t.ExecStmtContext
		}

		if _, err := exec(ctx, insertSQL, nameFail); err != nil {
			return fmt.Errorf("[TX-FAIL] first insert: %w", err)
		}
		if _, err := exec(ctx, insertSQL, nameFail); err != nil {
			log.Printf("[TX-FAIL] %s second insert expected error: %v\n", mode, err)
			return err // trigger rollback
		}
		log.Println("[TX-FAIL] should not reach here (expect unique error)", mode)
		return nil
	})
	if err == nil {
		return fmt.Errorf("[TX-FAIL] expected error on second insert, got nil")
	}
	log.Printf("[TX-FAIL] %s tx rolled back as expected: %v\n", mode, err)

	var c2 int
	rows2, err := query(ctx, countSQL, nameFail)
	if err != nil {
		return err
	}
	for rows2.Next() {
		if err := rows2.Scan(&c2); err != nil {
			rows2.Close()
			return err
		}
	}
	rows2.Close()
	log.Printf("[TX-FAIL] %s count after rollback (outside) = %d\n", mode, c2)
	if c2 != 0 {
		return fmt.Errorf("[TX-FAIL] expected count=0 after rollback, got %d", c2)
	}

	return nil
}
