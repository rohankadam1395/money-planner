import React from 'react';

interface BankFilterProps {
  availableBanks: string[];
  selectedBanks: string[];
  onBankChange: (banks: string[]) => void;
}

const BANK_NAMES: { [key: string]: string } = {
  HDFC: 'HDFC Bank',
  ICICI: 'ICICI Bank',
  AXIS: 'Axis Bank',
  SBI: 'State Bank of India',
};

export const BankFilter: React.FC<BankFilterProps> = ({
  availableBanks,
  selectedBanks,
  onBankChange,
}) => {
  const toggleBank = (bankCode: string) => {
    if (selectedBanks.includes(bankCode)) {
      onBankChange(selectedBanks.filter((b) => b !== bankCode));
    } else {
      onBankChange([...selectedBanks, bankCode]);
    }
  };

  const selectAll = () => {
    onBankChange(availableBanks);
  };

  const clearAll = () => {
    onBankChange([]);
  };

  return (
    <div className="bank-filter">
      <h3>Filter by Bank</h3>

      <div className="filter-controls">
        <button onClick={selectAll} className="btn-secondary btn-sm">
          Select All
        </button>
        <button onClick={clearAll} className="btn-secondary btn-sm">
          Clear All
        </button>
      </div>

      <div className="bank-checkboxes">
        {availableBanks.map((bankCode) => (
          <label key={bankCode} className="checkbox-label">
            <input
              type="checkbox"
              checked={selectedBanks.includes(bankCode)}
              onChange={() => toggleBank(bankCode)}
              className="checkbox"
            />
            <span className="bank-name">
              {BANK_NAMES[bankCode] || bankCode}
            </span>
            <span className="bank-code">({bankCode})</span>
          </label>
        ))}
      </div>

      <style jsx>{`
        .bank-filter {
          padding: 16px;
          background-color: #f9f9f9;
          border-radius: 8px;
          margin-bottom: 16px;
        }

        .bank-filter h3 {
          margin-top: 0;
          font-size: 14px;
          font-weight: 600;
          text-transform: uppercase;
          color: #333;
        }

        .filter-controls {
          display: flex;
          gap: 8px;
          margin-bottom: 12px;
        }

        .btn-secondary {
          padding: 6px 12px;
          background-color: #e0e0e0;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 12px;
          font-weight: 500;
        }

        .btn-secondary:hover {
          background-color: #d0d0d0;
        }

        .btn-sm {
          font-size: 11px;
          padding: 4px 8px;
        }

        .bank-checkboxes {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: 8px;
        }

        .checkbox-label {
          display: flex;
          align-items: center;
          gap: 8px;
          cursor: pointer;
          padding: 8px;
          border-radius: 4px;
        }

        .checkbox-label:hover {
          background-color: #f0f0f0;
        }

        .checkbox {
          width: 16px;
          height: 16px;
          cursor: pointer;
        }

        .bank-name {
          font-size: 13px;
          font-weight: 500;
          color: #333;
        }

        .bank-code {
          font-size: 11px;
          color: #999;
        }
      `}</style>
    </div>
  );
};
