# Project: MoneyPlan AI

**Goal:** Upload bank statements (PDF/CSV), automatically analyze spending, generate budgets, and provide personalized financial recommendations.

## Features

### Phase 1 - Statement Import

* Upload bank statements from multiple banks
* Support:

  * PDF statements
  * CSV exports
  * Excel files
* Extract:

  * Date
  * Merchant
  * Amount
  * Credit/Debit
  * Balance
  * Description

---

### Phase 2 - Transaction Categorization

Automatically classify transactions.

Example:

| Transaction      | Category      |
| ---------------- | ------------- |
| Swiggy           | Food          |
| Zomato           | Food          |
| Amazon           | Shopping      |
| Uber             | Transport     |
| Salary           | Income        |
| Rent             | Housing       |
| Electricity Bill | Utilities     |
| Netflix          | Entertainment |

Initially use rule-based matching.

Later add an LLM for unknown merchants.

---

### Phase 3 - Spending Dashboard

Show

* Monthly spending
* Spending trend
* Income vs expenses
* Cash flow
* Savings rate
* Net worth trend
* Biggest expense categories

Charts

* Pie chart
* Line chart
* Bar chart
* Calendar heatmap

---

### Phase 4 - AI Financial Advisor

Ask questions like

> Where am I wasting money?

> Can I save ₹20,000/month?

> Compare this month with last month.

> Which subscriptions should I cancel?

> Why was April expensive?

---

### Phase 5 - Budget Planning

User sets goals.

Example

Monthly salary

₹1,50,000

Goal

Save ₹50,000/month

AI generates

Food
₹12,000

Transport
₹5,000

Shopping
₹6,000

Entertainment
₹3,000

Investments
₹50,000

---

### Phase 6 - Predict Future Spending

Use historical data to predict

* End-of-month balance
* Next month's expenses
* Cash shortage warnings
* Upcoming EMIs

---

### Phase 7 - Detect Unusual Transactions

Example

Normal Swiggy

₹400

Today

₹3,800

Flag it.

Same for

* Duplicate payments
* Suspicious charges
* Forgotten subscriptions

---

### Phase 8 - Financial Health Score

Calculate

* Savings rate
* Debt ratio
* Spending volatility
* Emergency fund
* Investment ratio

Return

Financial Score

82/100

---

### Phase 9 - Goals

Examples

* Europe Trip
* Buy Bike
* Buy House
* Emergency Fund

Track

Current

₹2,10,000

Goal

₹8,00,000

ETA

18 months

---

### Phase 10 - AI Insights

Examples

> You spent 18% more on restaurants this month.

> Amazon purchases increased by ₹12,400.

> You can save ₹3,500/month by reducing food delivery.

> Your electricity bill is consistently rising.

---

# LLM Integration

Instead of simple dashboards, let users chat:

> Explain my spending.

> How much do I spend on coffee?

> Find unnecessary expenses.

> Can I afford a car EMI?

> What if my salary increases by 20%?

---

# Architecture

```
Frontend
    React / Next.js

Backend
    Go

API

Transaction Service
Budget Service
AI Insight Service
Forecast Service

Database
PostgreSQL

Search
OpenSearch

Object Storage
S3 / MinIO

LLM
OpenAI / Local Llama

Background Jobs
Temporal / Asynq / Redis

Analytics
DuckDB
```

---

# Interesting Engineering Challenges

* PDF parsing with different bank formats
* Merchant normalization (e.g., "AMZN Mktp", "Amazon Pay", "AMAZON SELLER")
* Duplicate detection
* Transaction categorization
* Embedding transactions for semantic search
* AI-generated financial advice grounded in the user's own data
* Privacy-first design with all processing running locally if desired

---

# Advanced Features

* Multi-bank support
* Credit card + bank account merging
* UPI analysis
* Investment portfolio tracking
* Mutual fund analysis
* EPF/NPS integration
* Tax estimation
* GST/business expense detection
* Receipt OCR
* Shared family budgets
* Email statement import
* WhatsApp spending summaries
* Voice assistant ("How much did I spend on food this month?")

Given your background in Go, AWS, architecture, and your interest in AI-assisted developer tools, this project would let you demonstrate backend design, data pipelines, LLM integration, analytics, and a polished frontend in a single portfolio piece. It can also be built incrementally, with each phase producing a usable feature.
