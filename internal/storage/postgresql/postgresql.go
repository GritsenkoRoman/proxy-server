package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"proxy-server/internal/storage"
	"proxy-server/internal/task"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
}

func New(connection string) (*Postgres, error) {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id               TEXT PRIMARY KEY,
			status           TEXT NOT NULL,
			http_status_code INTEGER,
			headers          JSONB,
			length           BIGINT,
			error            TEXT,
		)
	`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Postgres{
		db: db,
	}, nil
}

func (s *Postgres) CreateEmptyResponse() (task.Response, error) {
	result := task.Response{
		ID:     uuid.NewString(),
		Status: task.StatusNew,
	}

	_, err := s.db.Exec(`
		INSERT INTO tasks (id, status)
		VALUES ($1, $2)
	`, result.ID, result.Status)
	if err != nil {
		if isUniqueViolation(err) {
			return task.Response{}, storage.ErrTaskExists
		}
		return task.Response{}, err
	}

	return result, nil
}

func (s *Postgres) SetStatus(
	id string,
	status task.Status,
	resp *http.Response,
	err error,
) {
	switch status {
	case task.StatusDone:
		if resp == nil {
			return
		}

		headers, marshalErr := json.Marshal(resp.Header.Clone())
		if marshalErr != nil {
			return
		}

		_, _ = s.db.Exec(`
			UPDATE tasks
			SET status = $2,
			    http_status_code = $3,
			    headers = $4,
			    length = $5,
			    error = NULL
			WHERE id = $1
		`,
			id,
			status,
			resp.StatusCode,
			headers,
			resp.ContentLength,
		)

	case task.StatusError:
		if err == nil {
			return
		}

		_, _ = s.db.Exec(`
			UPDATE tasks
			SET status = $2,
			    error = $3
			WHERE id = $1
		`, id, status, err.Error())

	default:
		_, _ = s.db.Exec(`
			UPDATE tasks
			SET status = $2
			WHERE id = $1
		`, id, status)
	}
}

func (s *Postgres) Get(id string) (task.Response, error) {
	var (
		result     task.Response
		status     string
		statusCode sql.NullInt64
		length     sql.NullInt64
		headers    []byte
		errText    sql.NullString
	)

	err := s.db.QueryRow(`
		SELECT id, status, http_status_code, headers, length, error
		FROM tasks
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&status,
		&statusCode,
		&headers,
		&length,
		&errText,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return task.Response{}, storage.ErrTaskNotFound
	}
	if err != nil {
		return task.Response{}, err
	}

	result.Status = task.Status(status)

	if statusCode.Valid {
		result.HTTPStatusCode = int(statusCode.Int64)
	}

	if length.Valid {
		result.Length = length.Int64
	}

	if errText.Valid {
		result.Error = errText.String
	}

	if len(headers) > 0 {
		if err := json.Unmarshal(headers, &result.Headers); err != nil {
			return task.Response{}, err
		}
	}

	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface {
		SQLState() string
	}

	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
