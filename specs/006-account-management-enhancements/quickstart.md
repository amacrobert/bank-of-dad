# Quickstart: Account Management Enhancements

## Prerequisites

- Backend running on :8080
- Frontend running on :5173
- A parent logged in with at least one child in the family
- Child has some balance (deposit via Manage > Deposit first)

## Verification Steps

### 1. Parent Views Transaction History

1. Log in as parent
2. Click "Manage" on a child with transactions
3. Verify transaction history appears below the balance section
4. Confirm deposits, withdrawals, allowances, and interest transactions all display correctly

### 2. Interest Rate Pre-populated

1. Set an interest rate for a child (e.g., 5%)
2. Close the manage view
3. Re-open manage for the same child
4. Verify the interest rate field shows "5.00" (not "0.00")

### 3. Allowance in Child Management

1. Open manage child view
2. Verify the standalone "Allowance Schedules" section is gone from the dashboard
3. In the manage view, see the allowance section
4. Create a $10 weekly allowance on Friday
5. Close and re-open manage — verify the allowance is pre-populated
6. Pause the allowance — verify it shows as paused
7. Remove the allowance — verify it returns to "no allowance" state

### 4. One Allowance Per Child

1. Via API, attempt to create a second allowance for a child that already has one
2. Verify the system returns an error

### 5. Interest Accrual Schedule

1. Open manage child
2. Set interest rate to 5%
3. Configure interest schedule: monthly on the 1st
4. Verify the schedule is saved
5. Change to weekly on Sunday
6. Verify the schedule updates

### 6. Child Dashboard — Interest Info

1. Log in as the child
2. Verify the dashboard shows "5.00% annual interest"
3. Verify "Next interest payment: [date]" appears
4. If no interest rate is set, verify no interest info appears

### 7. API Verification

```bash
# Get child's allowance
curl -b cookies.txt http://localhost:8080/api/children/1/allowance

# Set allowance
curl -b cookies.txt -X PUT http://localhost:8080/api/children/1/allowance \
  -H "Content-Type: application/json" \
  -d '{"amount_cents": 1000, "frequency": "weekly", "day_of_week": 5}'

# Set interest schedule
curl -b cookies.txt -X PUT http://localhost:8080/api/children/1/interest-schedule \
  -H "Content-Type: application/json" \
  -d '{"frequency": "monthly", "day_of_month": 1}'

# Get interest schedule
curl -b cookies.txt http://localhost:8080/api/children/1/interest-schedule

# Check balance includes next_interest_at
curl -b cookies.txt http://localhost:8080/api/children/1/balance
```
