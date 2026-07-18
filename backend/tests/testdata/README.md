# Test Data for Statement Import

This directory contains sample bank statement data for testing the statement import feature.

## Files

### `hdfc_sample.csv`
Sample HDFC Bank statement with 29 transactions covering:
- **Period**: July 1-29, 2026
- **Starting Balance**: ₹50,000
- **Ending Balance**: ₹116,900.51
- **Transaction Types**: Salary, Bills, Shopping, ATM, Investments
- **Columns**: Date, Narration, Debit, Credit, Balance

**Key Features**:
- Multiple salary deposits
- Various expense categories (utilities, food, travel, insurance, etc.)
- Interest credits
- Large transactions (loan, investment)

---

### `icici_sample.csv`
Sample ICICI Bank statement with 29 transactions covering:
- **Period**: July 1-29, 2026
- **Starting Balance**: ₹60,000
- **Ending Balance**: ₹143,851.26
- **Column Names**: Txn Date, Particulars, Withdrawal, Deposit, Balance

**Key Features**:
- Different column naming convention (Withdrawal/Deposit instead of Debit/Credit)
- Detailed narration format
- Performance bonus and fixed deposit interest
- Real-world transaction descriptions

---

### `axis_sample.csv`
Sample Axis Bank statement with 29 transactions covering:
- **Period**: July 1-29, 2026
- **Starting Balance**: ₹45,000
- **Ending Balance**: ₹138,452
- **Column Names**: Value Date, Transaction Description, Debit Amount, Credit Amount, Running Balance

**Key Features**:
- Alternative column naming ("Value Date" vs "Date")
- Running balance terminology
- SIP investments
- FD maturity credits

---

### `categorization_sample.csv`
Sample bank statement with merchants for testing **transaction categorization** feature:
- **Period**: July 1-30, 2026
- **Starting Balance**: ₹50,000
- **Ending Balance**: ₹203,889
- **Transactions**: 29 transactions with known & unknown merchants
- **Purpose**: Testing rule-based and fuzzy matching categorization

**Key Features** (All merchants seeded in merchant_dictionary):
- **Food & Dining**: Swiggy, Zomato, Uber Eats (exact & fuzzy: "SWIGGY FD")
- **Shopping**: Amazon, Flipkart, Shopify
- **Transport**: Uber, Ola
- **Entertainment**: Netflix, Spotify
- **Utilities**: BSNL, Airtel
- **Healthcare**: Apollo Hospital
- **Education**: Coursera
- **Unknown Merchants**: "Unknown Merchant XYZ" (tests uncategorized flow)
- **Salary & Investments**: Multiple deposits for income testing

**Expected Categorization Results**:
- 23 transactions with rule-based category (exact merchant match)
- 1 transaction with fuzzy match (SWIGGY FD → Swiggy → Food)
- 1 transaction uncategorized (Unknown Merchant XYZ)
- 4 transactions with no debit/credit (salary/opening/closing)

---

## Testing Scenarios

### Scenario 1: Single Bank Upload (HDFC)
1. Upload `hdfc_sample.csv`
2. Expected: 28 valid transactions extracted (excluding opening/closing balances)
3. Validation: All transactions should pass validation
4. Period: 2026-07-01 to 2026-07-29

### Scenario 2: Different Bank Format (ICICI)
1. Upload `icici_sample.csv`
2. Expected: Parser should handle different column names
3. Validation: Should extract withdrawal/deposit as debit/credit
4. Column Flexibility Test

### Scenario 3: Alternative Format (Axis)
1. Upload `axis_sample.csv`
2. Expected: "Value Date" and "Running Balance" column names
3. Validation: Format flexibility verification

### Scenario 4: Multi-Bank Consolidation
1. Upload all three samples
2. Expected: 84 total transactions across 3 banks
3. Validation: Bank codes properly assigned (HDFC, ICICI, AXIS)
4. Query: Filter by bank_code to verify isolation

### Scenario 5: Duplicate Detection
1. Upload `hdfc_sample.csv`
2. Upload same file again
3. Expected: 409 Conflict response on second upload
4. Validation: File hash matching works

### Scenario 6: Date Range Filtering
1. Upload multiple statements
2. Query: Filter transactions between 2026-07-10 and 2026-07-20
3. Expected: ~8-10 transactions in date range per statement

### Scenario 7: Transaction Categorization (Rule-Based MVP)
1. Upload `categorization_sample.csv`
2. Go to statement preview
3. Expected Results:
   - ✅ Swiggy → Food & Dining (confidence: 1.0, method: rule_based)
   - ✅ Amazon → Shopping (confidence: 1.0, method: rule_based)
   - ✅ Zomato → Food & Dining (confidence: 1.0, method: rule_based)
   - ✅ Uber → Transport (confidence: 1.0, method: rule_based)
   - ✅ Netflix → Entertainment (confidence: 1.0, method: rule_based)
   - ✅ BSNL → Utilities (confidence: 1.0, method: rule_based)
   - ✅ SWIGGY FD DELIVERY → Food & Dining (confidence: 0.85-0.99, method: fuzzy)
   - ✅ Unknown Merchant XYZ → Uncategorized (confidence: 0.0, method: none)
4. Validation: Categories shown with badges, confidence scores, and categorization method

### Scenario 8: Fuzzy Matching Accuracy
1. Upload `categorization_sample.csv`
2. Verify "SWIGGY FD" (with typo/variation) matches "Swiggy" in merchant dictionary
3. Expected: Levenshtein distance ≥85% triggers fuzzy match
4. Validation: Confidence score between 0.85-0.99

---

## Data Quality Metrics

| Bank | Transactions | Amount Range | Avg Amount | Coverage |
|------|--------------|--------------|-----------|----------|
| HDFC | 28 | ₹150 - ₹75,000 | ₹8,543 | 29 days (100%) |
| ICICI | 28 | ₹250 - ₹80,000 | ₹9,421 | 29 days (100%) |
| Axis | 28 | ₹200 - ₹70,000 | ₹8,152 | 29 days (100%) |

---

## How to Use These Files

### Manual Testing
1. Start the backend server:
   ```bash
   cd backend
   go run cmd/statement-import-api/main.go
   ```

2. Upload a test file via the frontend:
   - Navigate to `/statements/upload`
   - Select one of these CSV files
   - Choose the corresponding bank
   - Verify extraction and import

### Automated Testing
```bash
# Test HDFC parser
curl -X POST http://localhost:8080/api/statements/upload \
  -F "file=@hdfc_sample.csv" \
  -F "bank_code=HDFC" \
  -H "Authorization: Bearer <token>"

# Expected Response (202 Accepted):
# {
#   "statement_id": "uuid",
#   "status": "PENDING",
#   "bank_code": "HDFC",
#   "file_name": "hdfc_sample.csv"
# }
```

### Integration Test Enhancement
Update `backend/tests/integration/csv_parser_test.go` to use these files:
```go
// Read actual file
data, _ := ioutil.ReadFile("testdata/hdfc_sample.csv")
transactions, _ := parser.ParseCSV(bytes.NewReader(data))

// Verify 28 transactions extracted
assert.GreaterOrEqual(t, len(transactions), 28)
```

---

## Transaction Categories Represented

- **Income**: Salary, Bonus, Interest, Dividends, FD Maturity
- **Utilities**: Electricity, Water, Internet, Telecom, DTH
- **Transportation**: Fuel, Travel Booking, Train Tickets
- **Shopping**: Groceries, Apparel, Electronics, Online
- **Services**: Insurance, Gym, Subscriptions, Dining
- **Investments**: Mutual Funds, Fixed Deposits, SIP
- **Housing**: Rent, Loan EMI, Home Maintenance
- **Healthcare**: Medical, Pharmacy, Insurance
- **Others**: ATM Withdrawal, Bank Charges, Movie Tickets

---

## Notes

- All dates are in July 2026 to avoid timezone issues
- All amounts are realistic for Indian bank accounts
- Column name variations test parser flexibility
- Statement periods are exactly 29 days (month-like period)
- Balance consistency is maintained throughout
- No malformed data or extreme outliers (for happy-path testing)

---

## Future Enhancements

- PDF versions of same data (for PDF parser testing)
- Excel format versions
- Edge case data (very large transactions, special characters)
- Multi-month statements
- Statements with missing fields for validation testing
