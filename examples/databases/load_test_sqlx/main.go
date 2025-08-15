package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	sqlx "github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx"
	_ "github.com/lib/pq"
)

type benchResp struct {
	Mode       string        `json:"mode"`         // "stmt" / "nonstmt"
	WithTx     bool          `json:"with_tx"`      // true/false
	Op         string        `json:"op"`           // "select" / "insert"
	Loops      int           `json:"loops"`        // loops executed
	DurationNs int64         `json:"duration_ns"`  // total
	AvgPerLoop time.Duration `json:"avg_per_loop"` // avg duration
	RowsTotal  int           `json:"rows_total"`   // sum of rows scanned/affected across loops (kasar)
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/appdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0)
	defer db.Close()

	r := sqlx.NewRDBMS(
		db,
		sqlx.WithStmtShardCount(16),
		sqlx.WithStmtJanitorInterval(15*time.Second),
		sqlx.WithStmtIdleTTL(20*time.Minute),
		// sqlx.UseDebugNql(false), // on/off sesuai kebutuhan
	)

	ctx := context.Background()

	// safety: index unik untuk skenario insert
	if _, err := r.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS users_name_uniq ON users(name)`); err != nil {
		log.Fatal("create unique index:", err)
	}

	// health
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// SELECT endpoints
	http.HandleFunc("/stmt/select", func(w http.ResponseWriter, req *http.Request) {
		handleSelect(w, req, r, r, true)
	})
	http.HandleFunc("/nonstmt/select", func(w http.ResponseWriter, req *http.Request) {
		handleSelect(w, req, r, r, false)
	})

	// INSERT endpoints (pakai ON CONFLICT DO NOTHING agar aman saat load)
	http.HandleFunc("/stmt/insert", func(w http.ResponseWriter, req *http.Request) {
		handleInsert(w, req, r, r, true)
	})
	http.HandleFunc("/nonstmt/insert", func(w http.ResponseWriter, req *http.Request) {
		handleInsert(w, req, r, r, false)
	})

	http.HandleFunc("/stmt/select/heavy", func(w http.ResponseWriter, req *http.Request) {
		handleSelectHeavy(w, req, r, r, true)
	})
	http.HandleFunc("/nonstmt/select/heavy", func(w http.ResponseWriter, req *http.Request) {
		handleSelectHeavy(w, req, r, r, false)
	})

	http.HandleFunc("/stmt/insert/heavy", func(w http.ResponseWriter, req *http.Request) {
		handleInsertHeavy(w, req, r, r, true)
	})
	http.HandleFunc("/nonstmt/insert/heavy", func(w http.ResponseWriter, req *http.Request) {
		handleInsertHeavy(w, req, r, r, false)
	})

	addr := ":8080"
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

/********* Handlers *********/

func handleSelect(w http.ResponseWriter, req *http.Request, r sqlx.RDBMS, tx sqlx.Tx, useStmt bool) {
	ctx := req.Context()
	name := qStr(req, "name", "bench-user")
	loops := qInt(req, "loops", 1000)
	withTx := qBool(req, "tx", false)

	query := `SELECT COUNT(*) FROM users WHERE name = $1`

	start := time.Now()
	rowsTotal := 0

	run := func(exec sqlx.RDBMS) error {
		for i := 0; i < loops; i++ {
			var rows *sql.Rows
			var err error
			if useStmt {
				rows, err = exec.QueryStmtContext(ctx, query, name)
			} else {
				rows, err = exec.QueryContext(ctx, query, name)
			}
			if err != nil {
				return err
			}
			for rows.Next() {
				var c int
				if err := rows.Scan(&c); err != nil {
					rows.Close()
					return err
				}
				rowsTotal += c
			}
			rows.Close()
		}
		return nil
	}

	var err error
	if withTx {
		err = tx.DoTxContext(ctx, nil, func(ctx context.Context, tx sqlx.RDBMS) error { return run(tx) })
	} else {
		err = run(r)
	}
	dur := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	resp := benchResp{
		Mode:       ternary(useStmt, "stmt", "nonstmt"),
		WithTx:     withTx,
		Op:         "select",
		Loops:      loops,
		DurationNs: dur.Nanoseconds(),
		AvgPerLoop: time.Duration(dur.Nanoseconds() / int64(max(loops, 1))),
		RowsTotal:  rowsTotal,
	}
	writeJSON(w, resp)
}

func handleInsert(w http.ResponseWriter, req *http.Request, r sqlx.RDBMS, tx sqlx.Tx, useStmt bool) {
	ctx := req.Context()
	prefix := qStr(req, "prefix", "bench-ins")
	loops := qInt(req, "loops", 1000)
	withTx := qBool(req, "tx", false)
	keep := qBool(req, "keep", false) // kalau false, kita cleanup habis insert

	insertSQL := `INSERT INTO users(name) VALUES($1) ON CONFLICT(name) DO NOTHING`
	deleteSQL := `DELETE FROM users WHERE name = $1`

	start := time.Now()
	rowsTotal := 0

	run := func(exec sqlx.RDBMS) error {
		for i := 0; i < loops; i++ {
			name := prefix + "-" + randStr(8)
			var res sql.Result
			var err error
			if useStmt {
				res, err = exec.ExecStmtContext(ctx, insertSQL, name)
			} else {
				res, err = exec.ExecContext(ctx, insertSQL, name)
			}
			if err != nil {
				return err
			}
			if ra, e := res.RowsAffected(); e == nil {
				rowsTotal += int(ra)
			}
			if !keep {
				if useStmt {
					if _, err := exec.ExecStmtContext(ctx, deleteSQL, name); err != nil {
						return err
					}
				} else {
					if _, err := exec.ExecContext(ctx, deleteSQL, name); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	var err error
	if withTx {
		err = tx.DoTxContext(ctx, nil, func(ctx context.Context, tx sqlx.RDBMS) error { return run(tx) })
	} else {
		err = run(r)
	}
	dur := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	resp := benchResp{
		Mode:       ternary(useStmt, "stmt", "nonstmt"),
		WithTx:     withTx,
		Op:         "insert",
		Loops:      loops,
		DurationNs: dur.Nanoseconds(),
		AvgPerLoop: time.Duration(dur.Nanoseconds() / int64(max(loops, 1))),
		RowsTotal:  rowsTotal,
	}
	writeJSON(w, resp)
}

func handleSelectHeavy(w http.ResponseWriter, req *http.Request, r sqlx.RDBMS, tx sqlx.Tx, useStmt bool) {
	ctx := req.Context()
	qText := qStr(req, "q", "product")     // FTS
	brand := qStr(req, "brand", "brand17") // JSONB filter
	topK := qInt(req, "top_per_category", 10)
	limit := qInt(req, "limit", 200)
	offset := qInt(req, "offset", 0)
	withTx := qBool(req, "tx", false)

	sqlHeavy := `
WITH recent_orders AS (
  SELECT o.id, o.customer_id, o.created_at
  FROM orders o
  WHERE o.created_at >= now() - interval '90 days'
),
filtered_products AS (
  SELECT p.id, p.title, p.attrs
  FROM products p
  WHERE p.tsv @@ plainto_tsquery($1)
    AND (p.attrs ? 'brand' AND p.attrs->>'brand' = $2)
),
sales AS (
  SELECT ro.customer_id,
         oi.product_id,
         SUM(oi.quantity * oi.unit_price_cents) AS revenue_cents,
         COUNT(*) AS lines
  FROM recent_orders ro
  JOIN order_items oi ON oi.order_id = ro.id
  GROUP BY ro.customer_id, oi.product_id
),
joined AS (
  SELECT s.customer_id,
         s.product_id,
         s.revenue_cents,
         s.lines,
         fp.title,
         c.name AS category_name
  FROM sales s
  JOIN filtered_products fp ON fp.id = s.product_id
  JOIN product_categories pc ON pc.product_id = fp.id
  JOIN categories c ON c.id = pc.id
),
win AS (
  SELECT j.*,
         SUM(revenue_cents) OVER (PARTITION BY category_name) AS cat_revenue_cents,
         RANK() OVER (PARTITION BY category_name ORDER BY revenue_cents DESC) AS rnk
  FROM joined j
)
SELECT w.category_name,
       w.title,
       w.customer_id,
       w.revenue_cents,
       w.cat_revenue_cents,
       w.rnk,
       ev.payload->>'ref' AS last_ref,
       ev.payload->>'path' AS last_path
FROM win w
JOIN LATERAL (
  SELECT e.payload
  FROM events e
  WHERE e.customer_id = w.customer_id
  ORDER BY e.created_at DESC
  LIMIT 1
) ev ON true
WHERE w.rnk <= $3
ORDER BY w.cat_revenue_cents DESC, w.category_name, w.revenue_cents DESC
LIMIT $4 OFFSET $5;`

	start := time.Now()
	rowsTotal := 0

	run := func(exec sqlx.RDBMS) error {
		var rows *sql.Rows
		var err error
		args := []any{qText, brand, topK, limit, offset}
		if useStmt {
			rows, err = exec.QueryStmtContext(ctx, sqlHeavy, args...)
		} else {
			rows, err = exec.QueryContext(ctx, sqlHeavy, args...)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			// scan minimal saja untuk counting; kalau mau validasi isi, define struct
			var (
				categoryName string
				title        string
				customerID   int64
				revenueCents int64
				catRevenue   int64
				rnk          int
				lastRef      sql.NullString
				lastPath     sql.NullString
			)
			if err := rows.Scan(&categoryName, &title, &customerID, &revenueCents, &catRevenue, &rnk, &lastRef, &lastPath); err != nil {
				return err
			}
			rowsTotal++
		}
		return rows.Err()
	}

	var err error
	if withTx {
		err = tx.DoTxContext(ctx, nil, func(ctx context.Context, tx sqlx.RDBMS) error { return run(tx) })
	} else {
		err = run(r)
	}
	dur := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	resp := benchResp{
		Mode:       ternary(useStmt, "stmt", "nonstmt"),
		WithTx:     withTx,
		Op:         "select_heavy",
		Loops:      1, // satu query berat per hit; kalau mau, tambahkan loops param juga
		DurationNs: dur.Nanoseconds(),
		AvgPerLoop: dur,
		RowsTotal:  rowsTotal,
	}
	writeJSON(w, resp)
}

// /********* Heavy Insert Handler *********/

func handleInsertHeavy(w http.ResponseWriter, req *http.Request, r sqlx.RDBMS, tx sqlx.Tx, useStmt bool) {
	ctx := req.Context()

	// params
	nOrders := qInt(req, "orders", 1000)             // berapa order baru yang dibuat
	itemsPerOrder := qInt(req, "items_per_order", 3) // item per order (fixed; cukup berat krn LATERAL)
	qText := qStr(req, "q", "")                      // FTS pada products.tsv (optional)
	brand := qStr(req, "brand", "")                  // filter brand pada products.attrs->>'brand' (optional)
	withTx := qBool(req, "tx", false)                // jalankan dalam TX?
	keep := qBool(req, "keep", false)                // kalau false, auto-cleanup setelah insert

	// heavy insert:
	// 1) insert nOrders ke orders untuk random customers (90 hari terakhir opsional, di sini pakai now())
	// 2) tiap order, LATERAL pilih itemsPerOrder produk (random) yg match filter FTS/brand
	// 3) insert order_items
	// 4) optional cleanup: hapus item & order yang baru dibuat dalam statement yang sama (tanpa CASCADE)
	sqlHeavyInsert := `
WITH orders_ins AS (
  INSERT INTO orders(customer_id, status, created_at)
  SELECT c.id,
         (ARRAY['paid','shipped','cancelled'])[1+floor(random()*3)]::text,
         now()
  FROM (
    SELECT id FROM customers ORDER BY random() LIMIT $1
  ) AS c
  RETURNING id
),
items_ins AS (
  INSERT INTO order_items(order_id, product_id, quantity, unit_price_cents)
  SELECT o.id,
         p.id,
         1 + (random()*3)::int,
         500 + (random()*50000)::int
  FROM orders_ins o
  JOIN LATERAL (
    SELECT pr.id
    FROM products pr
    WHERE ($2 = '' OR pr.tsv @@ plainto_tsquery($2))
      AND ($3 = '' OR pr.attrs->>'brand' = $3)
    ORDER BY random()
    LIMIT $4
  ) AS p ON true
  RETURNING 1
),
del_items AS (
  DELETE FROM order_items oi
  USING orders_ins o
  WHERE oi.order_id = o.id AND $5::bool = false
  RETURNING 1
),
del_orders AS (
  DELETE FROM orders o
  USING orders_ins oi
  WHERE o.id = oi.id AND $5::bool = false
  RETURNING 1
)
SELECT
  (SELECT COUNT(*) FROM orders_ins) AS orders_inserted,
  (SELECT COUNT(*) FROM items_ins)  AS items_inserted,
  (SELECT COUNT(*) FROM del_items)  AS items_deleted,
  (SELECT COUNT(*) FROM del_orders) AS orders_deleted;`

	start := time.Now()

	type resCounts struct {
		ordersInserted int64
		itemsInserted  int64
		itemsDeleted   int64
		ordersDeleted  int64
	}
	var counts resCounts

	run := func(exec sqlx.RDBMS) error {
		var rows *sql.Rows
		var err error
		args := []any{nOrders, qText, brand, itemsPerOrder, keep}
		if useStmt {
			rows, err = exec.QueryStmtContext(ctx, sqlHeavyInsert, args...)
		} else {
			rows, err = exec.QueryContext(ctx, sqlHeavyInsert, args...)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(
				&counts.ordersInserted,
				&counts.itemsInserted,
				&counts.itemsDeleted,
				&counts.ordersDeleted,
			); err != nil {
				return err
			}
		}
		return rows.Err()
	}

	var err error
	if withTx {
		err = tx.DoTxContext(ctx, nil, func(ctx context.Context, tx sqlx.RDBMS) error { return run(tx) })
	} else {
		err = run(r)
	}

	dur := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// RowsTotal aku isi total rows yang benar2 diinsert (orders+items); kalau keep=false, tetap laporkan insertednya.
	rowsTotal := int(counts.ordersInserted + counts.itemsInserted)

	resp := benchResp{
		Mode:       ternary(useStmt, "stmt", "nonstmt"),
		WithTx:     withTx,
		Op:         "insert_heavy",
		Loops:      1,
		DurationNs: dur.Nanoseconds(),
		AvgPerLoop: dur,
		RowsTotal:  rowsTotal,
	}
	writeJSON(w, resp)
}

/********* Utils *********/

func qStr(r *http.Request, key, def string) string {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	return v
}
func qInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
func qBool(r *http.Request, key string, def bool) bool {
	if v := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(key))); v != "" {
		return v == "1" || v == "true" || v == "yes"
	}
	return def
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func ternary[T any](b bool, x, y T) T {
	if b {
		return x
	}
	return y
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
