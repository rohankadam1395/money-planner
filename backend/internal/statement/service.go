package statement

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StatementService handles all statement import operations
type StatementService struct {
	stmtRepo *StatementRepository
	txnRepo  *TransactionRepository
	jobRepo  *ImportJobRepository
}

// NewStatementService creates a new statement service
func NewStatementService(
	stmtRepo *StatementRepository,
	txnRepo *TransactionRepository,
	jobRepo *ImportJobRepository,
) *StatementService {
	return &StatementService{
		stmtRepo: stmtRepo,
		txnRepo:  txnRepo,
		jobRepo:  jobRepo,
	}
}

// UploadRequest represents an upload request
type UploadRequest struct {
	FileContent []byte
	FileName    string
	FileFormat  string // PDF, CSV, XLSX
	BankCode    string
	UserID      string
}

// UploadResponse represents the response after upload
type UploadResponse struct {
	StatementID  string    `json:"statement_id"`
	Status       string    `json:"status"`
	BankCode     string    `json:"bank_code"`
	FileName     string    `json:"file_name"`
	FileFormat   string    `json:"file_format"`
	UploadedAt   time.Time `json:"uploaded_at"`
	FileHash     string    `json:"file_hash,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Upload handles statement file upload and initialization
func (s *StatementService) Upload(req *UploadRequest) (*UploadResponse, error) {
	// Validate input
	if err := s.validateUploadRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Compute file hash for duplicate detection
	fileHash := s.computeFileHash(req.FileContent)

	// Check for duplicates
	existingStmt, err := s.stmtRepo.GetByFileHash(fileHash)
	if err == nil && existingStmt != nil {
		return &UploadResponse{
			StatementID: existingStmt.StatementID.String(),
			Status:      "DUPLICATE",
			FileHash:    fileHash,
		}, fmt.Errorf("statement already imported")
	}

	// Create statement record
	stmt := &Statement{
		StatementID:   uuid.New(),
		UserID:        uuid.MustParse(req.UserID),
		FileName:      req.FileName,
		FileFormat:    req.FileFormat,
		FileSizeBytes: len(req.FileContent),
		FileHash:      fileHash,
		BankCode:      req.BankCode,
		Status:        "PENDING",
		UploadedAt:    time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create initial transaction hash (placeholder for now)
	stmt.AccountNumberHash = "0000000000000000000000000000000000000000000000000000000000000000"

	// Save statement record
	if err := s.stmtRepo.Create(stmt); err != nil {
		return nil, fmt.Errorf("failed to save statement: %w", err)
	}

	// Extract transactions immediately
	rawTxns, err := s.ExtractTransactions(stmt.StatementID.String(), req.FileContent, req.FileFormat)
	if err != nil {
		// Still create statement but update status to failed
		errorMsg := err.Error()
		stmt.Status = "FAILED"
		stmt.ErrorLog = &errorMsg
		s.stmtRepo.UpdateStatus(stmt.StatementID, stmt.Status)
		s.stmtRepo.UpdateError(stmt.StatementID, errorMsg)
		return &UploadResponse{
			StatementID: stmt.StatementID.String(),
			Status:      "FAILED",
			BankCode:    stmt.BankCode,
			FileName:    stmt.FileName,
			FileFormat:  stmt.FileFormat,
			UploadedAt:  stmt.UploadedAt,
			FileHash:    fileHash,
			ErrorMessage: err.Error(),
		}, err
	}

	// Convert and save transactions if extraction successful
	fmt.Printf("[DEBUG] rawTxns=%v, len=%d\n", rawTxns, len(rawTxns))
	if rawTxns != nil && len(rawTxns) > 0 {
		var txns []*Transaction
		fmt.Printf("[DEBUG] Converting %d raw transactions\n", len(rawTxns))
		for _, raw := range rawTxns {
			txn := &Transaction{
				TransactionID:     uuid.New().String(),
				StatementID:       stmt.StatementID.String(),
				UserID:            req.UserID,
				BankCode:          req.BankCode,
				AccountNumberHash: stmt.AccountNumberHash,
				TransactionDate:   raw.Date,
				Merchant:          raw.Merchant,
				Amount:            raw.Amount,
				Type:              raw.Type,
				Balance:           raw.Balance,
				Description:       raw.Description,
				Currency:          raw.Currency,
				RawData:           raw.RawData,
				ImportedAt:        time.Now(),
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}
			txns = append(txns, txn)
		}

		// Batch insert transactions
		if err := s.txnRepo.CreateBatch(txns); err != nil {
			// Log error but don't fail the upload
			fmt.Printf("error saving transactions: %v\n", err)
		}

		// Update statement with transaction count
		stmt.TransactionCount = len(txns)
		if err := s.stmtRepo.UpdateTransactionCount(stmt.StatementID, stmt.TransactionCount); err != nil {
			fmt.Printf("error updating transaction count: %v\n", err)
		}
	}

	// Create import job for async processing
	job := &ImportJob{
		JobID:       uuid.New(),
		StatementID: stmt.StatementID,
		UserID:      stmt.UserID,
		Status:      "COMPLETED",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.jobRepo.Create(job); err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	return &UploadResponse{
		StatementID: stmt.StatementID.String(),
		Status:      "SUCCESS",
		BankCode:    stmt.BankCode,
		FileName:    stmt.FileName,
		FileFormat:  stmt.FileFormat,
		UploadedAt:  stmt.UploadedAt,
		FileHash:    fileHash,
	}, nil
}

// ExtractTransactions extracts transactions from a statement file
func (s *StatementService) ExtractTransactions(statementID string, fileContent []byte, format string) ([]*RawTransaction, error) {
	if len(fileContent) == 0 {
		return nil, fmt.Errorf("empty file content")
	}

	// Choose parser based on file format
	var transactions []*RawTransaction
	var err error

	switch strings.ToUpper(format) {
	case "CSV":
		parser := NewCSVParser()
		transactions, err = parser.ParseCSV(strings.NewReader(string(fileContent)))
	case "PDF":
		parser := NewPDFParser(&HDFCFormat{})
		transactions, err = parser.ParsePDF(strings.NewReader(string(fileContent)))
	case "XLSX":
		parser := NewExcelParser(&ICICIFormat{})
		transactions, err = parser.ParseExcel(strings.NewReader(string(fileContent)))
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return transactions, nil
}

// PreviewResponse contains the preview data
type PreviewResponse struct {
	Transactions      []*Transaction        `json:"transactions"`
	ValidationSummary *ValidationSummary    `json:"validation_summary"`
	Categorization    interface{}           `json:"categorization,omitempty"`
	Status            string                `json:"status"`
	Message           string                `json:"message,omitempty"`
}

// ValidationSummary contains validation results
type ValidationSummary struct {
	TotalTransactions   int                   `json:"total_transactions"`
	ValidTransactions   int                   `json:"valid_transactions"`
	InvalidTransactions int                   `json:"invalid_transactions"`
	Errors              []map[string]interface{} `json:"errors"`
}

// PreviewTransactions returns extracted transactions for preview
func (s *StatementService) PreviewTransactions(statementID string, rawTxns []*RawTransaction, periodStart, periodEnd time.Time) (*PreviewResponse, error) {
	if len(rawTxns) == 0 {
		return &PreviewResponse{
			Status:        "EMPTY",
			Message:       "No transactions found in statement",
			Transactions: []*Transaction{},
			ValidationSummary: &ValidationSummary{
				TotalTransactions:   0,
				ValidTransactions:   0,
				InvalidTransactions: 0,
				Errors:              []map[string]interface{}{},
			},
		}, nil
	}

	var validTxns []*Transaction
	var validationErrors []map[string]interface{}
	validCount := 0
	invalidCount := 0

	for _, raw := range rawTxns {
		// Convert raw transaction to domain transaction
		txn := &Transaction{
			TransactionID:     uuid.New().String(),
			StatementID:       statementID,
			TransactionDate:   raw.Date,
			Merchant:          raw.Merchant,
			Amount:            raw.Amount,
			Type:              raw.Type,
			Balance:           raw.Balance,
			Description:       raw.Description,
			Currency:          "INR",
			BankCode:          "HDFC", // TODO: extract from statement
			AccountNumberHash: "0000000000000000000000000000000000000000000000000000000000000000",
			ImportedAt:        time.Now(),
		}

		// Validate transaction
		validation := ValidateTransaction(txn, periodStart, periodEnd)
		if validation.Valid {
			validCount++
			validTxns = append(validTxns, txn)
		} else {
			invalidCount++
			for _, err := range validation.Errors {
				validationErrors = append(validationErrors, map[string]interface{}{
					"field":   err.Field,
					"message": err.Message,
				})
			}
		}
	}

	return &PreviewResponse{
		Transactions: validTxns,
		ValidationSummary: &ValidationSummary{
			TotalTransactions:   len(rawTxns),
			ValidTransactions:   validCount,
			InvalidTransactions: invalidCount,
			Errors:              validationErrors,
		},
		Status: "READY",
	}, nil
}

// ConfirmImportRequest contains confirmation details
type ConfirmImportRequest struct {
	StatementID string
	UserID      string
	Transactions []*Transaction
}

// ConfirmImportResponse contains the confirmation result
type ConfirmImportResponse struct {
	StatementID      string    `json:"statement_id"`
	Status           string    `json:"status"`
	TransactionCount int       `json:"transaction_count"`
	ImportedAt       time.Time `json:"imported_at"`
	Message          string    `json:"message,omitempty"`
}

// ConfirmImport persists transactions to the database
func (s *StatementService) ConfirmImport(req *ConfirmImportRequest) (*ConfirmImportResponse, error) {
	if len(req.Transactions) == 0 {
		return nil, fmt.Errorf("no transactions to import")
	}

	// Create batch insert for transactions
	if err := s.txnRepo.CreateBatch(req.Transactions); err != nil {
		return nil, fmt.Errorf("failed to persist transactions: %w", err)
	}

	// Update statement status
	if err := s.stmtRepo.UpdateStatus(uuid.MustParse(req.StatementID), "SUCCESS"); err != nil {
		return nil, fmt.Errorf("failed to update statement status: %w", err)
	}

	return &ConfirmImportResponse{
		StatementID:      req.StatementID,
		Status:           "SUCCESS",
		TransactionCount: len(req.Transactions),
		ImportedAt:       time.Now(),
		Message:          fmt.Sprintf("Successfully imported %d transactions", len(req.Transactions)),
	}, nil
}

// CheckForDuplicates detects duplicate statements
func (s *StatementService) CheckForDuplicates(userID, bankCode, accountHash string, periodStart, periodEnd time.Time) ([]*Statement, error) {
	return s.stmtRepo.GetOverlapping(uuid.MustParse(userID), bankCode, accountHash, periodStart, periodEnd)
}

// ListStatements returns user's import history
func (s *StatementService) ListStatements(userID string, limit, offset int) ([]*Statement, error) {
	return s.stmtRepo.GetByUser(uuid.MustParse(userID), limit, offset)
}

// Helper methods

func (s *StatementService) validateUploadRequest(req *UploadRequest) error {
	if req.FileContent == nil || len(req.FileContent) == 0 {
		return fmt.Errorf("file content is empty")
	}

	fileSize := len(req.FileContent)
	if fileSize > 50*1024*1024 {
		return fmt.Errorf("file size exceeds 50MB limit")
	}

	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}

	req.FileFormat = strings.ToUpper(req.FileFormat)
	if req.FileFormat != "PDF" && req.FileFormat != "CSV" && req.FileFormat != "XLSX" {
		return fmt.Errorf("unsupported file format: %s", req.FileFormat)
	}

	if req.BankCode == "" {
		return fmt.Errorf("bank code is required")
	}

	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	return nil
}

func (s *StatementService) computeFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// GetStatement fetches a statement by ID
func (s *StatementService) GetStatement(statementID string) (*Statement, error) {
	id, err := uuid.Parse(statementID)
	if err != nil {
		return nil, fmt.Errorf("invalid statement ID format")
	}
	return s.stmtRepo.GetByID(id)
}

// GetTransactions fetches all transactions for a statement
func (s *StatementService) GetTransactions(statementID string) ([]*Transaction, error) {
	id, err := uuid.Parse(statementID)
	if err != nil {
		return nil, fmt.Errorf("invalid statement ID format")
	}
	return s.txnRepo.GetByStatement(id)
}
