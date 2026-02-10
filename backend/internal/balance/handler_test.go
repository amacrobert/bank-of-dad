package balance

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

	tmp, err := os.CreateTemp("", "bank-of-dad-balance-test-*.db")
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

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func createTestFamily(t *testing.T, db *store.DB) *store.Family {
	familyStore := store.NewFamilyStore(db)
	family, err := familyStore.Create("test-family")
	require.NoError(t, err)
	return family
}

func createTestParent(t *testing.T, db *store.DB, familyID int64) *store.Parent {
	parentStore := store.NewParentStore(db)
	parent, err := parentStore.Create("google-id-123", "parent@test.com", "Test Parent")
	require.NoError(t, err)
	err = parentStore.SetFamilyID(parent.ID, familyID)
	require.NoError(t, err)
	parent.FamilyID = familyID
	return parent
}

func createTestChild(t *testing.T, db *store.DB, familyID int64, name string) *store.Child {
	childStore := store.NewChildStore(db)
	child, err := childStore.Create(familyID, name, "password123")
	require.NoError(t, err)
	return child
}

// setRequestContext adds authentication context to a request
func setRequestContext(r *http.Request, userType string, userID, familyID int64) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyUserType, userType)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, middleware.ContextKeyFamilyID, familyID)
	return r.WithContext(ctx)
}

// =====================================================
// T023: Tests for POST /api/children/{id}/deposit
// =====================================================

func TestHandleDeposit_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 1000, "note": "Weekly allowance"}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), resp.Transaction.AmountCents)
	assert.Equal(t, "deposit", string(resp.Transaction.TransactionType))
	assert.Equal(t, "Weekly allowance", *resp.Transaction.Note)
	assert.Equal(t, int64(1000), resp.NewBalanceCents)
	assert.Equal(t, child.ID, resp.Transaction.ChildID)
	assert.Equal(t, parent.ID, resp.Transaction.ParentID)
}

func TestHandleDeposit_MultipleDeposits(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	// First deposit
	body := `{"amount_cents": 1000}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)
	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Second deposit
	body2 := `{"amount_cents": 500}`
	req2 := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body2))
	req2.SetPathValue("id", "1")
	req2 = setRequestContext(req2, "parent", parent.ID, family.ID)
	rr2 := httptest.NewRecorder()
	handler.HandleDeposit(rr2, req2)

	assert.Equal(t, http.StatusOK, rr2.Code)

	var resp TransactionResponse
	err := json.Unmarshal(rr2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(500), resp.Transaction.AmountCents)
	assert.Equal(t, int64(1500), resp.NewBalanceCents) // 1000 + 500
}

func TestHandleDeposit_WithoutNote(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 500}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Nil(t, resp.Transaction.Note)
}

// =====================================================
// T024: Tests for deposit validation (amount > 0, max amount)
// =====================================================

func TestHandleDeposit_InvalidAmount_Zero(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 0}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_amount", resp.Error)
}

func TestHandleDeposit_InvalidAmount_Negative(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": -100}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_amount", resp.Error)
}

func TestHandleDeposit_InvalidAmount_ExceedsMax(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	// Max is 99999999 cents ($999,999.99)
	body := `{"amount_cents": 100000000}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_amount", resp.Error)
}

func TestHandleDeposit_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{invalid json}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleDeposit_NoteTooLong(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	// Note longer than 500 characters
	longNote := make([]byte, 501)
	for i := range longNote {
		longNote[i] = 'a'
	}

	body := `{"amount_cents": 1000, "note": "` + string(longNote) + `"}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_note", resp.Error)
}

func TestHandleDeposit_NoteWhitespaceOnly(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	// Whitespace-only note should be treated as empty
	body := `{"amount_cents": 1000, "note": "   "}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	// Whitespace-only note should be treated as nil
	assert.Nil(t, resp.Transaction.Note)
}

// =====================================================
// T025: Tests for deposit authorization (parent only, own family)
// =====================================================

func TestHandleDeposit_Unauthorized_ChildCannotDeposit(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 1000}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	// Set context as child instead of parent
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	// Should be forbidden - only parents can deposit
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleDeposit_Forbidden_WrongFamily(t *testing.T) {
	db := setupTestDB(t)
	family1 := createTestFamily(t, db)
	family2 := &store.Family{ID: 999} // Different family ID

	parent := createTestParent(t, db, family1.ID)
	createTestChild(t, db, family1.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 1000}`
	req := httptest.NewRequest("POST", "/api/children/1/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	// Parent from family2 trying to deposit to child in family1
	req = setRequestContext(req, "parent", parent.ID, family2.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleDeposit_NotFound_ChildDoesNotExist(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 1000}`
	req := httptest.NewRequest("POST", "/api/children/999/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "999")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleDeposit_InvalidChildID(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	body := `{"amount_cents": 1000}`
	req := httptest.NewRequest("POST", "/api/children/abc/deposit", bytes.NewBufferString(body))
	req.SetPathValue("id", "abc")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleDeposit(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// =====================================================
// T037: Tests for POST /api/children/{id}/withdraw
// =====================================================

func TestHandleWithdraw_Success(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// First deposit some money
	_, _, err := txStore.Deposit(child.ID, parent.ID, 5000, "Initial deposit")
	require.NoError(t, err)

	// Now withdraw
	body := `{"amount_cents": 1500, "note": "Bought a book"}`
	req := httptest.NewRequest("POST", "/api/children/1/withdraw", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleWithdraw(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(1500), resp.Transaction.AmountCents)
	assert.Equal(t, "withdrawal", string(resp.Transaction.TransactionType))
	assert.Equal(t, "Bought a book", *resp.Transaction.Note)
	assert.Equal(t, int64(3500), resp.NewBalanceCents) // 5000 - 1500
}

// =====================================================
// T038: Tests for insufficient funds error
// =====================================================

func TestHandleWithdraw_InsufficientFunds(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Deposit only $10
	_, _, err := txStore.Deposit(child.ID, parent.ID, 1000, "")
	require.NoError(t, err)

	// Try to withdraw $20
	body := `{"amount_cents": 2000}`
	req := httptest.NewRequest("POST", "/api/children/1/withdraw", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleWithdraw(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "insufficient_funds", resp.Error)
	assert.Contains(t, resp.Message, "20.00")
	assert.Contains(t, resp.Message, "10.00")
}

func TestHandleWithdraw_InsufficientFunds_ZeroBalance(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	// Try to withdraw from $0 balance
	body := `{"amount_cents": 100}`
	req := httptest.NewRequest("POST", "/api/children/1/withdraw", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleWithdraw(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "insufficient_funds", resp.Error)
}

// =====================================================
// T039: Tests for withdraw to exactly $0.00 (should succeed)
// =====================================================

func TestHandleWithdraw_ExactBalance(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Deposit exactly $25.00
	_, _, err := txStore.Deposit(child.ID, parent.ID, 2500, "")
	require.NoError(t, err)

	// Withdraw exactly $25.00 - should succeed with $0 balance
	body := `{"amount_cents": 2500}`
	req := httptest.NewRequest("POST", "/api/children/1/withdraw", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleWithdraw(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(2500), resp.Transaction.AmountCents)
	assert.Equal(t, int64(0), resp.NewBalanceCents) // Should be exactly $0.00
}

func TestHandleWithdraw_ChildCannotWithdraw(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Give the child some money
	_, _, err := txStore.Deposit(child.ID, parent.ID, 1000, "")
	require.NoError(t, err)

	// Child tries to withdraw
	body := `{"amount_cents": 500}`
	req := httptest.NewRequest("POST", "/api/children/1/withdraw", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleWithdraw(rr, req)

	// Children cannot withdraw - only parents can
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleWithdraw_WrongFamily(t *testing.T) {
	db := setupTestDB(t)
	family1 := createTestFamily(t, db)
	parent := createTestParent(t, db, family1.ID)
	child := createTestChild(t, db, family1.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Give the child some money
	_, _, err := txStore.Deposit(child.ID, parent.ID, 1000, "")
	require.NoError(t, err)

	// Parent from different family tries to withdraw
	body := `{"amount_cents": 500}`
	req := httptest.NewRequest("POST", "/api/children/1/withdraw", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, 999) // Wrong family ID

	rr := httptest.NewRecorder()
	handler.HandleWithdraw(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// =====================================================
// T048: Tests for GET /api/children/{id}/balance
// =====================================================

func TestHandleGetBalance_ChildViewsOwn(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Give the child some money
	_, _, err := txStore.Deposit(child.ID, parent.ID, 2500, "")
	require.NoError(t, err)

	// Child views own balance
	req := httptest.NewRequest("GET", "/api/children/1/balance", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp BalanceResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, int64(2500), resp.BalanceCents)
}

func TestHandleGetBalance_ParentViewsChild(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Give the child some money
	_, _, err := txStore.Deposit(child.ID, parent.ID, 5000, "")
	require.NoError(t, err)

	// Parent views child's balance
	req := httptest.NewRequest("GET", "/api/children/1/balance", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp BalanceResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, int64(5000), resp.BalanceCents)
}

// =====================================================
// T049: Tests for GET /api/children/{id}/transactions
// =====================================================

func TestHandleGetTransactions_ChildViewsOwn(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Create some transactions
	_, _, err := txStore.Deposit(child.ID, parent.ID, 1000, "First deposit")
	require.NoError(t, err)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 500, "Second deposit")
	require.NoError(t, err)
	_, _, err = txStore.Withdraw(child.ID, parent.ID, 200, "Small withdrawal")
	require.NoError(t, err)

	// Child views own transactions
	req := httptest.NewRequest("GET", "/api/children/1/transactions", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetTransactions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Transactions, 3)
}

func TestHandleGetTransactions_ParentViewsChild(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Create a transaction
	_, _, err := txStore.Deposit(child.ID, parent.ID, 1000, "Allowance")
	require.NoError(t, err)

	// Parent views child's transactions
	req := httptest.NewRequest("GET", "/api/children/1/transactions", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetTransactions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Transactions, 1)
}

// =====================================================
// T050: Tests for child can only view own balance (not siblings)
// =====================================================

func TestHandleGetBalance_ChildCannotViewSibling(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child1 := createTestChild(t, db, family.ID, "Emma")
	child2 := createTestChild(t, db, family.ID, "Jack")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Give child2 some money
	_, _, err := txStore.Deposit(child2.ID, parent.ID, 5000, "")
	require.NoError(t, err)

	// Child1 tries to view child2's balance
	req := httptest.NewRequest("GET", "/api/children/2/balance", nil)
	req.SetPathValue("id", "2")
	req = setRequestContext(req, "child", child1.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetBalance(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleGetTransactions_ChildCannotViewSibling(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child1 := createTestChild(t, db, family.ID, "Emma")
	child2 := createTestChild(t, db, family.ID, "Jack")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Give child2 some money
	_, _, err := txStore.Deposit(child2.ID, parent.ID, 1000, "")
	require.NoError(t, err)

	// Child1 tries to view child2's transactions
	req := httptest.NewRequest("GET", "/api/children/2/transactions", nil)
	req.SetPathValue("id", "2")
	req = setRequestContext(req, "child", child1.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetTransactions(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// =====================================================
// T051: Tests for transactions ordered newest-first
// =====================================================

func TestHandleGetTransactions_OrderedNewestFirst(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	// Create transactions in order
	_, _, err := txStore.Deposit(child.ID, parent.ID, 100, "First")
	require.NoError(t, err)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 200, "Second")
	require.NoError(t, err)
	_, _, err = txStore.Deposit(child.ID, parent.ID, 300, "Third")
	require.NoError(t, err)

	// Get transactions
	req := httptest.NewRequest("GET", "/api/children/1/transactions", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetTransactions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Should be newest first (300, 200, 100)
	require.Len(t, resp.Transactions, 3)
	assert.Equal(t, int64(300), resp.Transactions[0].AmountCents)
	assert.Equal(t, int64(200), resp.Transactions[1].AmountCents)
	assert.Equal(t, int64(100), resp.Transactions[2].AmountCents)
}

func TestHandleGetTransactions_EmptyHistory(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	handler := NewHandler(
		store.NewTransactionStore(db),
		store.NewChildStore(db),
		store.NewInterestStore(db),
		store.NewInterestScheduleStore(db),
	)

	// No transactions exist
	req := httptest.NewRequest("GET", "/api/children/1/transactions", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "child", child.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetTransactions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp TransactionListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Transactions) // Should be empty array, not null
	assert.Len(t, resp.Transactions, 0)
}

// =====================================================
// T012: Tests for enhanced balance response with interest rate fields
// =====================================================

func TestHandleGetBalance_IncludesInterestRate(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	interestStore := store.NewInterestStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), interestStore, store.NewInterestScheduleStore(db))

	// Give the child money and set interest rate
	_, _, err := txStore.Deposit(child.ID, parent.ID, 10000, "")
	require.NoError(t, err)
	err = interestStore.SetInterestRate(child.ID, 500)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/children/1/balance", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp BalanceResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, child.ID, resp.ChildID)
	assert.Equal(t, "Emma", resp.FirstName)
	assert.Equal(t, int64(10000), resp.BalanceCents)
	assert.Equal(t, 500, resp.InterestRateBps)
	assert.Equal(t, "5.00%", resp.InterestRateDisplay)
}

func TestHandleGetBalance_DefaultInterestRateZero(t *testing.T) {
	db := setupTestDB(t)
	family := createTestFamily(t, db)
	parent := createTestParent(t, db, family.ID)
	child := createTestChild(t, db, family.ID, "Emma")

	txStore := store.NewTransactionStore(db)
	handler := NewHandler(txStore, store.NewChildStore(db), store.NewInterestStore(db), store.NewInterestScheduleStore(db))

	_, _, err := txStore.Deposit(child.ID, parent.ID, 5000, "")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/children/1/balance", nil)
	req.SetPathValue("id", "1")
	req = setRequestContext(req, "parent", parent.ID, family.ID)

	rr := httptest.NewRecorder()
	handler.HandleGetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp BalanceResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.InterestRateBps)
	assert.Equal(t, "0.00%", resp.InterestRateDisplay)
}
