-- Migration: Create statements and transactions schema
-- Date: 2026-07-05
-- Purpose: Create core tables for Statement Import feature

-- Statements table: metadata about imported bank statements
CREATE TABLE IF NOT EXISTS statements (
    statement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    file_name VARCHAR(256) NOT NULL,
    file_format VARCHAR(10) NOT NULL CHECK (file_format IN ('PDF', 'CSV', 'XLSX')),
    file_size_bytes INTEGER NOT NULL CHECK (file_size_bytes > 0),
    file_hash VARCHAR(64) NOT NULL UNIQUE,
    bank_code CHAR(4) NOT NULL,
    account_number_hash VARCHAR(64) NOT NULL,
    statement_period_start DATE NOT NULL,
    statement_period_end DATE NOT NULL CHECK (statement_period_end >= statement_period_start),
    transaction_count INTEGER NOT NULL CHECK (transaction_count >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED')),
    error_log TEXT,
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    imported_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_statements_user_uploaded ON statements(user_id, uploaded_at DESC);
CREATE INDEX idx_statements_file_hash ON statements(file_hash);
CREATE INDEX idx_statements_overlap_detection ON statements(bank_code, account_number_hash, statement_period_start);

-- Transactions table: individual financial transactions extracted from statements
CREATE TABLE IF NOT EXISTS transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    statement_id UUID NOT NULL REFERENCES statements(statement_id) ON DELETE CASCADE,
    transaction_date DATE NOT NULL,
    merchant VARCHAR(256) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    type VARCHAR(10) NOT NULL CHECK (type IN ('DEBIT', 'CREDIT')),
    balance DECIMAL(12, 2),
    description VARCHAR(512),
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    imported_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bank_code CHAR(4) NOT NULL,
    account_number_hash VARCHAR(64) NOT NULL,
    raw_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX idx_transactions_statement ON transactions(statement_id);
CREATE INDEX idx_transactions_duplicate_detection ON transactions(user_id, bank_code, account_number_hash, transaction_date);

-- Import jobs table: tracks async statement processing jobs
CREATE TABLE IF NOT EXISTS import_jobs (
    job_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id UUID NOT NULL REFERENCES statements(statement_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')),
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_import_jobs_statement ON import_jobs(statement_id);
CREATE INDEX idx_import_jobs_user_status ON import_jobs(user_id, status);
