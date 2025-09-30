package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dsn := "data/analytics.db"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS energy (
            timestamp TEXT NOT NULL,
            site TEXT NOT NULL,
            kwh REAL NOT NULL
        );`,
		`DELETE FROM energy;`,
		`INSERT INTO energy (timestamp, site, kwh) VALUES
            ('2024-01-01T00:00:00Z','A',10.5),
            ('2024-01-01T01:00:00Z','A',11.0),
            ('2024-01-01T00:00:00Z','B',7.2),
            ('2024-01-02T00:00:00Z','A',9.8),
            ('2024-01-02T00:00:00Z','B',8.1);`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			panic(fmt.Errorf("seed failed: %w", err))
		}
	}

	fmt.Println("Seeded data/analytics.db with table energy (5 rows)")
}

