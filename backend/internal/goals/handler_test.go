package goals

import (
	"bytes"
	"database/sql"
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

// =====================================================
// T018: Tests for POST /api/children/{id}/savings-goals/{goalId}/allocate (HandleAllocate)
//       and GET /api/children/{id}/savings-goals/{goalId}/allocations (HandleListAllocations)
// =====================================================

// setupHandlerWithBalance creates the standard test fixtures plus gives the child
// a $100 balance so allocation tests can move funds into goals.
func setupHandlerWithBalance(t *testing.T) (*Handler, *store.SavingsGoalStore, *store.ChildStore, *store.Family, *store.Parent, *store.Child, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Emma")

	// Give child $100
	txStore := store.NewTransactionStore(db)
	_, _, err := txStore.Deposit(child.ID, parent.ID, 10000, "initial balance")
	require.NoError(t, err)

	goalStore := store.NewSavingsGoalStore(db)
	childStore := store.NewChildStore(db)
	handler := NewHandler(goalStore, childStore)

	return handler, goalStore, childStore, family, parent, child, db
}

// --- HandleAllocate tests ---

func TestHandleAllocate_200_PositiveAmount(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	// Create a goal for the child
	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	body := `{"amount_cents": 2000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AllocateResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, int64(2000), resp.Goal.SavedCents)
	assert.Equal(t, int64(8000), resp.AvailableBalanceCents)
	assert.False(t, resp.Completed)
}

func TestHandleAllocate_200_NegativeAmount_DeAllocation(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	// Create a goal and allocate $30 to it directly via the store
	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	// Allocate $30 first via the handler so saved_cents = 3000
	body := `{"amount_cents": 3000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// Now de-allocate $10
	body = `{"amount_cents": -1000}`
	req = httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr = httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AllocateResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, int64(2000), resp.Goal.SavedCents)
	assert.False(t, resp.Completed)
}

func TestHandleAllocate_400_AmountIsZero(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	body := `{"amount_cents": 0}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleAllocate_400_ExceedsAvailableBalance(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	// Child has $100 balance
	goal, err := goalStore.Create(child.ID, "Expensive Thing", 50000, nil, nil)
	require.NoError(t, err)

	// Try to allocate $120, which exceeds the $100 available balance
	body := `{"amount_cents": 12000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleAllocate_400_DeAllocationExceedsSavedCents(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	// Allocate $20 first
	body := `{"amount_cents": 2000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// Try to de-allocate $30, which exceeds the $20 saved
	body = `{"amount_cents": -3000}`
	req = httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr = httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestHandleAllocate_403_ParentCannotAllocate(t *testing.T) {
	handler, goalStore, _, family, parent, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	body := `{"amount_cents": 2000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	// Parent tries to allocate
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
}

func TestHandleAllocate_404_GoalNotFound(t *testing.T) {
	handler, _, _, family, _, child, _ := setupHandlerWithBalance(t)

	body := `{"amount_cents": 2000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/99999/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", "99999")
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

// --- HandleListAllocations tests ---

func TestHandleListAllocations_200_ReturnsAllocations(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	// Allocate twice via the handler
	for _, amount := range []string{`{"amount_cents": 1000}`, `{"amount_cents": 1500}`} {
		req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(amount))
		req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
		req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
		req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

		rr := httptest.NewRecorder()
		handler.HandleAllocate(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	}

	// List allocations
	req := httptest.NewRequest("GET", "/api/children/1/savings-goals/1/allocations", nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleListAllocations(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AllocationsListResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Allocations, 2)
	// Verify allocation fields are populated
	assert.Equal(t, goal.ID, resp.Allocations[0].GoalID)
	assert.NotZero(t, resp.Allocations[0].AmountCents)
	assert.NotZero(t, resp.Allocations[0].CreatedAt)
}

func TestHandleListAllocations_200_ParentCanView(t *testing.T) {
	handler, goalStore, _, family, parent, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	// Allocate once as the child
	body := `{"amount_cents": 2000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/1/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// Parent lists allocations for the child's goal
	req = httptest.NewRequest("GET", "/api/children/1/savings-goals/1/allocations", nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr = httptest.NewRecorder()
	handler.HandleListAllocations(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AllocationsListResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Allocations, 1)
	assert.Equal(t, int64(2000), resp.Allocations[0].AmountCents)
}

// --- T030: Allocate completion response tests ---

func TestHandleAllocate_200_CompletesGoal(t *testing.T) {
	handler, goalStore, _, _, _, child, _ := setupHandlerWithBalance(t)

	// Create a goal with target $50
	goal, err := goalStore.Create(child.ID, "Small Goal", 5000, nil, nil)
	require.NoError(t, err)

	// Allocate exactly the target amount
	body := `{"amount_cents": 5000}`
	req := httptest.NewRequest("POST", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10)+"/allocate", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, child.FamilyID)

	rr := httptest.NewRecorder()
	handler.HandleAllocate(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AllocateResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.True(t, resp.Completed)
	assert.Equal(t, "completed", resp.Goal.Status)
	assert.NotNil(t, resp.Goal.CompletedAt)
}

// =====================================================
// Tests for PUT /api/children/{id}/savings-goals/{goalId} (HandleUpdate)
// =====================================================

// DeleteResponse mirrors the struct that will be defined in handler.go for
// DELETE /api/children/{id}/savings-goals/{goalId}.
type DeleteResponse struct {
	ReleasedCents         int64 `json:"released_cents"`
	AvailableBalanceCents int64 `json:"available_balance_cents"`
}

func TestHandleUpdate_200_Success(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	// Create a goal for the child
	goal, err := goalStore.Create(child.ID, "Old Name", 5000, nil, nil)
	require.NoError(t, err)

	body := `{"name": "New Name"}`
	req := httptest.NewRequest("PUT", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdate(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var updated store.SavingsGoal
	err = json.NewDecoder(rr.Body).Decode(&updated)
	require.NoError(t, err)

	assert.Equal(t, goal.ID, updated.ID)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, int64(5000), updated.TargetCents)
	assert.Equal(t, "active", updated.Status)
}

func TestHandleUpdate_400_NameTooLong(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	longName := strings.Repeat("a", 51)
	body := fmt.Sprintf(`{"name": "%s"}`, longName)
	req := httptest.NewRequest("PUT", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_name", resp.Error)
}

func TestHandleUpdate_400_TargetCentsNegative(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	body := `{"target_cents": -100}`
	req := httptest.NewRequest("PUT", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdate(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_target", resp.Error)
}

func TestHandleUpdate_403_ParentCannotUpdate(t *testing.T) {
	handler, goalStore, _, family, parent, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	body := `{"name": "New Name"}`
	req := httptest.NewRequest("PUT", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	// Parent tries to update
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdate(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
}

func TestHandleUpdate_404_NotFound(t *testing.T) {
	handler, _, _, family, _, child, _ := setupHandlerWithBalance(t)

	body := `{"name": "New Name"}`
	req := httptest.NewRequest("PUT", "/api/children/1/savings-goals/99999", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", "99999")
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdate(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// =====================================================
// Tests for DELETE /api/children/{id}/savings-goals/{goalId} (HandleDelete)
// =====================================================

func TestHandleDelete_200_Success(t *testing.T) {
	handler, goalStore, _, family, _, child, _ := setupHandlerWithBalance(t)

	// Create a goal and allocate $20 to it so it has saved_cents
	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	// Allocate $20 via the handler
	allocBody := `{"amount_cents": 2000}`
	allocReq := httptest.NewRequest("POST", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10)+"/allocate", bytes.NewBufferString(allocBody))
	allocReq.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	allocReq.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	allocReq = testutil.SetRequestContext(allocReq, "child", child.ID, family.ID)

	allocRR := httptest.NewRecorder()
	handler.HandleAllocate(allocRR, allocReq)
	require.Equal(t, http.StatusOK, allocRR.Code)

	// Now delete the goal
	req := httptest.NewRequest("DELETE", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDelete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp DeleteResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, int64(2000), resp.ReleasedCents)
	// Child started with $100 (10000 cents), allocated $20 (2000 cents), now released back
	// Available balance should be back to $100
	assert.Equal(t, int64(10000), resp.AvailableBalanceCents)
}

func TestHandleDelete_403_ParentCannotDelete(t *testing.T) {
	handler, goalStore, _, family, parent, child, _ := setupHandlerWithBalance(t)

	goal, err := goalStore.Create(child.ID, "Skateboard", 5000, nil, nil)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/api/children/1/savings-goals/"+strconv.FormatInt(goal.ID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", strconv.FormatInt(goal.ID, 10))
	// Parent tries to delete
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDelete(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp ErrorResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
}

func TestHandleDelete_404_NotFoundOrCompleted(t *testing.T) {
	handler, _, _, family, _, child, _ := setupHandlerWithBalance(t)

	req := httptest.NewRequest("DELETE", "/api/children/1/savings-goals/99999", nil)
	req.SetPathValue("id", strconv.FormatInt(child.ID, 10))
	req.SetPathValue("goalId", "99999")
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDelete(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
