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

	ERR_STATE_OUTSIDE_CONSTRAINTS = "CompletionState is not a value within constraints"
)

type SQLiteDB struct {
	name string

	path string

	db *sqlx.DB
}

func NewSQLite(inMemory bool) *SQLiteDB {
	path := ":memory:"
	name := "In-Memory SQLite DB"

	if inMemory {
		goto ret
	}

	name = "File-Based SQLite DB"

	path = filepath.Join(".", "Database")
	appUtils.MkdirAll(path)

	path = filepath.Join(path, "default.db")

ret:
	return &SQLiteDB{
		name: name,
		path: path,
	}
}

func (db *SQLiteDB) GetDB() *sqlx.DB {
	return db.db
}

func (sdb *SQLiteDB) Open() error {
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

func (sdb *SQLiteDB) Close() error {
	return sdb.db.Close()
}

func (sdb *SQLiteDB) Name() string {
	return sdb.name
}

// Returns the total number of stored tasks
func (sdb *SQLiteDB) Count() (int, error) {
	var count int

	err := sdb.db.Get(&count, "SELECT COUNT(*) FROM "+TABLE_NAME)
	if err != nil {
		return 0, fmt.Errorf("SQLite: count query failed: %v", err)
	}

	return count, nil
}

// Returns the total count of tasks in a specific completion state.
//
// Also returns an error if the specified completion state is invalid
func (sdb *SQLiteDB) CountFromState(completionState int) (int, error) {
	var count int = 0

	if completionState < 0 || completionState >= states.MaxCompletionState() {
		return count, fmt.Errorf(ERR_STATE_OUTSIDE_CONSTRAINTS)
	}

	err := sdb.db.Get(
		&count,
		fmt.Sprintf(
			"SELECT COUNT(*) FROM %s WHERE DownloadState = %d",
			TABLE_NAME,
			completionState,
		),
	)
	if err != nil {
		return 0, fmt.Errorf("SQLite: count query failed: %v", err)
	}

	return count, nil
}

// Returns the total count of tasks in a specific completion state. The count becomes -1 if an error happens
func (sdb *SQLiteDB) CountFromStateNoErr(state int) int {
	var count int = 0

	if state < 0 || state >= states.MaxCompletionState() {
		return -1
	}

	err := sdb.db.Get(
		&count,
		fmt.Sprintf(
			"SELECT COUNT(*) FROM %s WHERE DownloadState = %d",
			TABLE_NAME,
			state,
		),
	)
	if err != nil {
		return -1
	}

	return count
}

// Adds a task to the database
//
// Returns the Task ID and an eventual error
func (sdb *SQLiteDB) Insert(nv *task.Task) (string, error) {
	if found, _, _ := sdb.Find(nv.Slug); found {
		return nv.Id, fmt.Errorf("DB: Insert: Task already present in the database")
	}

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

	dbErr := ""
	if nv.Err != nil {
		dbErr = nv.Err.Error()
	}

	_, err = s.Exec(
		nv.Id,
		nv.Aggregator,
		nv.Slug,
		nv.AggregatorPageURL,
		nv.FilehostUrl,
		nv.DisplayName,
		nv.Filename,
		nv.DownloadState,
		dbErr,
	)

	return nv.Id, err
}

// Checks whether a task with an equal value is already present in the database
//
// Returns:
//
//   - Wether at least an entry has been found
//
//   - The number of found entries
//
//   - An eventual error
func (sdb *SQLiteDB) Find(slugOrID string) (bool, int, error) {
	count := -1

	if err := sdb.db.Get(
		&count,
		`SELECT COUNT(ID) * FROM `+TABLE_NAME+` WHERE ID = ? OR Slug LIKE ?`,
		slugOrID, "%"+slugOrID+"%",
	); err != nil {
		return false, count, err
	}

	return true, count, nil
}

// Checks whether a task with an equal value is already present in the database
func (sdb *SQLiteDB) Get(slug string) (*task.Task, error) {
	var dest *task.Task = new(task.Task)

	row := sdb.db.QueryRowx(
		`SELECT 
			ID, 
            COALESCE(Aggregator, '') AS Aggregator, 
            COALESCE(Slug, '') AS Slug, 
            COALESCE(AggregatorPageURL, '') AS AggregatorPageURL, 
            COALESCE(FilehostUrl, '') AS FilehostUrl, 
            COALESCE(DisplayName, '') AS DisplayName, 
            COALESCE(Filename, '') AS Filename, 
            DownloadState, 
            COALESCE(Err, '') AS Err
		FROM `+TABLE_NAME+`
		WHERE ID = ? OR Slug LIKE ?
		LIMIT 1`,
		slug, "%"+slug+"%",
	)

	if row.Err() != nil {
		return nil, fmt.Errorf("SQLite: query error: %v", row.Err())
	}

	var dbErr string

	err := row.Scan(
		&dest.Id,
		&dest.Aggregator,
		&dest.Slug,
		&dest.AggregatorPageURL,
		&dest.FilehostUrl,
		&dest.DisplayName,
		&dest.Filename,
		&dest.DownloadState,
		&dbErr,
	)
	if err != nil {
		return dest, err
	}

	if dbErr == "" {
		dest.Err = nil
	} else {
		dest.Err = fmt.Errorf("%s", dbErr)
	}

	return dest, nil
}

// Checks whether a task with an equal value is already present in the database
func (sdb *SQLiteDB) GetFromState(state int) (*task.Task, error) {
	var dest *task.Task = new(task.Task)

	row := sdb.db.QueryRowx(
		`SELECT 
			ID, 
            COALESCE(Aggregator, '') AS Aggregator, 
            COALESCE(Slug, '') AS Slug, 
            COALESCE(AggregatorPageURL, '') AS AggregatorPageURL, 
            COALESCE(FilehostUrl, '') AS FilehostUrl, 
            COALESCE(DisplayName, '') AS DisplayName, 
            COALESCE(Filename, '') AS Filename, 
            DownloadState, 
            COALESCE(Err, '') AS Err
		FROM `+TABLE_NAME+`
		WHERE DownloadState = ?
		LIMIT 1`,
		state,
	)

	if row.Err() != nil {
		return nil, fmt.Errorf("SQLite: query error: %v", row.Err())
	}

	var dbErr string

	err := row.Scan(
		&dest.Id,
		&dest.Aggregator,
		&dest.Slug,
		&dest.AggregatorPageURL,
		&dest.FilehostUrl,
		&dest.DisplayName,
		&dest.Filename,
		&dest.DownloadState,
		&dbErr,
	)
	if err != nil {
		return dest, err
	}

	if dbErr == "" {
		dest.Err = nil
	} else {
		dest.Err = fmt.Errorf("%s", dbErr)
	}

	return dest, nil
}

// Returns all the tasks in the database
func (sdb *SQLiteDB) GetAll() ([]*task.Task, error) {
	dest := make([]*task.Task, 0)

	rows, err := sdb.db.Query(`
        SELECT 
            ID, 
            COALESCE(Aggregator, ''), 
            COALESCE(Slug, ''), 
            COALESCE(AggregatorPageURL, ''), 
            COALESCE(FilehostUrl, ''), 
            COALESCE(DisplayName, ''), 
            COALESCE(Filename, ''), 
            DownloadState, 
            COALESCE(Err, '')
        FROM ` + TABLE_NAME,
	)
	if err != nil {
		return dest, err
	}
	defer rows.Close()

	for rows.Next() {
		t := task.NewTask("")
		var dbErr string

		err := rows.Scan(
			&t.Id,
			&t.Aggregator,
			&t.Slug,
			&t.AggregatorPageURL,
			&t.FilehostUrl,
			&t.DisplayName,
			&t.Filename,
			&t.DownloadState,
			&dbErr,
		)
		if err != nil {
			return dest, err
		}

		if dbErr == "" {
			t.Err = nil
		} else {
			t.Err = fmt.Errorf("%s", dbErr)
		}

		dest = append(dest, t)
	}

	return dest, nil
}

// Returns all the tasks in the database with that state
func (sdb *SQLiteDB) GetAllWithState(state int) ([]*task.Task, error) {
	dest := make([]*task.Task, 0)

	if state < 0 || state >= states.MaxCompletionState() {
		return dest, fmt.Errorf(ERR_STATE_OUTSIDE_CONSTRAINTS)
	}

	rows, err := sdb.db.Queryx(`
        SELECT 
            ID, 
            COALESCE(Aggregator, ''), 
            COALESCE(Slug, ''), 
            COALESCE(AggregatorPageURL, ''), 
            COALESCE(FilehostUrl, ''), 
            COALESCE(DisplayName, ''), 
            COALESCE(Filename, ''), 
            DownloadState, 
            COALESCE(Err, '')
        FROM `+TABLE_NAME+`
		WHERE DownloadState = ?`,
		state,
	)
	if err != nil {
		return dest, err
	}
	defer rows.Close()

	for rows.Next() {
		t := task.NewTask("")
		var dbErr string

		err := rows.Scan(
			&t.Id,
			&t.Aggregator,
			&t.Slug,
			&t.AggregatorPageURL,
			&t.FilehostUrl,
			&t.DisplayName,
			&t.Filename,
			&t.DownloadState,
			&dbErr,
		)
		if err != nil {
			return dest, err
		}

		if dbErr == "" {
			t.Err = nil
		} else {
			t.Err = fmt.Errorf("%s", dbErr)
		}

		dest = append(dest, t)
	}

	return dest, nil
}

func (sdb *SQLiteDB) Update(t *task.Task) error {
	s, err := sdb.db.Prepare(`
		UPDATE ` + TABLE_NAME + `
		SET
			Slug = ?,
			AggregatorPageURL = ?,
			FilehostUrl = ?,
			DisplayName = ?,
			Filename = ?,
			DownloadState = ?,
			Err = ?
		WHERE
			ID = ?
	`)
	if err != nil {
		return err
	}
	defer s.Close()

	dbErr := ""
	if t.Err != nil {
		dbErr = t.Err.Error()
	}

	_, err = s.Exec(
		t.Slug,
		t.AggregatorPageURL,
		t.FilehostUrl,
		t.DisplayName,
		t.Filename,
		t.DownloadState,
		dbErr,
		t.Id,
	)

	return err
}

// Removes a task from the database
//
// Returns an error if trying to remove a task in a running state
func (sdb *SQLiteDB) Remove(t *task.Task) error {
	_, err := sdb.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE id = ?`, t.Id)

	return err
}

// Removes a task from the database
//
// Returns an error if trying to remove a task in a running state
func (sdb *SQLiteDB) RemoveFromString(slugOrID string) error {
	_, err := sdb.db.Exec(
		`SELECT COUNT(ID) * FROM `+TABLE_NAME+` WHERE ID = ? OR Slug LIKE ?`,
		slugOrID,
		"%"+slugOrID+"%",
	)

	return err
}

// Removes multiple tasks with the same state from the database
//
// Returns the number of affected tasks and. If -1, then the state is out of range
//
// Also returns an error if something goes wrong while handling the database
func (sdb *SQLiteDB) RemoveFromState(state int) (int, error) {
	if state < 0 || state >= states.MaxCompletionState() {
		return 0, fmt.Errorf(ERR_STATE_OUTSIDE_CONSTRAINTS)
	}

	res, err := sdb.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE DownloadState = ?`, state)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()

	return int(count), err
}

func (sdb *SQLiteDB) RemoveCompletedNoErr() (int, error) {
	res, err := sdb.db.Exec(
		`DELETE FROM `+TABLE_NAME+` WHERE DownloadState = ? AND (Err = '' OR Err IS NULL)`,
		states.TASK_STATE_COMPLETED,
	)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()

	return int(count), err
}

func (sdb *SQLiteDB) RemoveCompletedWithErr() (int, error) {
	res, err := sdb.db.Exec(
		`DELETE FROM `+TABLE_NAME+` WHERE DownloadState = ? AND Err IS NOT NULL`,
		states.TASK_STATE_COMPLETED,
	)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()

	return int(count), err
}

// Empties the database
func (sdb *SQLiteDB) RemoveAll() error {
	_, err := sdb.db.Exec(`DELETE FROM ` + TABLE_NAME)

	return err
}

// Resets the state of EVERY task in the specified completion state
//
// Returns the affected records count and an error either if the completion state is invalid or if trying to reset tunning tasks
func (sdb *SQLiteDB) ResetFromCompletionState(state int) (int, error) {
	if state < 0 || state > states.MaxCompletionState() {
		return 0, fmt.Errorf(ERR_STATE_OUTSIDE_CONSTRAINTS)
	}

	if state == states.TASK_STATE_RUNNING {
		return 0, fmt.Errorf("Cannot reset running tasks")
	}

	res, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ?, Err = "" WHERE DownloadState = ?`,
		states.TASK_STATE_QUEUED,
		state,
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
func (sdb *SQLiteDB) getStateInt(t *task.Task) (int, error) {
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
func (sdb *SQLiteDB) GetState(t *task.Task) (string, error) {
	stateID, err := sdb.getStateInt(t)
	if err != nil {
		return "", err
	}

	state := states.GetStateStr(stateID)

	return state, nil
}

// Sets the state of a specific task. Returns an error if the task has not been found
func (sdb *SQLiteDB) SetState(t *task.Task, newState int) error {
	if newState < 0 || newState > states.MaxCompletionState() {
		return fmt.Errorf(ERR_STATE_OUTSIDE_CONSTRAINTS)
	}

	dbErr := ""
	if t.Err != nil {
		dbErr = t.Err.Error()
	}

	_, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ?, Err = ? WHERE ID = ?`,
		newState,
		dbErr,
		t.ID(),
	)

	return err
}

// Advances the completion state of a specific task
//
// Returns an error if the task has reached a completion state and the updated state value
func (sdb *SQLiteDB) AdvanceState(t *task.Task) (int, error) {
	state, err := sdb.getStateInt(t)

	if state >= states.MaxCompletionState() {
		return state, fmt.Errorf("Cannot advance the status of this task anymore")
	}

	state = state + 1

	dbErr := ""
	if t.Err != nil {
		dbErr = t.Err.Error()
	}

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ?, Err = ? WHERE ID = ?`,
		state,
		dbErr,
		t.Id,
	)
	if err != nil {
		return state, err
	}

	t.DownloadState = state

	return state, nil
}

// Regresses the completion state of a specific task
//
// Returns an error if the task has reached a queued state and the updated state value
func (sdb *SQLiteDB) RegressState(t *task.Task) (int, error) {
	state, err := sdb.getStateInt(t)

	if state <= 0 {
		return state, fmt.Errorf("Cannot regress the status of this task anymore")
	}

	state = state - 1

	_, err = sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ?, Err = ? WHERE ID = ?`,
		state,
		"",
		t.Id,
	)
	if err != nil {
		return 0, err
	}

	t.DownloadState = state

	return state, nil
}

// Resets the state of a specific task to a queued state
//
// Returns an error if trying to reset the state of a task in a running state and the updated state value
func (sdb *SQLiteDB) ResetState(t *task.Task) (int, error) {
	if t.DownloadState == states.TASK_STATE_RUNNING {
		return -1, fmt.Errorf("Cannot reset a running task")
	}

	_, err := sdb.db.Exec(
		`UPDATE `+TABLE_NAME+` SET DownloadState = ?, Err = ? WHERE ID = ?`,
		states.TASK_STATE_QUEUED,
		"",
		t.Id,
	)

	t.DownloadState = states.TASK_STATE_QUEUED
	t.Err = nil

	return states.TASK_STATE_QUEUED, err
}

// Drops specified table name
func (sdb *SQLiteDB) Drop(table string) error {
	_, err := sdb.db.Exec(`DROP TABLE ` + TABLE_NAME)

	return err
}
