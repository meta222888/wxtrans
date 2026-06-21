package database

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	settingsKeyPasswordHash = "password_hash"
	defaultPassword         = "eee333"
	bcryptCost              = bcrypt.DefaultCost
)

var (
	ErrWrongPassword    = errors.New("密码错误")
	ErrPasswordTooShort = errors.New("新密码至少 4 位")
	ErrPasswordMismatch = errors.New("两次输入的新密码不一致")
)

func (db *DB) initPassword() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
)`)
	if err != nil {
		return err
	}

	var exists int
	err = db.conn.QueryRow(
		`SELECT COUNT(*) FROM settings WHERE key = ?`, settingsKeyPasswordHash,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return db.setPasswordHash(defaultPassword)
	}
	return nil
}

func (db *DB) setPasswordHash(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(`
INSERT INTO settings (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		settingsKeyPasswordHash, string(hash),
	)
	return err
}

func (db *DB) VerifyPassword(password string) error {
	var hash string
	err := db.conn.QueryRow(
		`SELECT value FROM settings WHERE key = ?`, settingsKeyPasswordHash,
	).Scan(&hash)
	if err != nil {
		return fmt.Errorf("读取密码失败: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return ErrWrongPassword
	}
	return nil
}

func (db *DB) ChangePassword(oldPassword, newPassword, confirmPassword string) error {
	if len(newPassword) < 4 {
		return ErrPasswordTooShort
	}
	if newPassword != confirmPassword {
		return ErrPasswordMismatch
	}
	if err := db.VerifyPassword(oldPassword); err != nil {
		return err
	}
	return db.setPasswordHash(newPassword)
}
