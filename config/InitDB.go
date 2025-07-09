package config

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", "./archive.db")
	if err != nil {
		panic("打开数据库失败: " + err.Error())
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS archive_records (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            execute_time TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		panic("建表失败: " + err.Error())
	}

	return db
}
