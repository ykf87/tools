package ipquality

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteCache struct {
	db *sql.DB
}

func NewSQLiteCache(path string) (*SQLiteCache, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.Exec(`
    CREATE TABLE IF NOT EXISTS ip_cache (
        ip TEXT PRIMARY KEY,
        score INTEGER,
        type TEXT,
        allow INTEGER,
        reason TEXT,
        updated_at INTEGER
    )`)

	return &SQLiteCache{db: db}, nil
}

func (s *SQLiteCache) Get(ip string) (*Decision, bool) {
	row := s.db.QueryRow("SELECT score,type,allow,reason FROM ip_cache WHERE ip=?", ip)

	var score int
	var t string
	var allow int
	var reason string

	if err := row.Scan(&score, &t, &allow, &reason); err != nil {
		return nil, false
	}

	return &Decision{
		Score:  score,
		Type:   IPType(t),
		Allow:  allow == 1,
		Reason: reason,
	}, true
}

func (s *SQLiteCache) Set(ip string, d *Decision) {
	s.db.Exec(`
    INSERT OR REPLACE INTO ip_cache(ip,score,type,allow,reason,updated_at)
    VALUES(?,?,?,?,?,strftime('%s','now'))
    `, ip, d.Score, d.Type, boolToInt(d.Allow), d.Reason)
}
