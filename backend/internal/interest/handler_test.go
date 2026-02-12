package interest

import (
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

func newTestHandler(t *testing.T, db *store.DB) *Handler {
	t.Helper()
	return NewHandler(store.NewInterestStore(db), store.NewChildStore(db), store.NewInterestScheduleStore(db))
}

func intPtr(i int) *int {
	return &i
}

// GET /api/children/{childId}/interest-schedule

func TestHandleGetInterestSchedule_Exists(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create schedule directly via store
	iss := store.NewInterestScheduleStore(db)
	_, err := iss.Create(&store.InterestSchedule{
		ChildID:    child.ID,
		ParentID:   parent.ID,
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(15),
		Status:     store.ScheduleStatusActive,
	})
	require.NoError(t, err)

	// Get schedule
	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp store.InterestSchedule
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
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

	// Create schedule directly via store
	iss := store.NewInterestScheduleStore(db)
	_, err := iss.Create(&store.InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(5),
		Status:    store.ScheduleStatusActive,
	})
	require.NoError(t, err)

	// Child sees own schedule
	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = setRequestContext(req, "child", child.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
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
