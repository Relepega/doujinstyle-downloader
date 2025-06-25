package db

import (
	"database/sql"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

const TABLE_NAME string = "dsdl"

type SQliteDB[T task.Insertable] struct {
	DB[T]

	name string

	fn string

	db *sqlx.DB
}

func NewSQliteDB[T task.Insertable]() DB[T] {
	fn := filepath.Join(".", "Database")
	appUtils.MkdirAll(fn)

	fn = filepath.Join(fn, "default.db")

	return &SQliteDB[T]{
		name: "SQLite",
		fn:   fn,
	}
}

func (sdb *SQliteDB[T]) Open() error {
	// db, err := sql.Open("sqlite", sdb.fn)
	// if err != nil {
	// 	return err
	// }
	//
	// db.SetMaxOpenConns(2)
	//
	// if _, err := db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS ` + TABLE_NAME + ` (
	// 		Id STRING PRIMARY KEY,
	// 		Aggregator STRING,
	// 		Slug STRING,
	// 		AggregatorPageURL STRING,
	// 		FilehostUrl STRING,
	// 		DisplayName STRING,
	// 		Filename STRING,
	// 		DownloadState INTEGER,
	// 		Progress INTEGER,
	// 		Err STRING
	// 	);
	// `); err != nil {
	// 	return err
	// }
	//
	// sdb.db = db
	//
	// return nil

	// db, err := sqlx.Connect("sqlite3", ":memory:")
	db, err := sqlx.Connect("sqlite3", sdb.fn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(2)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + TABLE_NAME + ` (
			Id STRING PRIMARY KEY,
			Aggregator STRING,
			Slug STRING,
			AggregatorPageURL STRING,
			FilehostUrl STRING,
			DisplayName STRING,
			Filename STRING,
			DownloadState INTEGER,
			Progress INTEGER,
			Err STRING
		);
	`); err != nil {
		return err
	}

	sdb.db = db

	return nil
}

func (sdb *SQliteDB[T]) Close() error {
	return sdb.db.Close()
}

func (sdb *SQliteDB[T]) Name() string {
	return sdb.name
}

func (sdb *SQliteDB[T]) Count() (int, error) {
	var count int

	rows, err := sdb.db.Query(`SELECT COUNT(*) FROM ` + TABLE_NAME)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if !rows.Next() {
		return -1, rows.Err()
	}

	err = rows.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (sdb *SQliteDB[T]) CountFromState(completionState int) (int, error) {
	return 0, nil
}

func (sdb *SQliteDB[T]) Insert(nv T) error {
	s, err := sdb.db.Prepare(`
		INSERT INTO ` + TABLE_NAME + ` (
			Id,
			Aggregator,
			Slug,
			AggregatorPageURL,
			FilehostUrl,
			DisplayName,
			Filename,
			DownloadState,
			Progress,
			Err
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = s.Exec(
		nv.GetID(),
		nv.GetAggregator(),
		nv.GetSlug(),
		nv.GetAggregatorPageURL(),
		nv.GetFilehostUrl(),
		nv.GetDisplayName(),
		nv.GetFilename(),
		nv.GetDownloadState(),
		nv.GetProgress(),
		nv.GetErrMsg(),
	)
	return err
}

func (sdb *SQliteDB[T]) Get(slug string) (T, error) {
	var entry T

	if err := sdb.db.Get(
		&entry,
		`SELECT * FROM `+TABLE_NAME+` WHERE Id = ? OR Slug LIKE ? LIMIT 1;`,
		slug, "%"+slug+"%",
	); err != nil {
		return entry, err
	}

	return entry, sql.ErrNoRows
}

func (sdb *SQliteDB[T]) GetAll() ([]T, error) {
	return make([]T, 0), nil
}

func (sdb *SQliteDB[T]) Remove(nv T) error {
	return nil
}

func (sdb *SQliteDB[T]) RemoveFromState(completionState int) (int, error) {
	return 0, nil
}

func (sdb *SQliteDB[T]) RemoveAll() error {
	return nil
}

func (sdb *SQliteDB[T]) ResetFromCompletionState(completionState int) error {
	return nil
}

func (sdb *SQliteDB[T]) GetState(nv T) (string, error) {
	return "", nil
}

func (sdb *SQliteDB[T]) SetState(nv T, newState int) error {
	return nil
}

func (sdb *SQliteDB[T]) AdvanceState(nv T) (int, error) {
	return -1, nil
}

func (sdb *SQliteDB[T]) RegressState(nv T) (int, error) {
	return -1, nil
}

func (sdb *SQliteDB[T]) ResetState(nv T) (int, error) {
	return -1, nil
}

func (sdb *SQliteDB[T]) Drop(table string) error {
	_, err := sdb.db.Exec(`DROP TABLE ` + TABLE_NAME)

	return err
}
