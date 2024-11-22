package migrations

import (
	"gorm.io/gorm"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

type Migration struct {
	Version string
	UpSQL   string
	DownSQL string
}

func ApplyMigrations(db *gorm.DB) error {
	migrationsDir := "migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var migrations []Migration
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			version := strings.TrimSuffix(strings.TrimSuffix(file.Name(), ".up.sql"), "00000")
			upSQL, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				return err
			}
			downSQL, err := ioutil.ReadFile(filepath.Join(migrationsDir, strings.Replace(file.Name(), ".up.sql", ".down.sql", 1)))
			if err != nil {
				return err
			}
			migrations = append(migrations, Migration{
				Version: version,
				UpSQL:   string(upSQL),
				DownSQL: string(downSQL),
			})
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	for _, migration := range migrations {
		var count int64
		db.Raw("SELECT COUNT(*) FROM migrations WHERE version = ?", migration.Version).Scan(&count)
		if count == 0 {
			if err := db.Exec(migration.UpSQL).Error; err != nil {
				return err
			}
			if err := db.Exec("INSERT INTO migrations (version) VALUES (?)", migration.Version).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
