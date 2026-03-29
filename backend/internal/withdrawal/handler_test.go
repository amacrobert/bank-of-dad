package withdrawal

import (
	"bytes"
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
)

func setupHandler(t *testing.T) (*Handler, *repositories.WithdrawalRequestRepo, *repositories.TransactionRepo, *repositories.ChildRepo, *repositories.SavingsGoalRepo) {
	db := testutil.SetupTestDB(t)
	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)
	return handler, wrRepo, txRepo, childRepo, goalRepo
}

// =====================================================
// Tests for POST /api/child/withdrawal-requests (HandleSubmitRequest)
// =====================================================

func TestHandleSubmitRequest_Success(t *testing.T) {
	handler, _, _, _, _ := setupHandler(t)
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	// Give child a balance
	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, 1, 5000, "seed")
	require.NoError(t, err)

	// Re-create handler with this DB
	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler = NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	body := `{"amount_cents":2000,"reason":"New video game"}`
	req := httptest.NewRequest("POST", "/api/child/withdrawal-requests", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSubmitRequest(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var wr models.WithdrawalRequest
	err = json.Unmarshal(resp["withdrawal_request"], &wr)
	require.NoError(t, err)

	assert.Equal(t, 2000, wr.AmountCents)
	assert.Equal(t, "New video game", wr.Reason)
	assert.Equal(t, models.WithdrawalRequestStatusPending, wr.Status)
	assert.Equal(t, child.ID, wr.ChildID)
	assert.Equal(t, family.ID, wr.FamilyID)
}

func TestHandleSubmitRequest_InsufficientFunds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	// Child has $0 balance
	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	body := `{"amount_cents":2000,"reason":"Something"}`
	req := httptest.NewRequest("POST", "/api/child/withdrawal-requests", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSubmitRequest(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestHandleSubmitRequest_AlreadyPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, 1, 10000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	// Create first request
	_, err = wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "First request",
	})
	require.NoError(t, err)

	// Try to create second request
	body := `{"amount_cents":2000,"reason":"Second request"}`
	req := httptest.NewRequest("POST", "/api/child/withdrawal-requests", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSubmitRequest(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestHandleSubmitRequest_DisabledAccount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	// Disable the child's account
	db.Model(&models.Child{}).Where("id = ?", child.ID).Update("is_disabled", true)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	body := `{"amount_cents":100,"reason":"Something"}`
	req := httptest.NewRequest("POST", "/api/child/withdrawal-requests", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSubmitRequest(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleSubmitRequest_InvalidInput(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, 1, 5000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	tests := []struct {
		name string
		body string
	}{
		{"zero amount", `{"amount_cents":0,"reason":"test"}`},
		{"negative amount", `{"amount_cents":-100,"reason":"test"}`},
		{"empty reason", `{"amount_cents":100,"reason":""}`},
		{"amount too large", `{"amount_cents":100000000,"reason":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/child/withdrawal-requests", bytes.NewBufferString(tt.body))
			req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

			rr := httptest.NewRecorder()
			handler.HandleSubmitRequest(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestHandleSubmitRequest_ParentForbidden(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	body := `{"amount_cents":100,"reason":"test"}`
	req := httptest.NewRequest("POST", "/api/child/withdrawal-requests", bytes.NewBufferString(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleSubmitRequest(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// =====================================================
// Tests for POST /api/withdrawal-requests/{id}/approve (HandleApprove)
// =====================================================

func TestHandleApprove_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 5000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 2000,
		Reason:      "New game",
	})
	require.NoError(t, err)

	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/approve", wr.ID), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check balance was deducted
	var balanceCents int64
	err = json.Unmarshal(resp["new_balance_cents"], &balanceCents)
	require.NoError(t, err)
	assert.Equal(t, int64(3000), balanceCents)

	// Check request was updated
	var updatedWR models.WithdrawalRequest
	err = json.Unmarshal(resp["withdrawal_request"], &updatedWR)
	require.NoError(t, err)
	assert.Equal(t, models.WithdrawalRequestStatusApproved, updatedWR.Status)
	assert.NotNil(t, updatedWR.TransactionID)
}

func TestHandleApprove_InsufficientFunds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	// Give only 1000 cents
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 1000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	// Request 2000 cents (more than available)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 2000,
		Reason:      "Too much",
	})
	require.NoError(t, err)

	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/approve", wr.ID), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestHandleApprove_NotPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 5000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	// Deny it first
	err = wrRepo.Deny(wr.ID, parent.ID, "no")
	require.NoError(t, err)

	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/approve", wr.ID), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestHandleApprove_WrongFamily(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family1 := testutil.CreateTestFamily(t, db)
	parent1 := testutil.CreateTestParent(t, db, family1.ID)
	child := testutil.CreateTestChild(t, db, family1.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, parent1.ID, 5000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family1.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	// Use a different family ID
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/approve", wr.ID), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent1.ID, 9999)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleApprove_DisabledAccount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	txRepo := repositories.NewTransactionRepo(db)
	_, _, err := txRepo.Deposit(child.ID, parent.ID, 5000, "seed")
	require.NoError(t, err)

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	// Disable the child
	db.Model(&models.Child{}).Where("id = ?", child.ID).Update("is_disabled", true)

	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/approve", wr.ID), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleApprove(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

// =====================================================
// Tests for POST /api/withdrawal-requests/{id}/deny (HandleDeny)
// =====================================================

func TestHandleDeny_Success_WithReason(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	body := `{"reason":"Save up a bit more first"}`
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/deny", wr.ID), bytes.NewBufferString(body))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeny(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var updatedWR models.WithdrawalRequest
	err = json.Unmarshal(resp["withdrawal_request"], &updatedWR)
	require.NoError(t, err)
	assert.Equal(t, models.WithdrawalRequestStatusDenied, updatedWR.Status)
	assert.NotNil(t, updatedWR.DenialReason)
	assert.Equal(t, "Save up a bit more first", *updatedWR.DenialReason)
}

func TestHandleDeny_Success_WithoutReason(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/withdrawal-requests/%d/deny", wr.ID), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeny(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// =====================================================
// Tests for POST /api/child/withdrawal-requests/{id}/cancel (HandleCancelRequest)
// =====================================================

func TestHandleCancelRequest_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/child/withdrawal-requests/%d/cancel", wr.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCancelRequest(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var updatedWR models.WithdrawalRequest
	err = json.Unmarshal(resp["withdrawal_request"], &updatedWR)
	require.NoError(t, err)
	assert.Equal(t, models.WithdrawalRequestStatusCancelled, updatedWR.Status)
}

func TestHandleCancelRequest_NotPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	// Deny it first
	err = wrRepo.Deny(wr.ID, parent.ID, "no")
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/child/withdrawal-requests/%d/cancel", wr.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCancelRequest(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestHandleCancelRequest_WrongChild(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	testutil.CreateTestParent(t, db, family.ID)
	child1 := testutil.CreateTestChild(t, db, family.ID, "Alice")
	child2 := testutil.CreateTestChild(t, db, family.ID, "Bob")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child1.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	// child2 tries to cancel child1's request
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/child/withdrawal-requests/%d/cancel", wr.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", wr.ID))
	req = testutil.SetRequestContext(req, "child", child2.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleCancelRequest(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

// =====================================================
// Tests for GET /api/child/withdrawal-requests (HandleChildListRequests)
// =====================================================

func TestHandleChildListRequests_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	_, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "pending request",
	})
	require.NoError(t, err)

	// Create a second request that is denied (cancel first, create new, deny)
	_ = parent // used to denote we have a parent for deny

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("GET", "/api/child/withdrawal-requests", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var requests []models.WithdrawalRequest
	err = json.Unmarshal(resp["withdrawal_requests"], &requests)
	require.NoError(t, err)
	assert.Len(t, requests, 1)
}

func TestHandleChildListRequests_WithStatusFilter(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	wr, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "first",
	})
	require.NoError(t, err)

	// Deny it
	err = wrRepo.Deny(wr.ID, parent.ID, "no")
	require.NoError(t, err)

	// Create another pending
	_, err = wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 500,
		Reason:      "second",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	// Filter for pending only
	req := httptest.NewRequest("GET", "/api/child/withdrawal-requests?status=pending", nil)
	req = testutil.SetRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleChildListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var requests []models.WithdrawalRequest
	err = json.Unmarshal(resp["withdrawal_requests"], &requests)
	require.NoError(t, err)
	assert.Len(t, requests, 1)
	assert.Equal(t, models.WithdrawalRequestStatusPending, requests[0].Status)
}

// =====================================================
// Tests for GET /api/withdrawal-requests (HandleParentListRequests)
// =====================================================

func TestHandleParentListRequests_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child := testutil.CreateTestChild(t, db, family.ID, "Alice")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	_, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "test",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("GET", "/api/withdrawal-requests", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleParentListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]json.RawMessage
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	var requests []repositories.WithdrawalRequestWithChild
	err = json.Unmarshal(resp["withdrawal_requests"], &requests)
	require.NoError(t, err)
	assert.Len(t, requests, 1)
	assert.Equal(t, "Alice", requests[0].ChildName)
}

// =====================================================
// Tests for GET /api/withdrawal-requests/pending/count (HandlePendingCount)
// =====================================================

func TestHandlePendingCount_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	family := testutil.CreateTestFamily(t, db)
	parent := testutil.CreateTestParent(t, db, family.ID)
	child1 := testutil.CreateTestChild(t, db, family.ID, "Alice")
	child2 := testutil.CreateTestChild(t, db, family.ID, "Bob")

	wrRepo := repositories.NewWithdrawalRequestRepo(db)
	_, err := wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child1.ID,
		FamilyID:    family.ID,
		AmountCents: 1000,
		Reason:      "alice request",
	})
	require.NoError(t, err)
	_, err = wrRepo.Create(&models.WithdrawalRequest{
		ChildID:     child2.ID,
		FamilyID:    family.ID,
		AmountCents: 500,
		Reason:      "bob request",
	})
	require.NoError(t, err)

	txRepo := repositories.NewTransactionRepo(db)
	childRepo := repositories.NewChildRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	handler := NewHandler(wrRepo, txRepo, childRepo, goalRepo)

	req := httptest.NewRequest("GET", "/api/withdrawal-requests/pending/count", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandlePendingCount(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]int64
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp["count"])
}
