package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wxtrans/internal/models"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
	path string
}

func DefaultPath() string {
	home, err := os.UserConfigDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, "wxtrans")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "ledger.db")
}

func Open(path string) (*DB, error) {
	if path == "" {
		path = DefaultPath()
	}
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	db := &DB{conn: conn, path: path}
	if err := db.migrate(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) Path() string { return db.path }

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	trans_time TEXT NOT NULL,
	trans_type TEXT NOT NULL DEFAULT '',
	counterparty TEXT NOT NULL DEFAULT '',
	product TEXT NOT NULL DEFAULT '',
	direction TEXT NOT NULL DEFAULT '',
	amount REAL NOT NULL,
	payment_method TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT '',
	trans_no TEXT NOT NULL UNIQUE,
	merchant_no TEXT NOT NULL DEFAULT '',
	remark TEXT NOT NULL DEFAULT '',
	imported_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_transactions_trans_time ON transactions(trans_time);
CREATE INDEX IF NOT EXISTS idx_transactions_direction ON transactions(direction);
CREATE INDEX IF NOT EXISTS idx_transactions_trans_type ON transactions(trans_type);
CREATE INDEX IF NOT EXISTS idx_transactions_counterparty ON transactions(counterparty);
`
	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) InsertTransaction(tx *models.Transaction) (bool, error) {
	res, err := db.conn.Exec(`
INSERT OR IGNORE INTO transactions
(trans_time, trans_type, counterparty, product, direction, amount,
 payment_method, status, trans_no, merchant_no, remark)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tx.TransTime.Format(time.RFC3339),
		tx.TransType,
		tx.Counterparty,
		tx.Product,
		tx.Direction,
		tx.Amount,
		tx.PaymentMethod,
		tx.Status,
		tx.TransNo,
		tx.MerchantNo,
		tx.Remark,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (db *DB) CountAll() (int, error) {
	var n int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&n)
	return n, err
}

func (db *DB) buildWhere(filter models.SearchFilter) (string, []any) {
	var parts []string
	var args []any

	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		like := "%" + kw + "%"
		parts = append(parts, `(trans_type LIKE ? OR counterparty LIKE ? OR product LIKE ?
			OR payment_method LIKE ? OR status LIKE ? OR trans_no LIKE ? OR merchant_no LIKE ? OR remark LIKE ?)`)
		for i := 0; i < 8; i++ {
			args = append(args, like)
		}
	}
	if filter.DateFrom != nil {
		parts = append(parts, `trans_time >= ?`)
		args = append(args, filter.DateFrom.Format(time.RFC3339))
	}
	if filter.DateTo != nil {
		end := filter.DateTo.Add(24*time.Hour - time.Second)
		parts = append(parts, `trans_time <= ?`)
		args = append(args, end.Format(time.RFC3339))
	}
	if d := strings.TrimSpace(filter.Direction); d != "" && d != "全部" {
		parts = append(parts, `direction = ?`)
		args = append(args, d)
	}
	if t := strings.TrimSpace(filter.TransType); t != "" {
		parts = append(parts, `trans_type = ?`)
		args = append(args, t)
	}

	where := ""
	if len(parts) > 0 {
		where = "WHERE " + strings.Join(parts, " AND ")
	}
	return where, args
}

func scanTransaction(row interface{ Scan(...any) error }) (*models.Transaction, error) {
	var tx models.Transaction
	var transTime, importedAt string
	err := row.Scan(
		&tx.ID, &transTime, &tx.TransType, &tx.Counterparty, &tx.Product,
		&tx.Direction, &tx.Amount, &tx.PaymentMethod, &tx.Status,
		&tx.TransNo, &tx.MerchantNo, &tx.Remark, &importedAt,
	)
	if err != nil {
		return nil, err
	}
	tx.TransTime, _ = time.Parse(time.RFC3339, transTime)
	tx.ImportedAt, _ = time.Parse(time.RFC3339, importedAt)
	return &tx, nil
}

func (db *DB) Search(filter models.SearchFilter) ([]models.Transaction, int, error) {
	where, args := db.buildWhere(filter)

	var total int
	countSQL := `SELECT COUNT(*) FROM transactions ` + where
	if err := db.conn.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 500
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
SELECT id, trans_time, trans_type, counterparty, product, direction, amount,
       payment_method, status, trans_no, merchant_no, remark, imported_at
FROM transactions ` + where + `
ORDER BY trans_time DESC
LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), limit, offset)

	rows, err := db.conn.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []models.Transaction
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *tx)
	}
	return list, total, rows.Err()
}

func (db *DB) Summary(filter models.SearchFilter) (*models.Summary, error) {
	where, args := db.buildWhere(filter)
	query := `
SELECT
  COUNT(*),
  COALESCE(SUM(CASE WHEN direction='收入' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN direction='支出' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN direction='收入' THEN amount ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN direction='支出' THEN amount ELSE 0 END), 0)
FROM transactions ` + where

	var s models.Summary
	err := db.conn.QueryRow(query, args...).Scan(
		&s.TotalCount, &s.IncomeCount, &s.ExpenseCount, &s.IncomeAmount, &s.ExpenseAmount,
	)
	if err != nil {
		return nil, err
	}
	s.NetAmount = s.IncomeAmount - s.ExpenseAmount
	return &s, nil
}

func (db *DB) SummaryByType(filter models.SearchFilter) ([]models.TypeSummary, error) {
	where, args := db.buildWhere(filter)
	query := `
SELECT trans_type, direction, COUNT(*), COALESCE(SUM(amount), 0)
FROM transactions ` + where + `
GROUP BY trans_type, direction
ORDER BY SUM(amount) DESC`

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.TypeSummary
	for rows.Next() {
		var item models.TypeSummary
		if err := rows.Scan(&item.TransType, &item.Direction, &item.Count, &item.Amount); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (db *DB) SummaryByMonth(filter models.SearchFilter) ([]models.MonthSummary, error) {
	where, args := db.buildWhere(filter)
	query := `
SELECT strftime('%Y-%m', trans_time) AS month,
  COALESCE(SUM(CASE WHEN direction='收入' THEN amount ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN direction='支出' THEN amount ELSE 0 END), 0)
FROM transactions ` + where + `
GROUP BY month
ORDER BY month DESC`

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.MonthSummary
	for rows.Next() {
		var item models.MonthSummary
		if err := rows.Scan(&item.Month, &item.IncomeAmount, &item.ExpenseAmount); err != nil {
			return nil, err
		}
		item.NetAmount = item.IncomeAmount - item.ExpenseAmount
		list = append(list, item)
	}
	return list, rows.Err()
}

func (db *DB) DistinctTypes() ([]string, error) {
	rows, err := db.conn.Query(`SELECT DISTINCT trans_type FROM transactions ORDER BY trans_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

func FormatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
