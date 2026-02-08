package interest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmp, err := os.CreateTemp("", "bank-of-dad-interest-test-*.db")
	require.NoError(t, err)
	path := tmp.Name()
	tmp.Close()
	t.Cleanup(func() {
		os.Remove(path)
		os.Remove(path + "-wal")
		os.Remove(path + "-shm")
	})
	db, err := store.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func createTestFamily(t *testing.T, db *store.DB) *store.Family {
	t.Helper()
	fs := store.NewFamilyStore(db)
	f, err := fs.Create("test-family")
	require.NoError(t, err)
	return f
}

func createTestParent(t *testing.T, db *store.DB, familyID int64) *store.Parent {
	t.Helper()
	ps := store.NewParentStore(db)
	p, err := ps.Create("google-id-123", "parent@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(p.ID, familyID)
	require.NoError(t, err)
	p.FamilyID = familyID
	return p
}

func createTestChild(t *testing.T, db *store.DB, familyID int64, name string) *store.Child {
	t.Helper()
	cs := store.NewChildStore(db)
	c, err := cs.Create(familyID, name, "password123")
	require.NoError(t, err)
	return c
}

func setRequestContext(r *http.Request, userType string, userID, familyID int64) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyUserType, userType)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, middleware.ContextKeyFamilyID, familyID)
	return r.WithContext(ctx)
}

// T011: Tests for PUT /api/children/{id}/interest-rate

func TestHandleSetInterestRate_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	body := `{"interest_rate_bps": 500}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp InterestRateResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, 500, resp.InterestRateBps)
	assert.Equal(t, "5.00%", resp.InterestRateDisplay)
}

func TestHandleSetInterestRate_SetToZero(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	// First set a rate
	body := `{"interest_rate_bps": 500}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Now disable it
	body2 := `{"interest_rate_bps": 0}`
	req2 := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body2))
	req2.SetPathValue("id", "1")
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr2, req2)

	assert.Equal(t, http.StatusOK, rr2.Code)

	var resp InterestRateResponse
	err := json.Unmarshal(rr2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.InterestRateBps)
	assert.Equal(t, "0.00%", resp.InterestRateDisplay)
}

func TestHandleSetInterestRate_ValidationError_Negative(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	body := `{"interest_rate_bps": -1}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSetInterestRate_ValidationError_TooHigh(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	body := `{"interest_rate_bps": 10001}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSetInterestRate_Forbidden_WrongFamily(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	body := `{"interest_rate_bps": 500}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, 999) // Wrong family

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleSetInterestRate_Forbidden_ChildCannot(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	body := `{"interest_rate_bps": 500}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleSetInterestRate_NotFound(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db))

	body := `{"interest_rate_bps": 500}`
	req := httptest.NewRequest("PUT", "/api/children/999/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "999")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
