package statement

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ImportJobRepository handles database operations for import jobs
type ImportJobRepository struct {
	db *sql.DB
}

func NewImportJobRepository(db *sql.DB) *ImportJobRepository {
	return &ImportJobRepository{db: db}
}

func (ijr *ImportJobRepository) Create(job *ImportJob) error {
	query := `
		INSERT INTO import_jobs (
			job_id, statement_id, user_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	job.JobID = uuid.New()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	_, err := ijr.db.Exec(query,
		job.JobID, job.StatementID, job.UserID, job.Status, job.CreatedAt, job.UpdatedAt,
	)

	return err
}

func (ijr *ImportJobRepository) GetByID(jobID uuid.UUID) (*ImportJob, error) {
	query := `
		SELECT job_id, statement_id, user_id, status, error_message, started_at, completed_at, created_at, updated_at
		FROM import_jobs
		WHERE job_id = $1
	`

	job := &ImportJob{}
	err := ijr.db.QueryRow(query, jobID).Scan(
		&job.JobID, &job.StatementID, &job.UserID, &job.Status,
		&job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (ijr *ImportJobRepository) UpdateStatus(jobID uuid.UUID, status string) error {
	query := `
		UPDATE import_jobs
		SET status = $1, updated_at = $2
		WHERE job_id = $3
	`

	_, err := ijr.db.Exec(query, status, time.Now(), jobID)
	return err
}

func (ijr *ImportJobRepository) UpdateError(jobID uuid.UUID, errMsg string) error {
	query := `
		UPDATE import_jobs
		SET error_message = $1, status = 'FAILED', updated_at = $2
		WHERE job_id = $3
	`

	_, err := ijr.db.Exec(query, errMsg, time.Now(), jobID)
	return err
}
