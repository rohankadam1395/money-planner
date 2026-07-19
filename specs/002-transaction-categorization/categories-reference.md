# Predefined Categories Reference

**Feature**: 002-Transaction-Categorization | **Version**: 1.0 | **Date**: 2026-07-18

This document defines the 10 predefined categories used across transaction categorization, analytics, and UI components. All categories are fixed for Phase 2 MVP; custom categories deferred to Phase 3+.

---

## Category Definitions

| ID | Name | Description | Color | Icon | Examples | Phase 2 |
|----|------|-------------|-------|------|----------|---------|
| 1 | **Food & Dining** | Restaurants, cafes, food delivery, groceries | `#FF6B6B` (Red) | 🍔 | Swiggy, Zomato, Big Basket, McDonald's, local restaurants | ✅ |
| 2 | **Shopping** | Retail, clothing, home goods, online marketplaces | `#4ECDC4` (Teal) | 🛍️ | Amazon, Flipkart, H&M, Nike, Myntra | ✅ |
| 3 | **Transport** | Ride-sharing, public transport, fuel, car services | `#45B7D1` (Blue) | 🚗 | Uber, Ola, Metro, toll, fuel pump (Shell, HP) | ✅ |
| 4 | **Housing** | Rent, mortgage, property taxes, home maintenance | `#F7B731` (Orange) | 🏠 | Landlord transfer, property tax, home repairs | ✅ |
| 5 | **Utilities** | Electricity, water, internet, phone, gas | `#5F27CD` (Purple) | 💡 | BSNL, Airtel broadband, electricity board, gas bill | ✅ |
| 6 | **Entertainment** | Movies, streaming, games, events, hobbies | `#EE5A6F` (Pink) | 🎬 | Netflix, Spotify, Amazon Prime, cinema tickets, concert | ✅ |
| 7 | **Income** | Salary, freelance, investment returns, refunds | `#2ECC71` (Green) | 💰 | Employer deposit, freelance payment, interest, refund | ✅ |
| 8 | **Healthcare** | Medical expenses, pharmacy, insurance, gym | `#FF4757` (Bright Red) | 🏥 | Pharmacy, hospital, insurance premium, gym membership | ✅ |
| 9 | **Education** | Tuition, courses, books, training | `#1E90FF` (Dodger Blue) | 📚 | Tuition fees, online courses, books, coaching | ✅ |
| 10 | **Miscellaneous** | Uncategorized, gifts, charity, other expenses | `#95A5A6` (Gray) | 📌 | Donations, gifts, tips, other unclassified | ✅ |

---

## Merchant Dictionary Seed Examples

**Food & Dining** (ID: 1):
- Exact: Swiggy, Zomato, Domino's, KFC, McDonald's, Starbucks, Café Coffee Day
- Fuzzy: SWIGGY, Swiggy FD, dominos pizza, mcd

**Shopping** (ID: 2):
- Exact: Amazon, Flipkart, H&M, Nike, Myntra, Uniqlo, Ikea
- Fuzzy: AMAZON, amazon.in, flipkart.com

**Transport** (ID: 3):
- Exact: Uber, Ola, Metro, Shell, HP, GoIbibo, MakeMyTrip
- Fuzzy: UBER TRIP, olacabs, metro toll

**Housing** (ID: 4):
- Exact: Rent transfer (common landlord names), property tax board
- Fuzzy: RENT, landlord, property

**Utilities** (ID: 5):
- Exact: BSNL, Airtel, Jio, electricity board, gas company
- Fuzzy: AIRTEL, bsnl broadband, power bill

**Entertainment** (ID: 6):
- Exact: Netflix, Spotify, Amazon Prime Video, Disney+, YouTube Premium, Bookmyshow
- Fuzzy: netflix.com, spotify subscription

**Income** (ID: 7):
- Pattern-based: Credit >0, source matches employer name or "Salary", "Freelance", "Interest", "Dividend"

**Healthcare** (ID: 8):
- Exact: Apollo, Fortis, Max Hospital, CVS Pharmacy, Ayurveda clinic
- Fuzzy: PHARMACY, hospital, medical

**Education** (ID: 9):
- Exact: Coursera, Udemy, BYJU'S, coaching centers
- Fuzzy: COURSE, tuition, training

**Miscellaneous** (ID: 10):
- Default fallback for unmatched transactions

---

## Implementation Notes

### Database Seeding
- Initialize `categories` table with these 10 rows (T006 migration)
- Seed `merchant_dictionary` table with ≥500 entries from examples above (T027 seed task)
- Use `is_predefined=true` flag for these 10 categories (no user-created categories in Phase 2)

### Frontend Color Scheme
- Use `category.color` field for CategoryBadge and dashboard visualizations
- Fallback to color in this reference if database field missing
- Ensure contrast ratio ≥4.5:1 against white/dark backgrounds (WCAG AA)

### Merchant Matching Strategy (Phase 2 MVP)
1. **Exact match**: Lookup merchant name in dictionary (case-insensitive)
2. **Fuzzy match**: Levenshtein distance ≥85% (handles "SWIGGY" vs "Swiggy", extra spaces, etc.)
3. **No match**: Mark as "Uncategorized" (Phase 4+ will call LLM)

### Confidence Scoring (Phase 2 MVP)
- Exact match → Confidence: 1.0 (100%)
- Fuzzy match → Confidence: 0.85-0.99 (based on Levenshtein distance)
- No match → Confidence: 0.0 (Uncategorized)

### Phase 2 Constraints
- Cannot modify these 10 categories (fixed)
- Cannot add new predefined categories
- User corrections update merchant_dictionary, not category definitions
- Custom categories deferred to Phase 3+

---

## Related Documents

- `spec.md` — Functional requirements and user stories
- `plan.md` — Implementation architecture and project structure
- `tasks.md` — Implementation tasks (T006 categories table, T027 merchant seeding)
- `data-model.md` — Database schema for categories table
- Backend: `backend/internal/categorization/models.go` — Category struct
- Frontend: `frontend/src/components/CategoryBadge.tsx` — Category UI component

