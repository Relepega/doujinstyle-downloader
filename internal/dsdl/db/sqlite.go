package db

import (
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

const (
	TABLE_NAME string = "dsdl"
)

type SQliteDB struct {
	name string

	path string

	db *sqlx.DB
}

func NewSQLite(inMemory bool, dbPath string) *SQliteDB {
	path := filepath.Join(".", "Database")
	name := "File-Based SQLite DB"

	if inMemory {
		path = ":memory:"
		name = "In-Memory SQLite DB"
		goto ret
	}

	appUtils.MkdirAll(path)
	path = filepath.Join(path, "default.db")

ret:
	return &SQliteDB{
		name: name,
		path: path,
	}
}

func (db *SQliteDB) GetDB() *sqlx.DB {
	return db.db
}

func (sdb *SQliteDB) Open() error {
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

func (sdb *SQliteDB) Close() error {
	return sdb.db.Close()
}

func (sdb *SQliteDB) Name() string {
	return sdb.name
}

// Returns the total number of stored tasks
func (sdb *SQliteDB) Count() (int, error) {
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

// Returns the total count of tasks in a specific completion state.
//
// Also returns an error if the specified completion state is invalid
func (sdb *SQliteDB) CountFromState(completionState int) (int, error) {
	return 0, nil
}

// Adds a task to the database
//
// Returns the Task ID and an eventual error
func (sdb *SQliteDB) Insert(nv *task.Task) (string, error) {
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
		return nv.Id, err
	}
	defer s.Close()

	_, err = s.Exec(
		nv.Id,
		nv.Aggregator,
		nv.Slug,
		nv.AggregatorPageURL,
		nv.FilehostUrl,
		nv.DisplayName,
		nv.Filename,
		nv.DownloadState,
		nv.Err,
	)
	return nv.Id, err
}

// Checks whether a task with an equal value is already present in the database
func (sdb *SQliteDB) Find(slug string) (bool, int, error) {
	count := -1

	if err := sdb.db.Get(
		&count,
		`SELECT COUNT(ID) * FROM `+TABLE_NAME+` WHERE ID = ? OR Slug LIKE ?`,
		slug, "%"+slug+"%",
	); err != nil {
		return false, count, err
	}

	return true, count, nil
}

// Checks whether a task with an equal value is already present in the database
func (sdb *SQliteDB) Get(slug string) (*task.Task, error) {
	var dest *task.Task

	if err := sdb.db.Get(
		dest,
		`SELECT * FROM `+TABLE_NAME+` WHERE ID = ? OR Slug LIKE ? LIMIT 1`,
		slug, "%"+slug+"%",
	); err != nil {
		return dest, err
	}

	return dest, nil
}

// Returns all the tasks in the database
func (sdb *SQliteDB) GetAll() ([]*task.Task, error) {
	dest := make([]*task.Task, 0)

	rows, err := sdb.db.Query(`SELECT * FROM ` + TABLE_NAME)
	if err != nil {
		return dest, err
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
			return dest, err
		}

		t.Err = fmt.Errorf("%s", t.DBErr)

		dest = append(dest, t)
	}

	return dest, nil
}

// Removes a task from the database
//
// Returns an error if trying to remove a task in a running state
func (sdb *SQliteDB) Remove(nv *task.Task) error {
	_, err := sdb.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE id = ?`, nv.Id)

	return err
}

// Removes multiple tasks with the same state from the database
//
// Returns the number of affected tasks and. If -1, then the state is out of range
//
// Also returns an error if something goes wrong while handling the database
func (sdb *SQliteDB) RemoveFromState(completionState int) (int, error) {
	if completionState < 0 || completionState >= states.MaxCompletionState() {
		return 0, fmt.Errorf("CompletionState is not a value within constraints")
	}

	res, err := sdb.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE DownloadState = ?`, completionState)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()

	return int(count), err
}

// Empties the database
func (sdb *SQliteDB) RemoveAll() error {
	_, err := sdb.db.Exec(`DELETE FROM ` + TABLE_NAME)

	return err
}

// Resets the state of EVERY task in the specified completion state
//
// Returns the affected records count and an error either if the completion state is invalid or if trying to reset tunning tasks
func (sdb *SQliteDB) ResetFromCompletionState(completionState int) (int, error) {
	if completionState < 0 || completionState > states.MaxCompletionState() {
		return 0, fmt.Errorf("CompletionState is not a value within constraints")
	}

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

// Returns the state of a specific task. Returns an error if the task has not been found
func (sdb *SQliteDB) getStateInt(t *task.Task) (int, error) {
	var stateID int

	rows, err := sdb.db.Query(`SELECT DownloadState FROM `+TABLE_NAME+` WHERE ID = ?`, t.Id)
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

// Returns the stringified state of a specific task. Returns an error if the task has not been found
func (sdb *SQliteDB) GetState(t *task.Task) (string, error) {
	stateID, err := sdb.getStateInt(t)
	if err != nil {
		return "", err
	}

	state := states.GetStateStr(stateID)

	return state, nil
}

// Sets the state of a specific task. Returns an error if the task has not been found
func (sdb *SQliteDB) SetState(t *task.Task, newState int) error {
	if newState < 0 || newState > states.MaxCompletionState() {
		return fmt.Errorf("newState is out of bounds")
	}

	_, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		newState,
		t.ID(),
	)

	return err
}

// Advances the completion state of a specific task
//
// Returns an error if the task has reached a completion state and the updated state value
func (sdb *SQliteDB) AdvanceState(t *task.Task) (int, error) {
	state, err := sdb.getStateInt(t)

	if state >= states.MaxCompletionState() {
		return state, fmt.Errorf("Cannot advance the status of this task anymore")
	}

	state = state + 1

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		state,
		t.Id,
	)

	return state, err
}

// Regresses the completion state of a specific task
//
// Returns an error if the task has reached a queued state and the updated state value
func (sdb *SQliteDB) RegressState(t *task.Task) (int, error) {
	state, err := sdb.getStateInt(t)

	if state <= 0 {
		return state, fmt.Errorf("Cannot regress the status of this task anymore")
	}

	state = state - 1

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		state,
		t.Id,
	)

	return state, err
}

// Resets the state of a specific task to a queued state
//
// Returns an error if trying to reset the state of a task in a running state and the updated state value
func (sdb *SQliteDB) ResetState(t *task.Task) (int, error) {
	state, err := sdb.getStateInt(t)
	if err != nil {
		return 0, err
	}

	if state == states.TASK_STATE_RUNNING {
		return -1, fmt.Errorf("Cannot reset a running task")
	}

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ? WHERE ID = ?`,
		states.TASK_STATE_QUEUED,
		t.Id,
	)

	return states.TASK_STATE_QUEUED, err
}

// Drops specified table name
func (sdb *SQliteDB) Drop(table string) error {
	_, err := sdb.db.Exec(`DROP TABLE ` + TABLE_NAME)

	return err
}
