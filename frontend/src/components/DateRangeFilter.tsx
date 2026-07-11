import React from 'react';

interface DateRangeFilterProps {
  dateFrom: string;
  dateTo: string;
  onDateFromChange: (date: string) => void;
  onDateToChange: (date: string) => void;
  onApply: () => void;
}

export const DateRangeFilter: React.FC<DateRangeFilterProps> = ({
  dateFrom,
  dateTo,
  onDateFromChange,
  onDateToChange,
  onApply,
}) => {
  const today = new Date().toISOString().split('T')[0];

  // Quick filter presets
  const setLast30Days = () => {
    const to = new Date();
    const from = new Date(to.getTime() - 30 * 24 * 60 * 60 * 1000);
    onDateFromChange(from.toISOString().split('T')[0]);
    onDateToChange(to.toISOString().split('T')[0]);
  };

  const setLast90Days = () => {
    const to = new Date();
    const from = new Date(to.getTime() - 90 * 24 * 60 * 60 * 1000);
    onDateFromChange(from.toISOString().split('T')[0]);
    onDateToChange(to.toISOString().split('T')[0]);
  };

  const setLast1Year = () => {
    const to = new Date();
    const from = new Date(to.getTime() - 365 * 24 * 60 * 60 * 1000);
    onDateFromChange(from.toISOString().split('T')[0]);
    onDateToChange(to.toISOString().split('T')[0]);
  };

  const clearDates = () => {
    const to = new Date();
    const from = new Date(to.getTime() - 365 * 24 * 60 * 60 * 1000);
    onDateFromChange(from.toISOString().split('T')[0]);
    onDateToChange(to.toISOString().split('T')[0]);
  };

  return (
    <div className="date-range-filter">
      <h3>Filter by Date Range</h3>

      <div className="quick-filters">
        <button onClick={setLast30Days} className="btn-quick">
          Last 30 Days
        </button>
        <button onClick={setLast90Days} className="btn-quick">
          Last 90 Days
        </button>
        <button onClick={setLast1Year} className="btn-quick">
          Last Year
        </button>
      </div>

      <div className="date-inputs">
        <div className="date-input-group">
          <label htmlFor="dateFrom">From</label>
          <input
            type="date"
            id="dateFrom"
            value={dateFrom}
            onChange={(e) => onDateFromChange(e.target.value)}
            max={today}
            className="date-input"
          />
        </div>

        <div className="date-input-group">
          <label htmlFor="dateTo">To</label>
          <input
            type="date"
            id="dateTo"
            value={dateTo}
            onChange={(e) => onDateToChange(e.target.value)}
            max={today}
            className="date-input"
          />
        </div>
      </div>

      <div className="filter-actions">
        <button onClick={onApply} className="btn-primary">
          Apply Filter
        </button>
        <button onClick={clearDates} className="btn-secondary">
          Reset
        </button>
      </div>

      <style jsx>{`
        .date-range-filter {
          padding: 16px;
          background-color: #f9f9f9;
          border-radius: 8px;
          margin-bottom: 16px;
        }

        .date-range-filter h3 {
          margin-top: 0;
          font-size: 14px;
          font-weight: 600;
          text-transform: uppercase;
          color: #333;
        }

        .quick-filters {
          display: flex;
          gap: 8px;
          margin-bottom: 16px;
          flex-wrap: wrap;
        }

        .btn-quick {
          padding: 6px 12px;
          background-color: #f0f0f0;
          border: 1px solid #ddd;
          border-radius: 4px;
          cursor: pointer;
          font-size: 12px;
          font-weight: 500;
          color: #666;
          transition: all 0.2s;
        }

        .btn-quick:hover {
          background-color: #e8e8e8;
          border-color: #ccc;
        }

        .date-inputs {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;
          margin-bottom: 12px;
        }

        .date-input-group {
          display: flex;
          flex-direction: column;
        }

        .date-input-group label {
          font-size: 12px;
          font-weight: 500;
          margin-bottom: 4px;
          color: #666;
        }

        .date-input {
          padding: 8px;
          border: 1px solid #ddd;
          border-radius: 4px;
          font-size: 13px;
          font-family: inherit;
        }

        .date-input:focus {
          outline: none;
          border-color: #4CAF50;
          box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.1);
        }

        .filter-actions {
          display: flex;
          gap: 8px;
        }

        .btn-primary {
          flex: 1;
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
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 13px;
          font-weight: 500;
        }

        .btn-secondary:hover {
          background-color: #d0d0d0;
        }
      `}</style>
    </div>
  );
};
