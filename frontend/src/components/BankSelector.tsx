import React, { useState } from 'react';

export interface BankOption {
  code: string;
  name: string;
  logo?: string;
}

const DEFAULT_BANKS: BankOption[] = [
  { code: 'HDFC', name: 'HDFC Bank' },
  { code: 'ICIC', name: 'ICICI Bank' },
  { code: 'AXIS', name: 'Axis Bank' },
  { code: 'SBI', name: 'State Bank of India' },
  { code: 'KOTAK', name: 'Kotak Mahindra Bank' },
  { code: 'YES', name: 'YES Bank' },
];

interface BankSelectorProps {
  value: string;
  onChange: (bankCode: string) => void;
  banks?: BankOption[];
  required?: boolean;
}

export function BankSelector({
  value,
  onChange,
  banks = DEFAULT_BANKS,
  required = false,
}: BankSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);

  const selectedBank = banks.find((b) => b.code === value);

  return (
    <div className="relative w-full">
      <label htmlFor="bank-select" className="block text-sm font-medium text-gray-700 mb-2">
        Select Bank {required && <span className="text-red-500">*</span>}
      </label>

      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-4 py-2 text-left bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      >
        <div className="flex items-center justify-between">
          <span className={selectedBank ? 'text-gray-900' : 'text-gray-500'}>
            {selectedBank?.name || 'Select a bank...'}
          </span>
          <svg
            className={`w-5 h-5 transition-transform ${isOpen ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 14l-7 7m0 0l-7-7m7 7V3"
            />
          </svg>
        </div>
      </button>

      {isOpen && (
        <div className="absolute z-10 w-full mt-2 bg-white border border-gray-300 rounded-lg shadow-lg">
          <ul className="max-h-60 overflow-auto">
            {banks.map((bank) => (
              <li key={bank.code}>
                <button
                  type="button"
                  onClick={() => {
                    onChange(bank.code);
                    setIsOpen(false);
                  }}
                  className={`w-full text-left px-4 py-2 hover:bg-blue-50 transition-colors ${
                    value === bank.code ? 'bg-blue-100 text-blue-900 font-medium' : ''
                  }`}
                >
                  {bank.name}
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
