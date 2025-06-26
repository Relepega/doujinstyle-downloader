package db

import (
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

const TABLE_NAME string = "dsdl"

type SQliteDB[T task.Insertable] struct {
	DB[T]

	name string

	path string

	db *sqlx.DB
}

func NewSQliteDB[T task.Insertable](path string) DB[T] {
	return &SQliteDB[T]{
		name: "SQLite",
		path: path,
	}
}

func (sdb *SQliteDB[T]) Open() error {
	// db, err := sqlx.Connect("sqlite3", ":memory:")
	db, err := sqlx.Connect("sqlite3", sdb.path)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(2)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + TABLE_NAME + ` (
			ID STRING PRIMARY KEY,
			Aggregator STRING,
			Slug STRING,
			AggregatorPageURL STRING,
			FilehostUrl STRING,
			DisplayName STRING,
			Filename STRING,
			DownloadState INTEGER,
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

func (sdb *SQliteDB[T]) Insert(nv T) (string, error) {
	s, err := sdb.db.Prepare(`
		INSERT INTO ` + TABLE_NAME + ` (
			ID,
			Aggregator,
			Slug,
			AggregatorPageURL,
			FilehostUrl,
			DisplayName,
			Filename,
			DownloadState,
			Err
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nv.GetID(), err
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
		nv.GetErrMsg(),
	)
	return nv.GetID(), err
}

func (sdb *SQliteDB[T]) Get(slug string) (T, error) {
	var entry T
	t := task.NewTask("")

	if reflect.TypeOf(entry) != reflect.TypeOf(t) {
		return entry, fmt.Errorf(
			`DB: TypeError: Wanted "%v", got "%v"`,
			reflect.TypeOf(t),
			reflect.TypeOf(entry),
		)
	}

	if err := sdb.db.Get(
		t,
		`SELECT * FROM `+TABLE_NAME+` WHERE ID = ? OR Slug LIKE ? LIMIT 1`,
		slug, "%"+slug+"%",
	); err != nil {
		return entry, err
	}

	t.Err = fmt.Errorf("%s", t.DBErr)

	entry = any(t).(T)

	return entry, nil
}

func (sdb *SQliteDB[T]) GetAll() ([]T, error) {
	var entry []T

	rows, err := sdb.db.Query(`SELECT * FROM ` + TABLE_NAME)
	if err != nil {
		return entry, err
	}
	defer rows.Close()

	for rows.Next() {
		t := task.NewTask("")

		err := rows.Scan(
			&t.Id,
			&t.Aggregator,
			&t.Slug,
			&t.AggregatorPageURL,
			&t.FilehostUrl,
			&t.DisplayName,
			&t.Filename,
			&t.DownloadState,
			&t.DBErr,
		)
		if err != nil {
			return entry, err
		}

		t.Err = fmt.Errorf("%s", t.DBErr)

		entry = append(entry, any(t).(T))
	}

	return entry, nil
}

func (sdb *SQliteDB[T]) Remove(nv T) error {
	_, err := sdb.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE id = ?`, nv.GetID())

	return err
}

func (sdb *SQliteDB[T]) RemoveFromState(completionState int) (int, error) {
	res, err := sdb.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE DownloadState = ?`, completionState)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()

	return int(count), err
}

func (sdb *SQliteDB[T]) RemoveAll() error {
	_, err := sdb.db.Exec(`DELETE FROM ` + TABLE_NAME)

	return err
}

func (sdb *SQliteDB[T]) ResetFromCompletionState(completionState int) (int, error) {
	if completionState == states.TASK_STATE_RUNNING {
		return 0, fmt.Errorf("Cannot reset running tasks")
	}

	res, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE DownloadState = ?`,
		states.TASK_STATE_QUEUED,
		completionState,
	)
	if err != nil {
		return 0, err
	}

	count64, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	count := int(count64)

	return count, nil
}

func (sdb *SQliteDB[T]) getStateInt(nv T) (int, error) {
	var stateID int

	rows, err := sdb.db.Query(`SELECT DownloadState FROM `+TABLE_NAME+` WHERE ID = ?`, nv.GetID())
	if err != nil {
		return stateID, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&stateID)
		if err != nil {
			return stateID, err
		}
	} else {
		return stateID, rows.Err()
	}

	return stateID, nil
}

func (sdb *SQliteDB[T]) GetState(nv T) (string, error) {
	stateID, err := sdb.getStateInt(nv)
	if err != nil {
		return "", err
	}

	state := states.GetStateStr(stateID)

	return state, nil
}

func (sdb *SQliteDB[T]) SetState(nv T, newState int) error {
	_, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		newState,
		nv.GetID(),
	)

	return err
}

func (sdb *SQliteDB[T]) AdvanceState(nv T) (int, error) {
	state, err := sdb.getStateInt(nv)

	if state >= states.MaxCompletionState() {
		return state, fmt.Errorf("Cannot advance the status of this task anymore")
	}

	state = state + 1

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		state,
		nv.GetID(),
	)

	return state, err
}

func (sdb *SQliteDB[T]) RegressState(nv T) (int, error) {
	state, err := sdb.getStateInt(nv)

	if state <= 0 {
		return state, fmt.Errorf("Cannot regress the status of this task anymore")
	}

	state = state - 1

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		state,
		nv.GetID(),
	)

	return state, err
}

func (sdb *SQliteDB[T]) ResetState(nv T) (int, error) {
	_, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		states.TASK_STATE_QUEUED,
		nv.GetID(),
	)

	return states.TASK_STATE_QUEUED, err
}

func (sdb *SQliteDB[T]) Drop(table string) error {
	_, err := sdb.db.Exec(`DROP TABLE ` + TABLE_NAME)

	return err
}
