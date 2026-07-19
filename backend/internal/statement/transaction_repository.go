package statement

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (tr *TransactionRepository) Create(txn *Transaction) error {
	query := `
		INSERT INTO transactions (
			transaction_id, statement_id, user_id, bank_code, account_number_hash,
			transaction_date, amount, merchant, description, type,
			balance, raw_data, imported_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	if txn.TransactionID == "" {
		txn.TransactionID = uuid.New().String()
	}
	if txn.CreatedAt.IsZero() {
		txn.CreatedAt = time.Now()
	}
	txn.UpdatedAt = time.Now()

	// Convert RawData to JSON
	var rawDataJSON []byte
	if txn.RawData != nil {
		var err error
		rawDataJSON, err = json.Marshal(txn.RawData)
		if err != nil {
			return err
		}
	}

	_, err := tr.db.Exec(query,
		txn.TransactionID, txn.StatementID, txn.UserID, txn.BankCode,
		txn.AccountNumberHash, txn.TransactionDate, txn.Amount, txn.Merchant,
		txn.Description, txn.Type, txn.Balance, rawDataJSON,
		txn.ImportedAt, txn.CreatedAt, txn.UpdatedAt,
	)

	return err
}

func (tr *TransactionRepository) CreateBatch(txns []*Transaction) error {
	if len(txns) == 0 {
		return nil
	}

	// Use a transaction for batch insert
	tx, err := tr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO transactions (
			transaction_id, statement_id, user_id, bank_code, account_number_hash,
			transaction_date, amount, merchant, description, type,
			balance, raw_data, imported_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, txn := range txns {
		if txn.TransactionID == "" {
			txn.TransactionID = uuid.New().String()
		}
		if txn.CreatedAt.IsZero() {
			txn.CreatedAt = time.Now()
		}
		txn.UpdatedAt = time.Now()

		// Convert RawData to JSON
		var rawDataJSON []byte
		if txn.RawData != nil {
			var err error
			rawDataJSON, err = json.Marshal(txn.RawData)
			if err != nil {
				return err
			}
		}

		_, err := stmt.Exec(
			txn.TransactionID, txn.StatementID, txn.UserID, txn.BankCode,
			txn.AccountNumberHash, txn.TransactionDate, txn.Amount, txn.Merchant,
			txn.Description, txn.Type, txn.Balance, rawDataJSON,
			txn.ImportedAt, txn.CreatedAt, txn.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (tr *TransactionRepository) GetByStatement(statementID uuid.UUID) ([]*Transaction, error) {
	query := `
		SELECT transaction_id, statement_id, user_id, bank_code, account_number_hash,
		       transaction_date, amount, merchant, description, type,
		       balance, raw_data, imported_at, created_at, updated_at
		FROM transactions
		WHERE statement_id = $1
		ORDER BY transaction_date ASC
	`

	rows, err := tr.db.Query(query, statementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		txn := &Transaction{}
		var rawDataJSON []byte
		err := rows.Scan(
			&txn.TransactionID, &txn.StatementID, &txn.UserID, &txn.BankCode,
			&txn.AccountNumberHash, &txn.TransactionDate, &txn.Amount, &txn.Merchant,
			&txn.Description, &txn.Type, &txn.Balance, &rawDataJSON,
			&txn.ImportedAt, &txn.CreatedAt, &txn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON to map
		if len(rawDataJSON) > 0 {
			err = json.Unmarshal(rawDataJSON, &txn.RawData)
			if err != nil {
				return nil, err
			}
		}

		transactions = append(transactions, txn)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (tr *TransactionRepository) GetByUser(userID uuid.UUID, limit, offset int) ([]*Transaction, error) {
	query := `
		SELECT transaction_id, statement_id, user_id, bank_code, account_number_hash,
		       transaction_date, amount, merchant, description, type,
		       balance, raw_data, imported_at, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY transaction_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := tr.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		txn := &Transaction{}
		var rawDataJSON []byte
		err := rows.Scan(
			&txn.TransactionID, &txn.StatementID, &txn.UserID, &txn.BankCode,
			&txn.AccountNumberHash, &txn.TransactionDate, &txn.Amount, &txn.Merchant,
			&txn.Description, &txn.Type, &txn.Balance, &rawDataJSON,
			&txn.ImportedAt, &txn.CreatedAt, &txn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON to map
		if len(rawDataJSON) > 0 {
			err = json.Unmarshal(rawDataJSON, &txn.RawData)
			if err != nil {
				return nil, err
			}
		}

		transactions = append(transactions, txn)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (tr *TransactionRepository) GetByUserAndBank(userID uuid.UUID, bankCode string) ([]*Transaction, error) {
	query := `
		SELECT transaction_id, statement_id, user_id, bank_code, account_number_hash,
		       transaction_date, amount, merchant, description, type,
		       balance, raw_data, imported_at, created_at, updated_at
		FROM transactions
		WHERE user_id = $1 AND bank_code = $2
		ORDER BY transaction_date DESC
	`

	rows, err := tr.db.Query(query, userID, bankCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		txn := &Transaction{}
		var rawDataJSON []byte
		err := rows.Scan(
			&txn.TransactionID, &txn.StatementID, &txn.UserID, &txn.BankCode,
			&txn.AccountNumberHash, &txn.TransactionDate, &txn.Amount, &txn.Merchant,
			&txn.Description, &txn.Type, &txn.Balance, &rawDataJSON,
			&txn.ImportedAt, &txn.CreatedAt, &txn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON to map
		if len(rawDataJSON) > 0 {
			err = json.Unmarshal(rawDataJSON, &txn.RawData)
			if err != nil {
				return nil, err
			}
		}

		transactions = append(transactions, txn)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (tr *TransactionRepository) DeleteByStatement(statementID uuid.UUID) error {
	// First delete transaction_categories (foreign key constraint)
	delCatQuery := `DELETE FROM transaction_categories
	                 WHERE transaction_id IN (SELECT transaction_id FROM transactions WHERE statement_id = $1)`
	if _, err := tr.db.Exec(delCatQuery, statementID); err != nil {
		return fmt.Errorf("failed to delete transaction categories: %w", err)
	}

	// Then delete transactions
	query := `DELETE FROM transactions WHERE statement_id = $1`
	_, err := tr.db.Exec(query, statementID)
	return err
}

func (tr *TransactionRepository) CountByStatement(statementID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE statement_id = $1`
	var count int
	err := tr.db.QueryRow(query, statementID).Scan(&count)
	return count, err
}
