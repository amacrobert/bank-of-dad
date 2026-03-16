# Quickstart: GORM Backend Refactor

**Date**: 2026-03-16 | **Feature**: 029-gorm-backend-refactor

## Prerequisites

- Go 1.24+
- PostgreSQL 17 running locally
- Test database `bankofdad_test` exists

## New Dependencies

```bash
cd backend
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

## Key Files to Create

### 1. Models (`backend/models/`)

One file per entity. Example:

```go
// backend/models/family.go
package models

import "time"

type Family struct {
    ID                            int64      `gorm:"primaryKey"`
    Slug                          string     `gorm:"uniqueIndex;not null"`
    Timezone                      string     `gorm:"not null;default:America/New_York"`
    AccountType                   string     `gorm:"not null;default:free"`
    StripeCustomerID              *string    `gorm:"uniqueIndex"`
    StripeSubscriptionID          *string    `gorm:"uniqueIndex"`
    SubscriptionStatus            *string
    SubscriptionCurrentPeriodEnd  *time.Time
    SubscriptionCancelAtPeriodEnd bool       `gorm:"not null;default:false"`
    BankName                      string     `gorm:"not null;default:Dad"`
    CreatedAt                     time.Time  `gorm:"autoCreateTime"`
}
```

### 2. Repositories (`backend/repositories/`)

One file per entity. Example:

```go
// backend/repositories/family_repo.go
package repositories

import (
    "errors"
    "gorm.io/gorm"
    "bank-of-dad/models"
)

type FamilyRepo struct {
    db *gorm.DB
}

func NewFamilyRepo(db *gorm.DB) *FamilyRepo {
    return &FamilyRepo{db: db}
}

func (r *FamilyRepo) GetByID(id int64) (*models.Family, error) {
    var f models.Family
    err := r.db.First(&f, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &f, err
}
```

### 3. DB Connection (`backend/repositories/db.go`)

```go
package repositories

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    runMigrations(sqlDB) // existing golang-migrate logic
    return db, nil
}
```

## Running Tests

```bash
cd backend && go test -p 1 ./...
```

## Migration Order

1. Create all models in `backend/models/`
2. Create `repositories/db.go` (GORM connection + migrations)
3. Create repositories one at a time with tests
4. Update handlers one package at a time to use repositories
5. Remove `internal/store/` once all consumers are migrated

## Important Constraints

- **DO NOT** use `gorm.AutoMigrate` — migrations stay in `golang-migrate`
- **DO NOT** change API request/response shapes — this is a pure internal refactor
- **DO** use `db.Raw()` for complex SQL that GORM builder can't express cleanly
- **DO** preserve int64 cents for all money fields
- **DO** return `(nil, nil)` for not-found in `GetByX` methods
