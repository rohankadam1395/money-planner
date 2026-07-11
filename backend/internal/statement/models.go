package statement

import (
	"time"

	"github.com/google/uuid"
)

// Statement represents a bank statement file and its import metadata
type Statement struct {
	StatementID        uuid.UUID
	UserID             uuid.UUID
	FileName           string
	FileFormat         string // PDF, CSV, XLSX
	FileSizeBytes      int
	FileHash           string // SHA-256 hash
	BankCode           string
	AccountNumberHash  string
	StatementPeriodStart time.Time
	StatementPeriodEnd   time.Time
	TransactionCount   int
	Status             string // PENDING, SUCCESS, FAILED
	ErrorLog           *string
	UploadedAt         time.Time
	ImportedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ImportJob tracks async statement processing jobs
type ImportJob struct {
	JobID        uuid.UUID
	StatementID  uuid.UUID
	UserID       uuid.UUID
	Status       string // PENDING, PROCESSING, COMPLETED, FAILED
	ErrorMessage *string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
