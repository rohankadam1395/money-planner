import React from 'react';

interface StatementInfo {
  id: string;
  bank: string;
  periodStart: string;
  periodEnd: string;
  uploadDate: string;
  transactionCount: number;
}

interface DuplicateWarningProps {
  isOpen: boolean;
  existingStatement?: StatementInfo;
  newStatementInfo?: Partial<StatementInfo>;
  duplicateType: 'EXACT_FILE' | 'DATE_RANGE' | 'TRANSACTION_OVERLAP' | 'NONE';
  onContinue: () => void;
  onCancel: () => void;
  onShowHistory?: () => void;
}

export const DuplicateWarning: React.FC<DuplicateWarningProps> = ({
  isOpen,
  existingStatement,
  newStatementInfo,
  duplicateType,
  onContinue,
  onCancel,
  onShowHistory,
}) => {
  if (!isOpen) return null;

  const getWarningTitle = (): string => {
    switch (duplicateType) {
      case 'EXACT_FILE':
        return 'Duplicate File Detected';
      case 'DATE_RANGE':
        return 'Overlapping Statement Period';
      case 'TRANSACTION_OVERLAP':
        return 'Potential Duplicate Transactions';
      default:
        return 'Upload Confirmation';
    }
  };

  const getWarningMessage = (): string => {
    switch (duplicateType) {
      case 'EXACT_FILE':
        return 'This exact statement file has already been imported. Uploading again will not add new transactions.';
      case 'DATE_RANGE':
        return `The statement period overlaps with a previously imported statement from ${existingStatement?.bank} ` +
               `(${existingStatement?.periodStart} to ${existingStatement?.periodEnd}). ` +
               'Duplicate transactions will be detected and excluded.';
      case 'TRANSACTION_OVERLAP':
        return 'Some transactions in this statement may be duplicates of previously imported transactions. ' +
               'Duplicates will be automatically detected during import.';
      default:
        return 'Proceed with uploading this statement?';
    }
  };

  const getWarningLevel = (): 'error' | 'warning' | 'info' => {
    switch (duplicateType) {
      case 'EXACT_FILE':
        return 'error';
      case 'DATE_RANGE':
      case 'TRANSACTION_OVERLAP':
        return 'warning';
      default:
        return 'info';
    }
  };

  const shouldShowContinue = duplicateType !== 'EXACT_FILE';

  return (
    <div className="duplicate-warning-overlay">
      <div className="duplicate-warning-modal">
        <div className={`warning-header ${getWarningLevel()}`}>
          <h2>{getWarningTitle()}</h2>
          <button
            className="close-button"
            onClick={onCancel}
            aria-label="Close"
          >
            ×
          </button>
        </div>

        <div className="warning-content">
          <p className="warning-message">{getWarningMessage()}</p>

          {existingStatement && duplicateType !== 'EXACT_FILE' && (
            <div className="existing-statement-info">
              <h4>Existing Statement Details</h4>
              <div className="info-grid">
                <div className="info-item">
                  <span className="label">Bank</span>
                  <span className="value">{existingStatement.bank}</span>
                </div>
                <div className="info-item">
                  <span className="label">Period</span>
                  <span className="value">
                    {existingStatement.periodStart} to {existingStatement.periodEnd}
                  </span>
                </div>
                <div className="info-item">
                  <span className="label">Uploaded</span>
                  <span className="value">{existingStatement.uploadDate}</span>
                </div>
                <div className="info-item">
                  <span className="label">Transactions</span>
                  <span className="value">{existingStatement.transactionCount}</span>
                </div>
              </div>
            </div>
          )}

          {newStatementInfo && (
            <div className="new-statement-info">
              <h4>New Statement Details</h4>
              <div className="info-grid">
                <div className="info-item">
                  <span className="label">Bank</span>
                  <span className="value">{newStatementInfo.bank}</span>
                </div>
                <div className="info-item">
                  <span className="label">Period</span>
                  <span className="value">
                    {newStatementInfo.periodStart} to {newStatementInfo.periodEnd}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="warning-actions">
          <button onClick={onCancel} className="btn-secondary">
            Cancel Upload
          </button>

          {onShowHistory && (
            <button onClick={onShowHistory} className="btn-tertiary">
              View History
            </button>
          )}

          {shouldShowContinue && (
            <button onClick={onContinue} className="btn-primary">
              Continue Upload
            </button>
          )}
        </div>
      </div>

      <style jsx>{`
        .duplicate-warning-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background-color: rgba(0, 0, 0, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
        }

        .duplicate-warning-modal {
          background: white;
          border-radius: 8px;
          box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
          max-width: 500px;
          width: 90%;
          max-height: 80vh;
          overflow-y: auto;
        }

        .warning-header {
          padding: 16px 20px;
          border-bottom: 1px solid #eee;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .warning-header h2 {
          margin: 0;
          font-size: 16px;
          font-weight: 600;
        }

        .warning-header.error {
          background-color: #ffebee;
          color: #c62828;
        }

        .warning-header.warning {
          background-color: #fff3e0;
          color: #e65100;
        }

        .warning-header.info {
          background-color: #e3f2fd;
          color: #1565c0;
        }

        .close-button {
          background: none;
          border: none;
          font-size: 24px;
          cursor: pointer;
          color: inherit;
          padding: 0;
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .close-button:hover {
          opacity: 0.7;
        }

        .warning-content {
          padding: 20px;
        }

        .warning-message {
          margin: 0 0 16px 0;
          font-size: 13px;
          line-height: 1.6;
          color: #333;
        }

        .existing-statement-info,
        .new-statement-info {
          margin-bottom: 16px;
          padding: 12px;
          background-color: #f9f9f9;
          border-radius: 4px;
          border-left: 3px solid #4CAF50;
        }

        .existing-statement-info h4,
        .new-statement-info h4 {
          margin: 0 0 12px 0;
          font-size: 12px;
          font-weight: 600;
          text-transform: uppercase;
          color: #666;
        }

        .info-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;
        }

        .info-item {
          display: flex;
          flex-direction: column;
        }

        .label {
          font-size: 11px;
          font-weight: 500;
          color: #999;
          text-transform: uppercase;
          margin-bottom: 4px;
        }

        .value {
          font-size: 13px;
          color: #333;
          font-weight: 500;
        }

        .warning-actions {
          display: flex;
          gap: 8px;
          padding: 16px 20px;
          border-top: 1px solid #eee;
          justify-content: flex-end;
          flex-wrap: wrap;
        }

        .btn-primary {
          padding: 8px 16px;
          background-color: #4CAF50;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 13px;
          font-weight: 600;
        }

        .btn-primary:hover {
          background-color: #45a049;
        }

        .btn-secondary {
          padding: 8px 16px;
          background-color: #e0e0e0;
          color: #333;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 13px;
          font-weight: 600;
        }

        .btn-secondary:hover {
          background-color: #d0d0d0;
        }

        .btn-tertiary {
          padding: 8px 16px;
          background-color: transparent;
          color: #1976d2;
          border: 1px solid #1976d2;
          border-radius: 4px;
          cursor: pointer;
          font-size: 13px;
          font-weight: 600;
        }

        .btn-tertiary:hover {
          background-color: #f5f5f5;
        }
      `}</style>
    </div>
  );
};
