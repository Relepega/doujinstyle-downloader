package dsdl

import (
	"log"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
)

func restoreDB(sqlite *db.SQLiteDB) *db.SQLiteDB {
	// check if app has perms to open a sorage-based db, fallbacks to memory db
	err := sqlite.Open()
	if err != nil {
		log.Printf(
			"Failed to open selected db: \"%s\", falling back to the in-memory db",
			sqlite.Name(),
		)

		return db.NewSQLite(true)
	}

	log.Println("DB: Using", sqlite.Name())

	// restore saved data
	count, err := sqlite.Count()
	if err != nil {
		log.Panicf("TQWrapper: constructor: cannot evaluate db count: %v", err)
	}

	if count != 0 {
		running, err := sqlite.GetAllWithState(states.TASK_STATE_RUNNING)
		if err != nil {
			log.Panicf("TQWrapper: DB error: %v", err)
		}

		queued, err := sqlite.GetAllWithState(states.TASK_STATE_QUEUED)
		if err != nil {
			log.Panicf("TQWrapper: DB error: %v", err)
		}

		for _, t := range running {
			t.DownloadState = states.TASK_STATE_QUEUED
			t.Err = nil
			sqlite.SetState(t, states.TASK_STATE_QUEUED)
		}

		for _, t := range queued {
			t.DownloadState = states.TASK_STATE_QUEUED
			t.Err = nil
			sqlite.SetState(t, states.TASK_STATE_QUEUED)
		}
	}

	return sqlite
}
