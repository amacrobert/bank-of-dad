package allowance

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
	tmp, err := os.CreateTemp("", "bank-of-dad-allowance-test-*.db")
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
	c, err := cs.Create(familyID, name, "password123", nil)
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

// =====================================================
// T010: Tests for POST /api/schedules (create weekly schedule)
// =====================================================

func TestHandleCreateSchedule_Success_Weekly(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"weekly","day_of_week":5,"note":"Weekly allowance"}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), sched.AmountCents)
	assert.Equal(t, store.FrequencyWeekly, sched.Frequency)
	assert.Equal(t, 5, *sched.DayOfWeek)
	assert.Equal(t, "Weekly allowance", *sched.Note)
	assert.Equal(t, store.ScheduleStatusActive, sched.Status)
	assert.NotNil(t, sched.NextRunAt)
}

func TestHandleCreateSchedule_InvalidFrequency(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"daily","day_of_week":5}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleCreateSchedule_MissingDayOfWeek(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"weekly"}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleCreateSchedule_ChildNotInFamily(t *testing.T) {
	db := setupTestDB(t)
	family1 := createTestFamily(t, db)
	parent := createTestParent(t, db, family1.ID)

	fs := store.NewFamilyStore(db)
	family2, _ := fs.Create("other-family")
	child := createTestChild(t, db, family2.ID, "Other")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"weekly","day_of_week":5}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family1.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleCreateSchedule_ChildRoleForbidden(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"weekly","day_of_week":5}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleCreateSchedule_InvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	// Zero amount
	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":0,"frequency":"weekly","day_of_week":5}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Exceeds max
	body2 := fmt.Sprintf(`{"child_id":%d,"amount_cents":100000000,"frequency":"weekly","day_of_week":5}`, child.ID)
	req2 := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body2))
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr2, req2)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}

// =====================================================
// T018: Tests for US2 - Schedule management handlers
// =====================================================

func createScheduleViaHandler(t *testing.T, db *store.DB, parentID, familyID, childID int64, frequency string, dayOfWeek *int, dayOfMonth *int) store.AllowanceSchedule {
	t.Helper()
	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	var body string
	if dayOfWeek != nil {
		body = fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"%s","day_of_week":%d}`, childID, frequency, *dayOfWeek)
	} else if dayOfMonth != nil {
		body = fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"%s","day_of_month":%d}`, childID, frequency, *dayOfMonth)
	}
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parentID, familyID)
	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	return sched
}

func TestHandleListSchedules_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", "/api/schedules", nil)
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleListSchedules(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ScheduleListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Schedules, 1)
	assert.Equal(t, "Emma", resp.Schedules[0].ChildFirstName)
	assert.Equal(t, int64(1000), resp.Schedules[0].AmountCents)
}

func TestHandleListSchedules_EmptyList(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	req := httptest.NewRequest("GET", "/api/schedules", nil)
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleListSchedules(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp ScheduleListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Schedules, 0)
}

func TestHandleGetSchedule_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/schedules/%d", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, created.ID, sched.ID)
}

func TestHandleGetSchedule_WrongFamily(t *testing.T) {
	db := setupTestDB(t)
	family1 := createTestFamily(t, db)
	parent1 := createTestParent(t, db, family1.ID)
	child1 := createTestChild(t, db, family1.ID, "Emma")

	fs := store.NewFamilyStore(db)
	family2, _ := fs.Create("other-family")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent1.ID, family1.ID, child1.ID, "weekly", &dow, nil)

	// Try to access from family2
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/schedules/%d", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent1.ID, family2.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetSchedule(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleUpdateSchedule_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	body := `{"amount_cents":2000,"note":"Updated"}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/schedules/%d", created.ID), bytes.NewBufferString(body))
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleUpdateSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var updated store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &updated)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.AmountCents)
	assert.Equal(t, "Updated", *updated.Note)
}

func TestHandleDeleteSchedule_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/schedules/%d", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleDeleteSchedule(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify it's gone
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/schedules/%d", created.ID), nil)
	getReq.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	getReq = setRequestContext(getReq, "parent", parent.ID, family.ID)
	getRR := httptest.NewRecorder()
	handler.HandleGetSchedule(getRR, getReq)
	assert.Equal(t, http.StatusNotFound, getRR.Code)
}

func TestHandlePauseSchedule_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/schedules/%d/pause", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandlePauseSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var paused store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &paused)
	require.NoError(t, err)
	assert.Equal(t, store.ScheduleStatusPaused, paused.Status)
}

func TestHandlePauseSchedule_AlreadyPaused(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	// Pause first
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/schedules/%d/pause", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandlePauseSchedule(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// Try to pause again
	req2 := httptest.NewRequest("POST", fmt.Sprintf("/api/schedules/%d/pause", created.ID), nil)
	req2.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	handler.HandlePauseSchedule(rr2, req2)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}

func TestHandleResumeSchedule_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	// Pause first
	pauseReq := httptest.NewRequest("POST", fmt.Sprintf("/api/schedules/%d/pause", created.ID), nil)
	pauseReq.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	pauseReq = setRequestContext(pauseReq, "parent", parent.ID, family.ID)
	pauseRR := httptest.NewRecorder()
	handler.HandlePauseSchedule(pauseRR, pauseReq)
	require.Equal(t, http.StatusOK, pauseRR.Code)

	// Resume
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/schedules/%d/resume", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleResumeSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resumed store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &resumed)
	require.NoError(t, err)
	assert.Equal(t, store.ScheduleStatusActive, resumed.Status)
	assert.NotNil(t, resumed.NextRunAt)
}

func TestHandleResumeSchedule_AlreadyActive(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	created := createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/schedules/%d/resume", created.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", created.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleResumeSchedule(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// =====================================================
// T025: Tests for US3 - Biweekly and monthly creation
// =====================================================

func TestHandleCreateSchedule_Success_Biweekly(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":2000,"frequency":"biweekly","day_of_week":1}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, store.FrequencyBiweekly, sched.Frequency)
	assert.Equal(t, 1, *sched.DayOfWeek)
	assert.NotNil(t, sched.NextRunAt)
}

func TestHandleCreateSchedule_Success_Monthly(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":5000,"frequency":"monthly","day_of_month":15}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, store.FrequencyMonthly, sched.Frequency)
	assert.Equal(t, 15, *sched.DayOfMonth)
	assert.NotNil(t, sched.NextRunAt)
}

func TestHandleCreateSchedule_MonthlyMissingDayOfMonth(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"monthly"}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleCreateSchedule_BiweeklyMissingDayOfWeek(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := fmt.Sprintf(`{"child_id":%d,"amount_cents":1000,"frequency":"biweekly"}`, child.ID)
	req := httptest.NewRequest("POST", "/api/schedules", bytes.NewBufferString(body))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleCreateSchedule(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// =====================================================
// T031: Tests for US4 - Upcoming allowances endpoint
// =====================================================

func TestHandleGetUpcomingAllowances_ParentSuccess(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/upcoming-allowances", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetUpcomingAllowances(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp UpcomingAllowancesResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Allowances, 1)
	assert.Equal(t, int64(1000), resp.Allowances[0].AmountCents)
}

func TestHandleGetUpcomingAllowances_ChildSeesOwn(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/upcoming-allowances", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "child", child.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetUpcomingAllowances(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp UpcomingAllowancesResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Allowances, 1)
}

func TestHandleGetUpcomingAllowances_ChildCannotSeeOther(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child1 := createTestChild(t, db, family.ID, "Emma")
	child2 := createTestChild(t, db, family.ID, "Jack")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child1.ID, "weekly", &dow, nil)

	// child2 tries to see child1's allowances
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/upcoming-allowances", child1.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child1.ID))
	req = setRequestContext(req, "child", child2.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetUpcomingAllowances(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleGetUpcomingAllowances_NoSchedules(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/upcoming-allowances", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetUpcomingAllowances(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp UpcomingAllowancesResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Allowances, 0)
}

// =====================================================
// T015 (006): Child-scoped allowance endpoint tests
// =====================================================

func TestHandleGetChildAllowance_Exists(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	// Create allowance first
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/allowance", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetChildAllowance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), sched.AmountCents)
	assert.Equal(t, store.FrequencyWeekly, sched.Frequency)
}

func TestHandleGetChildAllowance_None(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/allowance", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetChildAllowance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "null\n", rr.Body.String())
}

func TestHandleGetChildAllowance_ChildSeesOwn(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/allowance", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "child", child.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetChildAllowance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleGetChildAllowance_ChildCannotSeeOther(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child1 := createTestChild(t, db, family.ID, "Emma")
	child2 := createTestChild(t, db, family.ID, "Jack")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child1.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/allowance", child1.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child1.ID))
	req = setRequestContext(req, "child", child2.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleGetChildAllowance(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleSetChildAllowance_Create(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := `{"amount_cents":1500,"frequency":"weekly","day_of_week":3,"note":"Wednesday allowance"}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/children/%d/allowance", child.ID), bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleSetChildAllowance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, int64(1500), sched.AmountCents)
	assert.Equal(t, store.FrequencyWeekly, sched.Frequency)
	assert.Equal(t, 3, *sched.DayOfWeek)
	assert.Equal(t, "Wednesday allowance", *sched.Note)
	assert.Equal(t, store.ScheduleStatusActive, sched.Status)
	assert.NotNil(t, sched.NextRunAt)
}

func TestHandleSetChildAllowance_Update(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	// Create first
	body1 := `{"amount_cents":1000,"frequency":"weekly","day_of_week":5}`
	req1 := httptest.NewRequest("PUT", fmt.Sprintf("/api/children/%d/allowance", child.ID), bytes.NewBufferString(body1))
	req1.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req1 = setRequestContext(req1, "parent", parent.ID, family.ID)
	rr1 := httptest.NewRecorder()
	handler.HandleSetChildAllowance(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	// Update
	body2 := `{"amount_cents":2000,"frequency":"monthly","day_of_month":15,"note":"Monthly update"}`
	req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/children/%d/allowance", child.ID), bytes.NewBufferString(body2))
	req2.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	handler.HandleSetChildAllowance(rr2, req2)

	assert.Equal(t, http.StatusOK, rr2.Code)
	var updated store.AllowanceSchedule
	err := json.Unmarshal(rr2.Body.Bytes(), &updated)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.AmountCents)
	assert.Equal(t, store.FrequencyMonthly, updated.Frequency)
	assert.Equal(t, 15, *updated.DayOfMonth)
}

func TestHandleSetChildAllowance_ChildForbidden(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := `{"amount_cents":1000,"frequency":"weekly","day_of_week":5}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/children/%d/allowance", child.ID), bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "child", child.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleSetChildAllowance(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleSetChildAllowance_InvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	body := `{"amount_cents":0,"frequency":"weekly","day_of_week":5}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/children/%d/allowance", child.ID), bytes.NewBufferString(body))
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleSetChildAllowance(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleDeleteChildAllowance_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/children/%d/allowance", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleDeleteChildAllowance(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify it's gone
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/children/%d/allowance", child.ID), nil)
	getReq.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	getReq = setRequestContext(getReq, "parent", parent.ID, family.ID)
	getRR := httptest.NewRecorder()
	handler.HandleGetChildAllowance(getRR, getReq)
	assert.Equal(t, http.StatusOK, getRR.Code)
	assert.Equal(t, "null\n", getRR.Body.String())
}

func TestHandleDeleteChildAllowance_NotFound(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/children/%d/allowance", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleDeleteChildAllowance(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandlePauseChildAllowance_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/children/%d/allowance/pause", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandlePauseChildAllowance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, store.ScheduleStatusPaused, sched.Status)
}

func TestHandlePauseChildAllowance_NotFound(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/children/%d/allowance/pause", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandlePauseChildAllowance(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleResumeChildAllowance_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(store.NewScheduleStore(db), store.NewChildStore(db))
	dow := 5
	createScheduleViaHandler(t, db, parent.ID, family.ID, child.ID, "weekly", &dow, nil)

	// Pause first
	pauseReq := httptest.NewRequest("POST", fmt.Sprintf("/api/children/%d/allowance/pause", child.ID), nil)
	pauseReq.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	pauseReq = setRequestContext(pauseReq, "parent", parent.ID, family.ID)
	pauseRR := httptest.NewRecorder()
	handler.HandlePauseChildAllowance(pauseRR, pauseReq)
	require.Equal(t, http.StatusOK, pauseRR.Code)

	// Resume
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/children/%d/allowance/resume", child.ID), nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleResumeChildAllowance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var sched store.AllowanceSchedule
	err := json.Unmarshal(rr.Body.Bytes(), &sched)
	require.NoError(t, err)
	assert.Equal(t, store.ScheduleStatusActive, sched.Status)
	assert.NotNil(t, sched.NextRunAt)
}
