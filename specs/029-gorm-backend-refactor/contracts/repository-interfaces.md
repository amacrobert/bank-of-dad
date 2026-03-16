# Repository Interfaces: GORM Backend Refactor

**Date**: 2026-03-16 | **Feature**: 029-gorm-backend-refactor

> These define the method signatures for each repository in `backend/repositories/`. Signatures mirror the existing store methods to minimize handler changes. All repositories accept `*gorm.DB` in their constructor.

## FamilyRepo

```go
type FamilyRepo struct { db *gorm.DB }

func NewFamilyRepo(db *gorm.DB) *FamilyRepo

func (r *FamilyRepo) Create(slug string) (*models.Family, error)
func (r *FamilyRepo) GetByID(id int64) (*models.Family, error)           // nil,nil if not found
func (r *FamilyRepo) GetBySlug(slug string) (*models.Family, error)       // nil,nil if not found
func (r *FamilyRepo) SlugExists(slug string) (bool, error)
func (r *FamilyRepo) SuggestSlugs(base string) []string
func (r *FamilyRepo) UpdateTimezone(familyID int64, timezone string) error
func (r *FamilyRepo) UpdateBankName(familyID int64, bankName string) error
func (r *FamilyRepo) GetByStripeCustomerID(customerID string) (*models.Family, error)
func (r *FamilyRepo) UpdateStripeCustomer(familyID int64, customerID string) error
func (r *FamilyRepo) UpdateSubscription(familyID int64, subID, status string, periodEnd time.Time, cancelAtEnd bool) error
func (r *FamilyRepo) UpdateAccountType(familyID int64, accountType string) error
```

## ParentRepo

```go
type ParentRepo struct { db *gorm.DB }

func NewParentRepo(db *gorm.DB) *ParentRepo

func (r *ParentRepo) GetOrCreate(googleID, email, displayName string) (*models.Parent, bool, error)
func (r *ParentRepo) GetByID(id int64) (*models.Parent, error)
func (r *ParentRepo) SetFamily(parentID, familyID int64) error
func (r *ParentRepo) GetByFamilyID(familyID int64) (*models.Parent, error)
```

## ChildRepo

```go
type ChildRepo struct { db *gorm.DB }

func NewChildRepo(db *gorm.DB) *ChildRepo

func (r *ChildRepo) Create(familyID int64, firstName, passwordHash string) (*models.Child, error)
func (r *ChildRepo) GetByID(id int64) (*models.Child, error)
func (r *ChildRepo) GetByName(familyID int64, firstName string) (*models.Child, error)
func (r *ChildRepo) ListByFamily(familyID int64) ([]models.Child, error)
func (r *ChildRepo) GetBalance(id int64) (int64, error)
func (r *ChildRepo) UpdatePassword(id int64, passwordHash string) error
func (r *ChildRepo) UpdateInterestRate(id int64, rateBps int) error
func (r *ChildRepo) UpdateAvatar(id int64, avatar *string) error
func (r *ChildRepo) UpdateTheme(id int64, theme *string) error
func (r *ChildRepo) IncrementFailedLogins(id int64) error
func (r *ChildRepo) Lock(id int64) error
func (r *ChildRepo) ResetFailedLogins(id int64) error
func (r *ChildRepo) SetDisabled(id int64, disabled bool) error
func (r *ChildRepo) CountByFamily(familyID int64) (int, error)
func (r *ChildRepo) DeleteAll(familyID, parentID int64) error
```

## TransactionRepo

```go
type TransactionRepo struct { db *gorm.DB }

func NewTransactionRepo(db *gorm.DB) *TransactionRepo

func (r *TransactionRepo) Deposit(childID, parentID, amountCents int64, note string) (*models.Transaction, int64, error)
func (r *TransactionRepo) Withdraw(childID, parentID, amountCents int64, note string) (*models.Transaction, int64, error)
func (r *TransactionRepo) ListByChild(childID int64, limit, offset int) ([]models.Transaction, error)
func (r *TransactionRepo) GetMonthlySummary(childID int64) ([]MonthlySummary, error)
func (r *TransactionRepo) CreateAllowanceTransaction(childID, parentID, amountCents int64, note string, scheduleID int64) (*models.Transaction, int64, error)
func (r *TransactionRepo) CreateInterestTransaction(childID, parentID, amountCents int64) (*models.Transaction, int64, error)
```

## ScheduleRepo (Allowance)

```go
type ScheduleRepo struct { db *gorm.DB }

func NewScheduleRepo(db *gorm.DB) *ScheduleRepo

func (r *ScheduleRepo) Create(childID, parentID, amountCents int64, frequency string, dayOfWeek, dayOfMonth *int, note string, nextRunAt time.Time) (*models.AllowanceSchedule, error)
func (r *ScheduleRepo) GetByID(id int64) (*models.AllowanceSchedule, error)
func (r *ScheduleRepo) ListByChild(childID int64) ([]models.AllowanceSchedule, error)
func (r *ScheduleRepo) ListByFamily(familyID int64) ([]models.AllowanceSchedule, error)
func (r *ScheduleRepo) Update(id int64, amountCents int64, frequency string, dayOfWeek, dayOfMonth *int, note string, nextRunAt time.Time) error
func (r *ScheduleRepo) UpdateStatus(id int64, status string) error
func (r *ScheduleRepo) Delete(id int64) error
func (r *ScheduleRepo) ListDue(now time.Time) ([]models.AllowanceSchedule, error)
func (r *ScheduleRepo) UpdateNextRun(id int64, nextRunAt time.Time) error
```

## InterestScheduleRepo

```go
type InterestScheduleRepo struct { db *gorm.DB }

func NewInterestScheduleRepo(db *gorm.DB) *InterestScheduleRepo

func (r *InterestScheduleRepo) Create(childID, parentID int64, frequency string, dayOfWeek, dayOfMonth *int, nextRunAt time.Time) (*models.InterestSchedule, error)
func (r *InterestScheduleRepo) GetByChildID(childID int64) (*models.InterestSchedule, error)
func (r *InterestScheduleRepo) ListByFamily(familyID int64) ([]models.InterestSchedule, error)
func (r *InterestScheduleRepo) Update(id int64, frequency string, dayOfWeek, dayOfMonth *int, nextRunAt time.Time) error
func (r *InterestScheduleRepo) UpdateStatus(id int64, status string) error
func (r *InterestScheduleRepo) Delete(id int64) error
func (r *InterestScheduleRepo) ListDue(now time.Time) ([]models.InterestSchedule, error)
func (r *InterestScheduleRepo) UpdateNextRun(id int64, nextRunAt time.Time) error
```

## InterestRepo

```go
type InterestRepo struct { db *gorm.DB }

func NewInterestRepo(db *gorm.DB) *InterestRepo

func (r *InterestRepo) AccrueInterest(childID, parentID int64) (*models.Transaction, int64, error)
func (r *InterestRepo) UpdateLastInterestAt(childID int64, t time.Time) error
```

## RefreshTokenRepo

```go
type RefreshTokenRepo struct { db *gorm.DB }

func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo

func (r *RefreshTokenRepo) Create(tokenHash, userType string, userID, familyID int64, expiresAt time.Time) error
func (r *RefreshTokenRepo) GetByHash(tokenHash string) (*models.RefreshToken, error)
func (r *RefreshTokenRepo) DeleteByHash(tokenHash string) error
func (r *RefreshTokenRepo) DeleteByUser(userType string, userID int64) error
func (r *RefreshTokenRepo) DeleteExpired() (int64, error)
```

## AuthEventRepo

```go
type AuthEventRepo struct { db *gorm.DB }

func NewAuthEventRepo(db *gorm.DB) *AuthEventRepo

func (r *AuthEventRepo) Log(eventType, userType string, userID, familyID *int64, ipAddress, details string) error
func (r *AuthEventRepo) ListByFamily(familyID int64, limit int) ([]models.AuthEvent, error)
```

## WebhookEventRepo

```go
type WebhookEventRepo struct { db *gorm.DB }

func NewWebhookEventRepo(db *gorm.DB) *WebhookEventRepo

func (r *WebhookEventRepo) Exists(stripeEventID string) (bool, error)
func (r *WebhookEventRepo) Create(stripeEventID, eventType string) error
```

## SavingsGoalRepo

```go
type SavingsGoalRepo struct { db *gorm.DB }

func NewSavingsGoalRepo(db *gorm.DB) *SavingsGoalRepo

func (r *SavingsGoalRepo) Create(childID int64, name string, targetCents int64, emoji *string) (*models.SavingsGoal, error)
func (r *SavingsGoalRepo) GetByID(id int64) (*models.SavingsGoal, error)
func (r *SavingsGoalRepo) ListByChild(childID int64) ([]models.SavingsGoal, error)
func (r *SavingsGoalRepo) Update(id int64, name string, targetCents int64, emoji *string) (*models.SavingsGoal, error)
func (r *SavingsGoalRepo) Delete(id int64) error
func (r *SavingsGoalRepo) Allocate(goalID, childID, amountCents int64) (*models.GoalAllocation, error)
func (r *SavingsGoalRepo) Deallocate(goalID, childID, amountCents int64) (*models.GoalAllocation, error)
```

## GoalAllocationRepo

```go
type GoalAllocationRepo struct { db *gorm.DB }

func NewGoalAllocationRepo(db *gorm.DB) *GoalAllocationRepo

func (r *GoalAllocationRepo) ListByGoal(goalID int64) ([]models.GoalAllocation, error)
func (r *GoalAllocationRepo) ListByChild(childID int64) ([]models.GoalAllocation, error)
```

## Cross-Cutting Patterns

### Transaction Support
Repository methods that need atomicity use GORM transactions internally:
```go
func (r *TransactionRepo) Deposit(...) (*models.Transaction, int64, error) {
    var txn models.Transaction
    var newBalance int64
    err := r.db.Transaction(func(tx *gorm.DB) error {
        // update child balance
        // insert transaction record
        return nil
    })
    return &txn, newBalance, err
}
```

### Not-Found Convention
All `GetByX` methods return `(nil, nil)` when the entity doesn't exist:
```go
func (r *FamilyRepo) GetByID(id int64) (*models.Family, error) {
    var f models.Family
    err := r.db.First(&f, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &f, err
}
```

### DB Connection Setup
```go
// repositories/db.go
func Open(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    // run golang-migrate migrations via underlying *sql.DB
    sqlDB, _ := db.DB()
    runMigrations(sqlDB)
    return db, nil
}
```
