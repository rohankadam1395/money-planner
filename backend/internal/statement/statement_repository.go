package statement

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// StatementRepository handles database operations for statements
type StatementRepository struct {
	db *sql.DB
}

func NewStatementRepository(db *sql.DB) *StatementRepository {
	return &StatementRepository{db: db}
}

func (sr *StatementRepository) Create(stmt *Statement) error {
	query := `
		INSERT INTO statements (
			statement_id, user_id, file_name, file_format, file_size_bytes,
			file_hash, bank_code, account_number_hash, statement_period_start,
			statement_period_end, transaction_count, status, uploaded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	stmt.CreatedAt = time.Now()
	stmt.UpdatedAt = time.Now()

	err := sr.db.QueryRow(query,
		stmt.StatementID, stmt.UserID, stmt.FileName, stmt.FileFormat, stmt.FileSizeBytes,
		stmt.FileHash, stmt.BankCode, stmt.AccountNumberHash, stmt.StatementPeriodStart,
		stmt.StatementPeriodEnd, stmt.TransactionCount, stmt.Status, stmt.UploadedAt,
		stmt.CreatedAt, stmt.UpdatedAt,
	).Scan()

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}

func (sr *StatementRepository) GetByID(statementID uuid.UUID) (*Statement, error) {
	query := `
		SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
		       file_hash, bank_code, account_number_hash, statement_period_start,
		       statement_period_end, transaction_count, status, error_log, uploaded_at,
		       imported_at, created_at, updated_at
		FROM statements
		WHERE statement_id = $1
	`

	stmt := &Statement{}
	err := sr.db.QueryRow(query, statementID).Scan(
		&stmt.StatementID, &stmt.UserID, &stmt.FileName, &stmt.FileFormat,
		&stmt.FileSizeBytes, &stmt.FileHash, &stmt.BankCode, &stmt.AccountNumberHash,
		&stmt.StatementPeriodStart, &stmt.StatementPeriodEnd, &stmt.TransactionCount,
		&stmt.Status, &stmt.ErrorLog, &stmt.UploadedAt, &stmt.ImportedAt,
		&stmt.CreatedAt, &stmt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (sr *StatementRepository) GetByFileHash(fileHash string) (*Statement, error) {
	query := `
		SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
		       file_hash, bank_code, account_number_hash, statement_period_start,
		       statement_period_end, transaction_count, status, error_log, uploaded_at,
		       imported_at, created_at, updated_at
		FROM statements
		WHERE file_hash = $1
		LIMIT 1
	`

	stmt := &Statement{}
	err := sr.db.QueryRow(query, fileHash).Scan(
		&stmt.StatementID, &stmt.UserID, &stmt.FileName, &stmt.FileFormat,
		&stmt.FileSizeBytes, &stmt.FileHash, &stmt.BankCode, &stmt.AccountNumberHash,
		&stmt.StatementPeriodStart, &stmt.StatementPeriodEnd, &stmt.TransactionCount,
		&stmt.Status, &stmt.ErrorLog, &stmt.UploadedAt, &stmt.ImportedAt,
		&stmt.CreatedAt, &stmt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (sr *StatementRepository) GetByUser(userID uuid.UUID, limit, offset int) ([]*Statement, error) {
	query := `
		SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
		       file_hash, bank_code, account_number_hash, statement_period_start,
		       statement_period_end, transaction_count, status, error_log, uploaded_at,
		       imported_at, created_at, updated_at
		FROM statements
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := sr.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statements []*Statement
	for rows.Next() {
		stmt := &Statement{}
		err := rows.Scan(
			&stmt.StatementID, &stmt.UserID, &stmt.FileName, &stmt.FileFormat,
			&stmt.FileSizeBytes, &stmt.FileHash, &stmt.BankCode, &stmt.AccountNumberHash,
			&stmt.StatementPeriodStart, &stmt.StatementPeriodEnd, &stmt.TransactionCount,
			&stmt.Status, &stmt.ErrorLog, &stmt.UploadedAt, &stmt.ImportedAt,
			&stmt.CreatedAt, &stmt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return statements, nil
}

func (sr *StatementRepository) GetOverlapping(userID uuid.UUID, bankCode, accountHash string, periodStart, periodEnd time.Time) ([]*Statement, error) {
	query := `
		SELECT statement_id, user_id, file_name, file_format, file_size_bytes,
		       file_hash, bank_code, account_number_hash, statement_period_start,
		       statement_period_end, transaction_count, status, error_log, uploaded_at,
		       imported_at, created_at, updated_at
		FROM statements
		WHERE user_id = $1
		  AND bank_code = $2
		  AND account_number_hash = $3
		  AND NOT (statement_period_end < $4 OR statement_period_start > $5)
		ORDER BY statement_period_start DESC
	`

	rows, err := sr.db.Query(query, userID, bankCode, accountHash, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statements []*Statement
	for rows.Next() {
		stmt := &Statement{}
		err := rows.Scan(
			&stmt.StatementID, &stmt.UserID, &stmt.FileName, &stmt.FileFormat,
			&stmt.FileSizeBytes, &stmt.FileHash, &stmt.BankCode, &stmt.AccountNumberHash,
			&stmt.StatementPeriodStart, &stmt.StatementPeriodEnd, &stmt.TransactionCount,
			&stmt.Status, &stmt.ErrorLog, &stmt.UploadedAt, &stmt.ImportedAt,
			&stmt.CreatedAt, &stmt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return statements, nil
}

func (sr *StatementRepository) UpdateStatus(statementID uuid.UUID, status string) error {
	query := `
		UPDATE statements
		SET status = $1, updated_at = $2
		WHERE statement_id = $3
	`

	_, err := sr.db.Exec(query, status, time.Now(), statementID)
	return err
}

func (sr *StatementRepository) UpdateError(statementID uuid.UUID, errMsg string) error {
	query := `
		UPDATE statements
		SET error_log = $1, updated_at = $2
		WHERE statement_id = $3
	`

	_, err := sr.db.Exec(query, errMsg, time.Now(), statementID)
	return err
}

func (sr *StatementRepository) UpdateTransactionCount(statementID uuid.UUID, count int) error {
	query := `
		UPDATE statements
		SET transaction_count = $1, updated_at = $2
		WHERE statement_id = $3
	`

	_, err := sr.db.Exec(query, count, time.Now(), statementID)
	return err
}
