# Transaction Recategorization - Issues Fixed & Test Cases

## Overview
This document describes all the issues discovered and fixed during the implementation of the transaction recategorization feature, particularly for handling uncategorized transactions.

---

## Issue 1: Uncategorized Transactions Not Visible
**Problem:** After importing transactions, users couldn't see uncategorized transactions because there was no way to query or display them.

**Root Cause:** 
- Uncategorized transactions don't have an entry in `transaction_categories` table
- The categories list only showed predefined categories
- No endpoint to view uncategorized transactions

**Solution:**
1. Added special handling for `category_id = "uncategorized"` in `HandleGetCategoryTransactions`
2. Query transactions with no `transaction_categories` entry
3. Added uncategorized card to categories list with count and total

**Code Changes:**
- `backend/internal/api/categories.go`: Added uncategorized card to `HandleGetCategories` (lines 139-163)
- `backend/internal/api/categories.go`: Added special case in `HandleGetCategoryTransactions` for `categoryID == "uncategorized"` (lines 222-259)

**Test Cases (T101):**
- Uncategorized transactions count appears in categories list
- Uncategorized card only appears when transactions exist
- Uncategorized card shows correct transaction count and total

---

## Issue 2: Recategorizing Uncategorized Transactions Failed
**Problem:** When trying to recategorize an uncategorized transaction, the endpoint returned 404 "transaction category not found".

**Root Cause:**
- The `recategorize` endpoint tried to UPDATE an existing row in `transaction_categories`
- For uncategorized transactions, no row exists yet
- The SELECT query failed with `sql.ErrNoRows`, triggering a 404 error

**Solution:**
Changed the logic to:
1. Check if the transaction already has a category assignment
2. If not (uncategorized): **INSERT** new row into `transaction_categories`
3. If yes (already categorized): UPDATE existing row

**Code Changes:**
- `backend/internal/api/recategorize.go`: Changed from INSERT-only to INSERT/UPDATE logic (lines 125-145)
- Used `sql.NullString` to detect whether a category exists

**Test Cases (T102):**
- INSERT creates new transaction_categories entry when uncategorized
- old_category_name is "Uncategorized" when moving from uncategorized
- category_stats updated for new category after recategorization

---

## Issue 3: Category Stats Not Updating After Recategorization
**Problem:** After recategorizing a transaction, the categories page still showed old counts and totals.

**Root Cause:**
- `category_stats` is a cached/denormalized table for performance
- When transactions were recategorized, the stats table wasn't updated
- Frontend continued to display stale data

**Solution:**
Added stats recalculation after every recategorization:
1. Calculate current stats by querying actual `transaction_categories` data
2. Upsert results into `category_stats` table
3. Update stats for both old and new categories

**Code Changes:**
- `backend/internal/api/recategorize.go`: Added `updateCategoryStats` helper function (lines 194-213)
- Called after INSERT/UPDATE to recalculate both categories (lines 154-156)
- Used PostgreSQL `ON CONFLICT ... DO UPDATE` (UPSERT) for atomic updates

**Test Cases (T103, T105):**
- Category_stats updated for both old and new categories
- Stats match actual transaction data sum
- Multiple recategorizations result in correct final stats

---

## Issue 4: Missing User ID Filter in Uncategorized Query
**Problem:** Transactions from other users could appear in uncategorized list.

**Root Cause:**
- The NOT EXISTS subquery wasn't filtering by `user_id`
- It checked if transaction_id exists anywhere in `transaction_categories`
- Should have checked if it exists for the current user

**Solution:**
Added `AND tc.user_id = $1` to the NOT EXISTS subquery to filter by current user.

**Code Changes:**
- `backend/internal/api/categories.go`: Added user_id filter in uncategorized query (line 229)

**Test Cases (T106):**
- Uncategorized query filters by user_id in NOT EXISTS subquery
- Uncategorized count accurate per user

---

## Issue 5: Recategorizing to "Uncategorized" Failed
**Problem:** When trying to recategorize a transaction back to "uncategorized", the endpoint threw "failed to retrieve new category" error.

**Root Cause:**
- "uncategorized" is a synthetic/virtual category ID
- It doesn't exist in the `categories` table
- The endpoint tried to query the database for it

**Solution:**
1. Handle "uncategorized" as a special case
2. Don't look it up in the database
3. When recategorizing to "uncategorized": **DELETE** the transaction_categories entry (not INSERT)
4. Validate that other categories exist before accepting them

**Code Changes:**
- `backend/internal/api/recategorize.go`: Added validation before query (lines 73-85)
- Added special case for "uncategorized" to DELETE instead of INSERT/UPDATE (lines 153-168)

**Test Cases (T104):**
- DELETE removes transaction_categories entry when recategorizing to uncategorized
- "uncategorized" is treated as special case, not real category
- Transaction appears in uncategorized list after recategorization

---

## Issue 6: Wrong Column Names in Transaction Categories Insert
**Problem:** INSERT to `transaction_categories` was failing with DB_ERROR.

**Root Cause:**
- Using wrong column names: `created_at` (doesn't exist, should be `assigned_at`)
- Trying to specify `id` column (should auto-generate)
- Schema has `assigned_at` DEFAULT CURRENT_TIMESTAMP, `id` auto-generates with gen_random_uuid()

**Solution:**
Corrected the INSERT statement:
- Remove `id` column (auto-generates)
- Change `created_at` to `assigned_at` or omit it (uses default)
- Only specify required/custom columns

**Code Changes:**
- `backend/internal/api/recategorize.go`: Fixed INSERT column names (line 129-130)

**Test Cases (T110):**
- transaction_categories.id auto-generates with gen_random_uuid()
- transaction_categories uses assigned_at not created_at

---

## Data Flow Diagram

### Before (Broken)
```
User categorizes transaction from "Uncategorized" to "Healthcare"
  ↓
Recategorize endpoint tries UPDATE on non-existent row
  ↓
404 Error (transaction category not found)
  ↓
No change to database
```

### After (Fixed)
```
User categorizes transaction from "Uncategorized" to "Healthcare"
  ↓
Recategorize endpoint checks if category exists
  ↓
No row found → INSERT new entry
  ↓
Recalculate stats for Healthcare category
  ↓
Update category_stats table with new totals
  ↓
Frontend shows updated counts immediately
```

---

## Test Case Coverage Matrix

| Test Case | Issue(s) | Coverage |
|-----------|----------|----------|
| T101: Uncategorized Visibility | #1 | Uncategorized appears in list |
| T102: Recategorize From Uncategorized | #2, #3 | INSERT + stats update |
| T103: Recategorize Between Categories | #2, #3 | UPDATE + dual stats update |
| T104: Recategorize To Uncategorized | #5 | DELETE + stats update |
| T105: Stats Accuracy | #3 | Recalculation from actual data |
| T106: User Isolation | #4 | User ID filtering |
| T107: Edge Cases | #2, #5 | Error handling |
| T108: View Category Transactions | #1 | Query correctness |
| T109: Learn Correction | #2 | Merchant dictionary update |
| T110: Schema Requirements | #6 | Column names, constraints |

---

## Performance Considerations

### Category Stats Caching
- **Why:** Calculating SUM/COUNT/AVG on every page load is expensive with large datasets
- **Tradeoff:** Stats can be stale after changes
- **Solution:** Recalculate on every transaction change (INSERT/UPDATE/DELETE in transaction_categories)

### Uncategorized Query Efficiency
- **Current:** NOT EXISTS subquery
- **Alternative:** LEFT JOIN with IS NULL check
- **Note:** Performance is adequate for current scale

---

## Future Improvements

1. **Batch Recategorization:** Allow multiple transactions to be recategorized at once
2. **Undo Recategorization:** Store history of category changes
3. **Stats Refresh Trigger:** Consider database trigger to auto-update stats
4. **Merchant Learning:** Implement machine learning feedback loop for better categorization
5. **Category Merge/Split:** Allow users to manage custom categories

---

## Database Constraints Verified

✅ `transaction_categories.id` - Auto-generates with gen_random_uuid()
✅ `transaction_categories.UNIQUE(transaction_id)` - One category per transaction
✅ `category_stats.UNIQUE(user_id, category_id, period)` - One stats row per category/period/user
✅ `transaction_categories.user_id` NOT NULL - Enforces user ownership
✅ `transaction_categories.category_id` NOT NULL - Can't have null category (must DELETE for uncategorized)
