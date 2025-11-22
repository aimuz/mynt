package store

import (
	"encoding/json"

	"go.aimuz.me/mynt/task"
)

// TaskRepo persists task operations to the database.
type TaskRepo struct {
	db *DB
}

// NewTaskRepo creates a new task repository.
func NewTaskRepo(db *DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// Save creates a new task record.
func (r *TaskRepo) Save(op *task.Operation) error {
	query := `
	INSERT INTO tasks (id, name, state, progress, metadata, result, error, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	metaJSON, _ := json.Marshal(op.Metadata)
	resultJSON, _ := json.Marshal(op.Result)

	_, err := r.db.conn.Exec(query,
		op.ID, op.Name, op.State, op.Progress,
		string(metaJSON), string(resultJSON), op.Error,
		op.CreatedAt, op.UpdatedAt,
	)
	return err
}

// Update modifies an existing task record.
func (r *TaskRepo) Update(op *task.Operation) error {
	query := `
	UPDATE tasks SET state = ?, progress = ?, result = ?, error = ?, updated_at = ?
	WHERE id = ?
	`
	resultJSON, _ := json.Marshal(op.Result)

	_, err := r.db.conn.Exec(query,
		op.State, op.Progress, string(resultJSON), op.Error, op.UpdatedAt,
		op.ID,
	)
	return err
}

// List retrieves tasks with pagination.
func (r *TaskRepo) List(limit, offset int) ([]*task.Operation, error) {
	query := `
	SELECT id, name, state, progress, metadata, result, error, created_at, updated_at
	FROM tasks
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`
	rows, err := r.db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []*task.Operation
	for rows.Next() {
		var op task.Operation
		var metaJSON, resultJSON string
		if err := rows.Scan(
			&op.ID, &op.Name, &op.State, &op.Progress,
			&metaJSON, &resultJSON, &op.Error,
			&op.CreatedAt, &op.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metaJSON), &op.Metadata)
		json.Unmarshal([]byte(resultJSON), &op.Result)
		ops = append(ops, &op)
	}
	return ops, nil
}

// Get retrieves a single task by ID.
func (r *TaskRepo) Get(id string) (*task.Operation, error) {
	query := `
	SELECT id, name, state, progress, metadata, result, error, created_at, updated_at
	FROM tasks
	WHERE id = ?
	`
	row := r.db.conn.QueryRow(query, id)

	var op task.Operation
	var metaJSON, resultJSON string
	if err := row.Scan(
		&op.ID, &op.Name, &op.State, &op.Progress,
		&metaJSON, &resultJSON, &op.Error,
		&op.CreatedAt, &op.UpdatedAt,
	); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(metaJSON), &op.Metadata)
	json.Unmarshal([]byte(resultJSON), &op.Result)
	return &op, nil
}
