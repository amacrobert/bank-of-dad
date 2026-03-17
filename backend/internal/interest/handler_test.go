package interest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bank-of-dad/internal/testutil"
	"bank-of-dad/models"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	return NewHandler(repositories.NewInterestRepo(db), repositories.NewChildRepo(db), repositories.NewInterestScheduleRepo(db), repositories.NewFamilyRepo(db))
}

func intPtr(i int) *int {
	return &i
}

// GET /api/children/{childId}/interest-schedule

func TestHandleGetInterestSchedule_Exists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create schedule directly via repo
	iss := repositories.NewInterestScheduleRepo(db)
	_, err := iss.Create(&models.InterestSchedule{
		ChildID:    child.ID,
		ParentID:   parent.ID,
		Frequency:  models.FrequencyMonthly,
		DayOfMonth: intPtr(15),
		Status:     models.ScheduleStatusActive,
	})
	require.NoError(t, err)

	// Get schedule
	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.InterestSchedule
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, models.FrequencyMonthly, resp.Frequency)
}

func TestHandleGetInterestSchedule_None(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "null\n", rr.Body.String())
}

func TestHandleGetInterestSchedule_ChildSeesOwn(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	h := newTestHandler(t, db)

	// Create schedule directly via repo
	iss := repositories.NewInterestScheduleRepo(db)
	_, err := iss.Create(&models.InterestSchedule{
		ChildID:   child.ID,
		ParentID:  parent.ID,
		Frequency: models.FrequencyWeekly,
		DayOfWeek: intPtr(5),
		Status:    models.ScheduleStatusActive,
	})
	require.NoError(t, err)

	// Child sees own schedule
	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child.ID))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleGetInterestSchedule_ChildForbiddenOther(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child1 := testutil.CreateTestChild(t, db, family.ID, "Emma")
	child2 := testutil.CreateTestChild(t, db, family.ID, "Jake")

	h := newTestHandler(t, db)

	// Child2 tries to view Child1's schedule
	req := httptest.NewRequest("GET", "/api/children/1/interest-schedule", nil)
	req.SetPathValue("childId", fmt.Sprintf("%d", child1.ID))
	req = testutil.SetRequestContext(req, "child", child2.ID, family.ID)
	rr := httptest.NewRecorder()
	h.HandleGetInterestSchedule(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
