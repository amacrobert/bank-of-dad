package goals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"bank-of-dad/internal/store"
	"bank-of-dad/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a goals handler with the standard stores.
func setupHandler(t *testing.T) (*Handler, *store.SavingsGoalStore, *store.ChildStore, *store.Family, *store.Parent, *store.Child) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	goalStore := store.NewSavingsGoalStore(db)
	childStore := store.NewChildStore(db)
	handler := NewHandler(goalStore, childStore)

	return handler, goalStore, childStore, family, parent, child
}

// =====================================================
// T009: Tests for POST /api/children/{id}/savings-goals (HandleCreate)
// =====================================================

func TestHandleCreate_Success(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	body := `{"name": "New Skateboard", "target_cents": 4500}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var goal store.SavingsGoal
	err := json.NewDecoder(rr.Body).Decode(&goal)
	require.NoError(t, err)

	assert.Equal(t, child.ID, goal.ChildID)
	assert.Equal(t, "New Skateboard", goal.Name)
	assert.Equal(t, int64(4500), goal.TargetCents)
	assert.Equal(t, int64(0), goal.SavedCents)
	assert.Equal(t, "active", goal.Status)
	assert.Nil(t, goal.CompletedAt)
	assert.NotZero(t, goal.ID)
	assert.NotZero(t, goal.CreatedAt)
	assert.NotZero(t, goal.UpdatedAt)
}

func TestHandleCreate_SuccessWithOptionalFields(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	body := `{"name": "Video Game", "target_cents": 6000, "emoji": "🎮", "target_date": "2026-12-25"}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var goal store.SavingsGoal
	err := json.NewDecoder(rr.Body).Decode(&goal)
	require.NoError(t, err)

	assert.Equal(t, "Video Game", goal.Name)
	assert.Equal(t, int64(6000), goal.TargetCents)
	require.NotNil(t, goal.Emoji)
	assert.Equal(t, "🎮", *goal.Emoji)
	require.NotNil(t, goal.TargetDate)
}

func TestHandleCreate_400_MissingName(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	body := `{"name": "", "target_cents": 4500}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleCreate_400_MissingTargetCents(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	// target_cents is 0 (zero-value when omitted)
	body := `{"name": "Skateboard"}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleCreate_400_NegativeTargetCents(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	body := `{"name": "Skateboard", "target_cents": -100}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleCreate_400_NameTooLong(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	longName := strings.Repeat("a", 51)
	body := fmt.Sprintf(`{"name": "%s", "target_cents": 4500}`, longName)
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleCreate_403_ParentCannotCreate(t *testing.T) {
	handler, _, _, family, parent, child := setupHandler(t)

	body := `{"name": "Skateboard", "target_cents": 4500}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	// Parent tries to create a goal for the child
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
}

func TestHandleCreate_409_MaxActiveGoals(t *testing.T) {
	handler, goalStore, _, family, _, child := setupHandler(t)

	// Create 5 active goals directly via the store
	for i := 1; i <= 5; i++ {
		_, err := goalStore.Create(child.ID, fmt.Sprintf("Goal %d", i), int64(1000*i), nil, nil)
		require.NoError(t, err)
	}

	// Try to create a 6th goal via the handler
	body := `{"name": "Goal 6", "target_cents": 6000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreate(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

// =====================================================
// T009: Tests for GET /api/children/{id}/savings-goals (HandleList)
// =====================================================

func TestHandleList_200_ReturnsGoals(t *testing.T) {
	handler, goalStore, _, family, _, child := setupHandler(t)

	// Create a couple of goals directly via the store
	_, err := goalStore.Create(child.ID, "Skateboard", 4500, nil, nil)
	require.NoError(t, err)
	_, err = goalStore.Create(child.ID, "Video Game", 6000, nil, nil)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/children/1/savings-goals", nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleList(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp SavingsGoalsResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Goals, 2)
	// Verify goal fields are populated
	assert.Equal(t, "active", resp.Goals[0].Status)
	assert.Equal(t, child.ID, resp.Goals[0].ChildID)
}

func TestHandleList_200_EmptyList(t *testing.T) {
	handler, _, _, family, _, child := setupHandler(t)

	req := httptest.NewRequest("GET", "/api/children/1/savings-goals", nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleList(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp SavingsGoalsResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	// Should be empty array, not null
	assert.NotNil(t, resp.Goals)
	assert.Len(t, resp.Goals, 0)
}

func TestHandleList_403_ChildCannotSeeSiblingGoals(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	childA := testutil.CreateTestChild(t, db, family.ID, "Emma")
	childB := testutil.CreateTestChild(t, db, family.ID, "Jack")

	goalStore := store.NewSavingsGoalStore(db)
	childStore := store.NewChildStore(db)
	handler := NewHandler(goalStore, childStore)

	// Create a goal for child B
	_, err := goalStore.Create(childB.ID, "Skateboard", 4500, nil, nil)
	require.NoError(t, err)

	// Child A tries to view child B's goals
	req := httptest.NewRequest("GET", "/api/children/2/savings-goals", nil)
	req.SetPathValue("id", strconv.FormatInt(childB.ID, 10))
	req = testutil.SetRequestContext(req, "child", childA.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleList(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
}

func TestHandleList_200_ParentCanSeeChildGoals(t *testing.T) {
	handler, goalStore, _, family, parent, child := setupHandler(t)

	// Create goals for the child
	_, err := goalStore.Create(child.ID, "Skateboard", 4500, nil, nil)
	require.NoError(t, err)
	_, err = goalStore.Create(child.ID, "Video Game", 6000, nil, nil)
	require.NoError(t, err)

	// Parent views the child's goals
	req := httptest.NewRequest("GET", "/api/children/1/savings-goals", nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleList(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp SavingsGoalsResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Len(t, resp.Goals, 2)
}
