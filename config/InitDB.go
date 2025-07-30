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
		CREATE TABLE if NOT EXISTS autoBgi_config (
			id integer primary key AUTOINCREMENT,
			autobgi_key varchar(255),
			autobgi_value varchar(255)
		);
    `)
	if err != nil {
		panic("建表失败: " + err.Error())
	}

	// 插入默认数据
	_, err = db.Exec(`
		INSERT INTO autoBgi_config (autobgi_key, autobgi_value)
		SELECT 'BackupUserTime', datetime('now','+8 hours')
		WHERE NOT EXISTS (
			SELECT 1 FROM autoBgi_config WHERE autobgi_key = 'BackupUserTime'
		);
	`)
	if err != nil {
		panic("插入默认配置失败: " + err.Error())
	}

	return db
}
