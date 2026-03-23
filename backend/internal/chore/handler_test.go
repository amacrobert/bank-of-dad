package chore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"bank-of-dad/internal/testutil"
	"bank-of-dad/models"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================
// Tests for POST /api/chores (HandleCreateChore)
// =====================================================

func TestHandleCreateChore_Success_OneTime(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Mow the lawn","description":"Front yard","reward_cents":500,"recurrence":"one_time","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]json.RawMessage
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var choreResp ChoreResponse
	err = json.Unmarshal(resp["chore"], &choreResp)
	require.NoError(t, err)

	assert.Equal(t, "Mow the lawn", choreResp.Name)
	assert.NotNil(t, choreResp.Description)
	assert.Equal(t, "Front yard", *choreResp.Description)
	assert.Equal(t, 500, choreResp.RewardCents)
	assert.Equal(t, "one_time", choreResp.Recurrence)
	assert.True(t, choreResp.IsActive)
	assert.Equal(t, family.ID, choreResp.FamilyID)
	assert.Len(t, choreResp.Assignments, 1)
	assert.Equal(t, child.ID, choreResp.Assignments[0].ChildID)
	assert.Equal(t, "Alice", choreResp.Assignments[0].ChildName)

	// Verify instance was created for one-time chore
	instanceRepo := repositories.NewChoreInstanceRepo(db)
	available, _, _, err := instanceRepo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, available, 1)
	assert.Equal(t, 500, available[0].RewardCents)
}

func TestHandleCreateChore_Success_MultiChild(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child1 := testutil.CreateTestChild(t, db, family.ID, "Alice")
	child2 := testutil.CreateTestChild(t, db, family.ID, "Bob")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Dishes","reward_cents":300,"recurrence":"one_time","child_ids":[%d,%d]}`, child1.ID, child2.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]json.RawMessage
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var choreResp ChoreResponse
	err = json.Unmarshal(resp["chore"], &choreResp)
	require.NoError(t, err)

	assert.Len(t, choreResp.Assignments, 2)

	// Verify instances were created for both children
	instanceRepo := repositories.NewChoreInstanceRepo(db)
	avail1, _, _, err := instanceRepo.ListByChild(child1.ID)
	require.NoError(t, err)
	assert.Len(t, avail1, 1)

	avail2, _, _, err := instanceRepo.ListByChild(child2.ID)
	require.NoError(t, err)
	assert.Len(t, avail2, 1)
}

func TestHandleCreateChore_Success_RecurringWeekly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	dow := 1 // Monday
	body := fmt.Sprintf(`{"name":"Take out trash","reward_cents":200,"recurrence":"weekly","day_of_week":%d,"child_ids":[%d]}`, dow, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]json.RawMessage
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var choreResp ChoreResponse
	err = json.Unmarshal(resp["chore"], &choreResp)
	require.NoError(t, err)

	assert.Equal(t, "weekly", choreResp.Recurrence)
	assert.NotNil(t, choreResp.DayOfWeek)
	assert.Equal(t, 1, *choreResp.DayOfWeek)

	// No instances should be created for recurring chores
	instanceRepo := repositories.NewChoreInstanceRepo(db)
	avail, _, _, err := instanceRepo.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Len(t, avail, 0)
}

func TestHandleCreateChore_Forbidden_ChildUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Dishes","reward_cents":100,"recurrence":"one_time","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

func TestHandleCreateChore_InvalidName_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"","reward_cents":100,"recurrence":"one_time","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_name", errResp.Error)
}

func TestHandleCreateChore_InvalidName_TooLong(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	longName := strings.Repeat("a", 101)
	body := fmt.Sprintf(`{"name":"%s","reward_cents":100,"recurrence":"one_time","child_ids":[%d]}`, longName, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_name", errResp.Error)
}

func TestHandleCreateChore_InvalidAmount_Negative(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Dishes","reward_cents":-1,"recurrence":"one_time","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_amount", errResp.Error)
}

func TestHandleCreateChore_InvalidRecurrence(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Dishes","reward_cents":100,"recurrence":"biweekly","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_recurrence", errResp.Error)
}

func TestHandleCreateChore_MissingChildIDs(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := `{"name":"Dishes","reward_cents":100,"recurrence":"one_time","child_ids":[]}`
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_children", errResp.Error)
}

func TestHandleCreateChore_ChildNotInFamily(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family1 := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family1.ID)

	// Create a child in a different family
	familyRepo := repositories.NewFamilyRepo(db)
	family2, err := familyRepo.Create("other-family")
	require.NoError(t, err)
	otherChild := testutil.CreateTestChild(t, db, family2.ID, "OtherKid")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Dishes","reward_cents":100,"recurrence":"one_time","child_ids":[%d]}`, otherChild.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family1.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

func TestHandleCreateChore_WeeklyWithoutDayOfWeek(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := fmt.Sprintf(`{"name":"Take out trash","reward_cents":200,"recurrence":"weekly","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_recurrence", errResp.Error)
}

// =====================================================
// Tests for GET /api/chores (HandleListChores)
// =====================================================

func TestHandleListChores_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	// Create a chore first
	body := fmt.Sprintf(`{"name":"Mow the lawn","reward_cents":500,"recurrence":"one_time","child_ids":[%d]}`, child.ID)
	req := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleCreateChore(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	// Now list chores
	req = httptest.NewRequest("GET", "/api/chores", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)
	rr = httptest.NewRecorder()
	handler.HandleListChores(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Chores []ChoreResponse `json:"chores"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Chores, 1)
	assert.Equal(t, "Mow the lawn", resp.Chores[0].Name)
	assert.Len(t, resp.Chores[0].Assignments, 1)
	assert.Equal(t, "Alice", resp.Chores[0].Assignments[0].ChildName)
}

func TestHandleListChores_EmptyList(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/chores", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleListChores(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Chores []ChoreResponse `json:"chores"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Chores)
	assert.Len(t, resp.Chores, 0)
}

func TestHandleListChores_Forbidden_ChildUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/chores", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleListChores(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

// =====================================================
// Tests for GET /api/child/chores (HandleChildListChores)
// =====================================================

func TestHandleChildListChores_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	// Create a chore and instances
	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: 1,
		Name:              "Wash dishes",
		RewardCents:       300,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	// Create two instances: one available, one mark as complete (pending)
	inst1 := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 300,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created1, err := instanceRepo.CreateInstance(inst1)
	require.NoError(t, err)

	inst2 := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 300,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created2, err := instanceRepo.CreateInstance(inst2)
	require.NoError(t, err)

	// Mark one complete
	err = instanceRepo.MarkComplete(created2.ID, child.ID)
	require.NoError(t, err)

	_ = created1 // available instance

	req := httptest.NewRequest("GET", "/api/child/chores", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildListChores(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Available []InstanceResponse `json:"available"`
		Pending   []InstanceResponse `json:"pending"`
		Completed []InstanceResponse `json:"completed"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Available, 1)
	assert.Len(t, resp.Pending, 1)
	assert.Len(t, resp.Completed, 0)
	assert.Equal(t, "Wash dishes", resp.Available[0].ChoreName)
	assert.Equal(t, "Wash dishes", resp.Pending[0].ChoreName)
}

func TestHandleChildListChores_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/child/chores", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildListChores(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Available []InstanceResponse `json:"available"`
		Pending   []InstanceResponse `json:"pending"`
		Completed []InstanceResponse `json:"completed"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Available, 0)
	assert.Len(t, resp.Pending, 0)
	assert.Len(t, resp.Completed, 0)
}

func TestHandleChildListChores_ForbiddenForParent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/child/chores", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildListChores(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

// =====================================================
// Tests for POST /api/child/chores/{id}/complete (HandleCompleteChore)
// =====================================================

func TestHandleCompleteChore_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	// Create chore and instance
	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Clean room",
		RewardCents:       500,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 500,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/child/chores/"+strconv.FormatInt(created.ID, 10)+"/complete", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCompleteChore(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instance InstanceResponse `json:"instance"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusPendingApproval, resp.Instance.Status)
	assert.NotNil(t, resp.Instance.CompletedAt)
}

func TestHandleCompleteChore_AlreadyPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Clean room",
		RewardCents:       500,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 500,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	// Mark complete first
	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	// Try to complete again
	req := httptest.NewRequest("POST", "/api/child/chores/"+strconv.FormatInt(created.ID, 10)+"/complete", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCompleteChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_status", errResp.Error)
}

func TestHandleCompleteChore_WrongChild(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child1 := testutil.CreateTestChild(t, db, family.ID, "Alice")
	child2 := testutil.CreateTestChild(t, db, family.ID, "Bob")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Clean room",
		RewardCents:       500,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child1.ID,
		RewardCents: 500,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	// Child2 tries to complete child1's instance
	req := httptest.NewRequest("POST", "/api/child/chores/"+strconv.FormatInt(created.ID, 10)+"/complete", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "child", child2.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCompleteChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_status", errResp.Error)
}

func TestHandleCompleteChore_ForbiddenForParent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("POST", "/api/child/chores/1/complete", nil)
	req.SetPathValue("id", "1")
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCompleteChore(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

// =====================================================
// Tests for GET /api/chores/pending (HandleListPending)
// =====================================================

func TestHandleListPending_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Sweep floor",
		RewardCents:       200,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 200,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	// Mark complete to make it pending
	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/chores/pending", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleListPending(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instances []InstanceResponse `json:"instances"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Instances, 1)
	assert.Equal(t, "Sweep floor", resp.Instances[0].ChoreName)
	assert.Equal(t, "Alice", resp.Instances[0].ChildName)
	assert.Equal(t, 200, resp.Instances[0].RewardCents)
}

func TestHandleListPending_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/chores/pending", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleListPending(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instances []InstanceResponse `json:"instances"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Instances)
	assert.Len(t, resp.Instances, 0)
}

// =====================================================
// Tests for POST /api/chore-instances/{id}/approve (HandleApprove)
// =====================================================

func TestHandleApprove_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)
	txRepo := repositories.NewTransactionRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		txRepo,
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Mop kitchen",
		RewardCents:       400,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 400,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/approve", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instance   InstanceResponse `json:"instance"`
		NewBalance int64            `json:"new_balance"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusApproved, resp.Instance.Status)
	assert.NotNil(t, resp.Instance.TransactionID)
	assert.Equal(t, int64(400), resp.NewBalance)
}

func TestHandleApprove_ZeroReward(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Help sibling",
		RewardCents:       0,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 0,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/approve", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instance   InstanceResponse `json:"instance"`
		NewBalance int64            `json:"new_balance"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusApproved, resp.Instance.Status)
	assert.Nil(t, resp.Instance.TransactionID)
	assert.Equal(t, int64(0), resp.NewBalance)
}

func TestHandleApprove_NotPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Vacuum",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	// Try to approve without marking complete first
	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/approve", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_status", errResp.Error)
}

func TestHandleApprove_WrongFamily(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family1 := testutil.CreateTestFamily(t, db)
	parent1 := testutil.CreateTestParent(t, db, family1.ID)
	child1 := testutil.CreateTestChild(t, db, family1.ID, "Alice")

	// Create a second family with a different parent
	familyRepo := repositories.NewFamilyRepo(db)
	family2, err := familyRepo.Create("other-family")
	require.NoError(t, err)
	parentRepo := repositories.NewParentRepo(db)
	parent2, err := parentRepo.Create("google-id-456", "parent2@test.com", "Other Parent")
	require.NoError(t, err)
	err = parentRepo.SetFamilyID(parent2.ID, family2.ID)
	require.NoError(t, err)
	parent2.FamilyID = family2.ID

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family1.ID,
		CreatedByParentID: parent1.ID,
		Name:              "Vacuum",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child1.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	err = instanceRepo.MarkComplete(created.ID, child1.ID)
	require.NoError(t, err)

	// Parent2 from family2 tries to approve
	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/approve", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent2.ID, family2.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

// =====================================================
// Tests for POST /api/chore-instances/{id}/reject (HandleReject)
// =====================================================

func TestHandleReject_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Laundry",
		RewardCents:       250,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 250,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/reject", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleReject(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instance InstanceResponse `json:"instance"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusAvailable, resp.Instance.Status)
}

func TestHandleReject_WithReason(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Laundry",
		RewardCents:       250,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 250,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	body := `{"reason":"Not done properly"}`
	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/reject", bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleReject(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Instance InstanceResponse `json:"instance"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, models.ChoreInstanceStatusAvailable, resp.Instance.Status)
	assert.NotNil(t, resp.Instance.RejectionReason)
	assert.Equal(t, "Not done properly", *resp.Instance.RejectionReason)
}

func TestHandleReject_NotPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Vacuum",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	// Try to reject without marking complete first
	req := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(created.ID, 10)+"/reject", nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleReject(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_status", errResp.Error)
}

// =====================================================
// Tests for GET /api/child/chores/earnings (HandleChildEarnings)
// =====================================================

func TestHandleChildEarnings_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)
	txRepo := repositories.NewTransactionRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		txRepo,
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Wash dishes",
		RewardCents:       300,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	createdChore, err := choreRepo.Create(chore)
	require.NoError(t, err)

	// Create, complete, and approve an instance
	inst := &models.ChoreInstance{
		ChoreID:     createdChore.ID,
		ChildID:     child.ID,
		RewardCents: 300,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	created, err := instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	err = instanceRepo.MarkComplete(created.ID, child.ID)
	require.NoError(t, err)

	tx, _, err := txRepo.DepositChore(child.ID, parent.ID, 300, "Wash dishes")
	require.NoError(t, err)

	err = instanceRepo.Approve(created.ID, parent.ID, &tx.ID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/child/chores/earnings", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildEarnings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(300), resp["total_earned_cents"])
	assert.Equal(t, float64(1), resp["chores_completed"])
	recent, ok := resp["recent"].([]interface{})
	require.True(t, ok)
	assert.Len(t, recent, 1)
}

func TestHandleChildEarnings_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	child := testutil.CreateTestChild(t, db, family.ID, "Bob")

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/child/chores/earnings", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildEarnings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["total_earned_cents"])
	assert.Equal(t, float64(0), resp["chores_completed"])
}

func TestHandleChildEarnings_ForbiddenForParent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("GET", "/api/child/chores/earnings", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildEarnings(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", errResp.Error)
}

// =====================================================
// Tests for PUT /api/chores/{id} (HandleUpdateChore)
// =====================================================

func TestHandleUpdateChore_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	choreRepo := repositories.NewChoreRepo(db)

	handler := NewHandler(
		choreRepo,
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Old Name",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	created, err := choreRepo.Create(chore)
	require.NoError(t, err)

	body := `{"name":"New Name","reward_cents":500}`
	req := httptest.NewRequest("PUT", "/api/chores/"+strconv.FormatInt(created.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdateChore(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Chore ChoreResponse `json:"chore"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "New Name", resp.Chore.Name)
	assert.Equal(t, 500, resp.Chore.RewardCents)
}

func TestHandleUpdateChore_InvalidName(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	choreRepo := repositories.NewChoreRepo(db)

	handler := NewHandler(
		choreRepo,
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Valid Name",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	created, err := choreRepo.Create(chore)
	require.NoError(t, err)

	body := `{"name":""}`
	req := httptest.NewRequest("PUT", "/api/chores/"+strconv.FormatInt(created.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdateChore(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "validation_error", errResp.Error)
}

func TestHandleUpdateChore_WrongFamily(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family1 := testutil.CreateTestFamily(t, db)
	parent1 := testutil.CreateTestParent(t, db, family1.ID)

	familyRepo := repositories.NewFamilyRepo(db)
	family2, err := familyRepo.Create("other-fam")
	require.NoError(t, err)
	parentRepo := repositories.NewParentRepo(db)
	parent2, err := parentRepo.Create("gid-upd", "p2@test.com", "Parent 2")
	require.NoError(t, err)
	err = parentRepo.SetFamilyID(parent2.ID, family2.ID)
	require.NoError(t, err)

	choreRepo := repositories.NewChoreRepo(db)

	handler := NewHandler(
		choreRepo,
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family1.ID,
		CreatedByParentID: parent1.ID,
		Name:              "Chore",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	created, err := choreRepo.Create(chore)
	require.NoError(t, err)

	body := `{"name":"Hacked"}`
	req := httptest.NewRequest("PUT", "/api/chores/"+strconv.FormatInt(created.ID, 10), bytes.NewBufferString(body))
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent2.ID, family2.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdateChore(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleUpdateChore_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	body := `{"name":"Whatever"}`
	req := httptest.NewRequest("PUT", "/api/chores/99999", bytes.NewBufferString(body))
	req.SetPathValue("id", "99999")
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleUpdateChore(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// =====================================================
// Tests for DELETE /api/chores/{id} (HandleDeleteChore)
// =====================================================

func TestHandleDeleteChore_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)

	handler := NewHandler(
		choreRepo,
		instanceRepo,
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family.ID,
		CreatedByParentID: parent.ID,
		Name:              "Delete Me",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	created, err := choreRepo.Create(chore)
	require.NoError(t, err)

	// Create an available instance
	inst := &models.ChoreInstance{
		ChoreID:     created.ID,
		ChildID:     child.ID,
		RewardCents: 100,
		Status:      models.ChoreInstanceStatusAvailable,
	}
	_, err = instanceRepo.CreateInstance(inst)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/api/chores/"+strconv.FormatInt(created.ID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeleteChore(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify chore is gone
	deleted, err := choreRepo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestHandleDeleteChore_WrongFamily(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family1 := testutil.CreateTestFamily(t, db)
	parent1 := testutil.CreateTestParent(t, db, family1.ID)

	familyRepo := repositories.NewFamilyRepo(db)
	family2, err := familyRepo.Create("other-fam-del")
	require.NoError(t, err)
	parentRepo := repositories.NewParentRepo(db)
	parent2, err := parentRepo.Create("gid-del", "pd@test.com", "Parent D")
	require.NoError(t, err)
	err = parentRepo.SetFamilyID(parent2.ID, family2.ID)
	require.NoError(t, err)

	choreRepo := repositories.NewChoreRepo(db)

	handler := NewHandler(
		choreRepo,
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	chore := &models.Chore{
		FamilyID:          family1.ID,
		CreatedByParentID: parent1.ID,
		Name:              "Protected",
		RewardCents:       100,
		Recurrence:        models.ChoreRecurrenceOneTime,
		IsActive:          true,
	}
	created, err := choreRepo.Create(chore)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/api/chores/"+strconv.FormatInt(created.ID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(created.ID, 10))
	req = testutil.SetRequestContext(req, "parent", parent2.ID, family2.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeleteChore(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleDeleteChore_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	handler := NewHandler(
		repositories.NewChoreRepo(db),
		repositories.NewChoreInstanceRepo(db),
		repositories.NewTransactionRepo(db),
		repositories.NewChildRepo(db),
	)

	req := httptest.NewRequest("DELETE", "/api/chores/99999", nil)
	req.SetPathValue("id", "99999")
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeleteChore(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// =====================================================
// End-to-end integration test (T062)
// =====================================================

func TestChoreLifecycle_EndToEnd(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	choreRepo := repositories.NewChoreRepo(db)
	instanceRepo := repositories.NewChoreInstanceRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)

	handler := NewHandler(choreRepo, instanceRepo, txRepo, childRepo)

	// Step 1: Parent creates a chore
	createBody := fmt.Sprintf(`{"name":"Wash car","description":"Wash the family car","reward_cents":1000,"recurrence":"one_time","child_ids":[%d]}`, child.ID)
	createReq := httptest.NewRequest("POST", "/api/chores", bytes.NewBufferString(createBody))
	createReq = testutil.SetRequestContext(createReq, "parent", parent.ID, family.ID)

	createRR := httptest.NewRecorder()
	handler.HandleCreateChore(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)

	// Get instance ID from child's chore list
	listReq := httptest.NewRequest("GET", "/api/child/chores", nil)
	listReq = testutil.SetRequestContext(listReq, "child", child.ID, family.ID)
	listRR := httptest.NewRecorder()
	handler.HandleChildListChores(listRR, listReq)
	require.Equal(t, http.StatusOK, listRR.Code)

	var listResp struct {
		Available []InstanceResponse `json:"available"`
	}
	err := json.Unmarshal(listRR.Body.Bytes(), &listResp)
	require.NoError(t, err)
	require.Len(t, listResp.Available, 1)

	instanceID := listResp.Available[0].ID

	// Step 2: Child marks it complete
	completeReq := httptest.NewRequest("POST", "/api/child/chores/"+strconv.FormatInt(instanceID, 10)+"/complete", nil)
	completeReq.SetPathValue("id", strconv.FormatInt(instanceID, 10))
	completeReq = testutil.SetRequestContext(completeReq, "child", child.ID, family.ID)

	completeRR := httptest.NewRecorder()
	handler.HandleCompleteChore(completeRR, completeReq)
	require.Equal(t, http.StatusOK, completeRR.Code)

	// Step 3: Parent approves
	approveReq := httptest.NewRequest("POST", "/api/chore-instances/"+strconv.FormatInt(instanceID, 10)+"/approve", nil)
	approveReq.SetPathValue("id", strconv.FormatInt(instanceID, 10))
	approveReq = testutil.SetRequestContext(approveReq, "parent", parent.ID, family.ID)

	approveRR := httptest.NewRecorder()
	handler.HandleApprove(approveRR, approveReq)
	require.Equal(t, http.StatusOK, approveRR.Code)

	var approveResp struct {
		Instance   InstanceResponse `json:"instance"`
		NewBalance int64            `json:"new_balance"`
	}
	err = json.Unmarshal(approveRR.Body.Bytes(), &approveResp)
	require.NoError(t, err)

	// Verify instance is approved with transaction
	assert.Equal(t, models.ChoreInstanceStatusApproved, approveResp.Instance.Status)
	assert.NotNil(t, approveResp.Instance.TransactionID)

	// Verify child balance increased by reward
	assert.Equal(t, int64(1000), approveResp.NewBalance)

	// Verify transaction exists with correct type and note
	txs, err := txRepo.ListByChild(child.ID)
	require.NoError(t, err)
	require.Len(t, txs, 1)
	assert.Equal(t, models.TransactionTypeChore, txs[0].TransactionType)
	assert.Equal(t, int64(1000), txs[0].AmountCents)
	require.NotNil(t, txs[0].Note)
	assert.Contains(t, *txs[0].Note, "Wash car")
}
