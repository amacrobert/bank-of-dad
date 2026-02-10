package interest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

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

	handler := NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))

	body := `{"interest_rate_bps": 500}`
	req := httptest.NewRequest("PUT", "/api/children/999/interest-rate", bytes.NewBufferString(body))
	req.SetPathValue("id", "999")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSetInterestRate(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// T023: Tests for interest schedule endpoints

func newTestHandler(t *testing.T, db *store.DB) *Handler {
	t.Helper()
	return NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))
}

// PUT /api/children/{childId}/interest-schedule

func TestHandleSetInterestSchedule_CreateWeekly(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	body := `{"frequency":"weekly","day_of_week":5}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp store.InterestSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, store.FrequencyWeekly, resp.Frequency)
	assert.NotNil(t, resp.DayOfWeek)
	assert.Equal(t, 5, *resp.DayOfWeek)
	assert.Equal(t, store.ScheduleStatusActive, resp.Status)
	assert.NotNil(t, resp.NextRunAt)
}

func TestHandleSetInterestSchedule_CreateMonthly(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	body := `{"frequency":"monthly","day_of_month":15}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp store.InterestSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, store.FrequencyMonthly, resp.Frequency)
	assert.NotNil(t, resp.DayOfMonth)
	assert.Equal(t, 15, *resp.DayOfMonth)
}

func TestHandleSetInterestSchedule_UpdateExisting(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create weekly schedule
	body1 := `{"frequency":"weekly","day_of_week":1}`
	req1 := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body1))
	req1.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req1 = setRequestContext(req1, "parent", parent.ID, family.ID)
	rr1 := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Update to monthly
	body2 := `{"frequency":"monthly","day_of_month":20}`
	req2 := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body2))
	req2.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	var resp store.InterestSchedule
	err := json.Unmarshal(rr2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, store.FrequencyMonthly, resp.Frequency)
	assert.NotNil(t, resp.DayOfMonth)
	assert.Equal(t, 20, *resp.DayOfMonth)
}

func TestHandleSetInterestSchedule_ChildForbidden(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	body := `{"frequency":"weekly","day_of_week":5}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleSetInterestSchedule_InvalidFrequency(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	body := `{"frequency":"daily","day_of_week":5}`
	req := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSetInterestSchedule_ChildNotFound(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	h := newTestHandler(t, db)

	body := `{"frequency":"weekly","day_of_week":5}`
	req := httptest.NewRequest("PUT", "/api/children/999/interest-schedule", bytes.NewBufferString(body))
	req.SetPathValue("childId", "999")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// GET /api/children/{childId}/interest-schedule

func TestHandleGetInterestSchedule_Exists(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create schedule first
	body := `{"frequency":"monthly","day_of_month":15}`
	req1 := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req1.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req1 = setRequestContext(req1, "parent", parent.ID, family.ID)
	rr1 := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	// Get schedule
	req2 := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req2.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr2, req2)

	assert.Equal(t, http.StatusOK, rr2.Code)

	var resp store.InterestSchedule
	err := json.Unmarshal(rr2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, store.FrequencyMonthly, resp.Frequency)
}

func TestHandleGetInterestSchedule_None(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "null\n", rr.Body.String())
}

func TestHandleGetInterestSchedule_ChildSeesOwn(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create schedule
	body := `{"frequency":"weekly","day_of_week":5}`
	req1 := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req1.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req1 = setRequestContext(req1, "parent", parent.ID, family.ID)
	rr1 := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	// Child sees own schedule
	req2 := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req2.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req2 = setRequestContext(req2, "child", child.ID, family.ID)
	rr2 := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr2, req2)

	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestHandleGetInterestSchedule_ChildForbiddenOther(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child1 := createTestChild(t, db, family.ID, "Emma")
	child2 := createTestChild(t, db, family.ID, "Jake")

	h := newTestHandler(t, db)

	// Child2 tries to view Child1's schedule
	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child1.ID))
	req = setRequestContext(req, "child", child2.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// DELETE /api/children/{childId}/interest-schedule

func TestHandleDeleteInterestSchedule_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create schedule first
	body := `{"frequency":"weekly","day_of_week":5}`
	req1 := httptest.NewRequest("PUT", "/api/children/1/interest-schedule", bytes.NewBufferString(body))
	req1.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req1 = setRequestContext(req1, "parent", parent.ID, family.ID)
	rr1 := httptest.NewRecorder()
	h.HandleSetInterestSchedule(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	// Delete schedule
	req2 := httptest.NewRequest("DELETE", "/api/children/1/interest-schedule", nil)
	req2.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	h.HandleDeleteInterestSchedule(rr2, req2)

	assert.Equal(t, http.StatusNoContent, rr2.Code)

	// Verify it's gone
	req3 := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req3.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req3 = setRequestContext(req3, "parent", parent.ID, family.ID)
	rr3 := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
	assert.Equal(t, "null\n", rr3.Body.String())
}

func TestHandleDeleteInterestSchedule_NotFound(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	req := httptest.NewRequest("DELETE", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleDeleteInterestSchedule(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleDeleteInterestSchedule_ChildForbidden(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	req := httptest.NewRequest("DELETE", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "child", child.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleDeleteInterestSchedule(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
