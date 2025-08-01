package config

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	db, err := sql.Open("sqlite", "./archive.db")
	if err != nil {
		panic("打开数据库失败: " + err.Error())
	}

	// 建表
	_, err = db.Exec(`
				CREATE TABLE IF NOT EXISTS archive_records (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					title TEXT NOT NULL,
					execute_time TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
				
				CREATE TABLE IF NOT EXISTS autoBgi_config (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					autobgi_key VARCHAR(255),
					autobgi_value VARCHAR(255)
				);
				
				CREATE TABLE IF NOT EXISTS talent_domains (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					domain_name TEXT NOT NULL,
					weekday INTEGER NOT NULL,
					material_name TEXT NOT NULL
				);

				CREATE TABLE IF NOT EXISTS weapon_domains (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					domain_name TEXT NOT NULL,
					weekday INTEGER NOT NULL,
					material_name TEXT NOT NULL
				);
`)
	if err != nil {
		panic("建表失败: " + err.Error())
	}

	// 插入 autoBgi_config 默认数据，若不存在则插入
	_, err = db.Exec(`
			INSERT INTO autoBgi_config (autobgi_key, autobgi_value)
			SELECT 'BackupUserTime', datetime('now','+8 hours')
			WHERE NOT EXISTS (
				SELECT 1 FROM autoBgi_config WHERE autobgi_key = 'BackupUserTime'
			);
`)
	if err != nil {
		panic("插入autoBgi_config默认配置失败: " + err.Error())
	}

	// 查询 talent_domains 表是否有数据
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM talent_domains`).Scan(&count)
	if err != nil {
		panic("查询 talent_domains 失败: " + err.Error())
	}

	// 如果空，批量插入默认天赋本数据
	if count == 0 {
		_, err = db.Exec(`
					INSERT INTO talent_domains (domain_name, weekday, material_name) VALUES
					('忘却之峡', 1, '自由'),
					('忘却之峡', 4, '自由'),
					('忘却之峡', 2, '抗争'),
					('忘却之峡', 5, '抗争'),
					('忘却之峡', 3, '诗文'),
					('忘却之峡', 6, '诗文'),
					('太山府', 1, '繁荣'),
					('太山府', 4, '繁荣'),
					('太山府', 2, '勤劳'),
					('太山府', 5, '勤劳'),
					('太山府', 3, '黄金'),
					('太山府', 6, '黄金'),
					('堇色之庭', 1, '浮世'),
					('堇色之庭', 4, '浮世'),
					('堇色之庭', 2, '风雅'),
					('堇色之庭', 5, '风雅'),
					('堇色之庭', 3, '天光'),
					('堇色之庭', 6, '天光'),
					('昏识塔', 1, '诤言'),
					('昏识塔', 4, '诤言'),
					('昏识塔', 2, '巧思'),
					('昏识塔', 5, '巧思'),
					('昏识塔', 3, '笃行'),
					('昏识塔', 6, '笃行'),
					('苍白的遗荣', 1, '公平'),
					('苍白的遗荣', 4, '公平'),
					('苍白的遗荣', 2, '正义'),
					('苍白的遗荣', 5, '正义'),
					('苍白的遗荣', 3, '秩序'),
					('苍白的遗荣', 6, '秩序'),
					('蕴火的幽墟', 1, '角逐'),
					('蕴火的幽墟', 4, '角逐'),
					('蕴火的幽墟', 2, '焚燔'),
					('蕴火的幽墟', 5, '焚燔'),
					('蕴火的幽墟', 3, '纷争'),
					('蕴火的幽墟', 6, '纷争');
							`)
		if err != nil {
			panic("插入talent_domains默认数据失败: " + err.Error())
		}

	}

	// ================= 武器本默认数据 =================
	var count2 int
	err = db.QueryRow(`SELECT COUNT(*) FROM weapon_domains`).Scan(&count2)
	if err != nil {
		panic("查询 weapon_domains 失败: " + err.Error())
	}

	fmt.Println(count2)

	if count2 == 0 {
		_, err = db.Exec(`INSERT INTO weapon_domains (domain_name, weekday, material_name) VALUES
		-- 塞西莉亚苗圃
		('塞西莉亚苗圃', 1, '高塔孤王'), ('塞西莉亚苗圃', 4, '高塔孤王'),
		('塞西莉亚苗圃', 2, '凛风奔狼'), ('塞西莉亚苗圃', 5, '凛风奔狼'),
		('塞西莉亚苗圃', 3, '狮牙斗士'), ('塞西莉亚苗圃', 6, '狮牙斗士'),
		-- 震雷连山密宫
		('震雷连山密宫', 1, '孤云寒林'), ('震雷连山密宫', 4, '孤云寒林'),
		('震雷连山密宫', 2, '雾海云间'), ('震雷连山密宫', 5, '雾海云间'),
		('震雷连山密宫', 3, '漆黑陨铁'), ('震雷连山密宫', 6, '漆黑陨铁'),
		-- 砂流之庭
		('砂流之庭', 1, '远海夷地'), ('砂流之庭', 4, '远海夷地'),
		('砂流之庭', 2, '鸣神御灵'), ('砂流之庭', 5, '鸣神御灵'),
		('砂流之庭', 3, '今昔剧画'), ('砂流之庭', 6, '今昔剧画'),
		-- 有顶塔
		('有顶塔', 1, '谧林涓露'), ('有顶塔', 4, '谧林涓露'),
		('有顶塔', 2, '绿洲花园'), ('有顶塔', 5, '绿洲花园'),
		('有顶塔', 3, '烈日威权'), ('有顶塔', 6, '烈日威权'),
		-- 深潮的余响
		('深潮的余响', 1, '悠古弦音'), ('深潮的余响', 4, '悠古弦音'),
		('深潮的余响', 2, '纯圣滴露'), ('深潮的余响', 5, '纯圣滴露'),
		('深潮的余响', 3, '无垢之海'), ('深潮的余响', 6, '无垢之海'),
		-- 深古瞭望所
		('深古瞭望所', 1, '贡祭炽心'), ('深古瞭望所', 4, '贡祭炽心'),
		('深古瞭望所', 2, '谵妄圣主'), ('深古瞭望所', 5, '谵妄圣主'),
		('深古瞭望所', 3, '神合秘烟'), ('深古瞭望所', 6, '神合秘烟');`)
		if err != nil {
			panic("插入weapon_domains默认数据失败: " + err.Error())
		}
	}

	DB = db

}
