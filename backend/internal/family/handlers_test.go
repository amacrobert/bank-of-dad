package family

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bank-of-dad/internal/store"
	"bank-of-dad/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandlers(t *testing.T) (*Handlers, *store.FamilyStore, *store.ChildStore) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	familyStore := store.NewFamilyStore(db)
	parentStore := store.NewParentStore(db)
	childStore := store.NewChildStore(db)
	eventStore := store.NewAuthEventStore(db)
	h := NewHandlers(familyStore, parentStore, childStore, eventStore, []byte("test-key"))
	return h, familyStore, childStore
}

func TestHandleListFamilyChildren_WithChildren(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("smith-family")
	require.NoError(t, err)

	avatar := "🦊"
	_, err = childStore.Create(fam.ID, "Alice", "password123", &avatar)
	require.NoError(t, err)
	_, err = childStore.Create(fam.ID, "Bob", "password456", nil)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/families/smith-family/children", nil)
	req.SetPathValue("slug", "smith-family")
	rr := httptest.NewRecorder()

	h.HandleListFamilyChildren(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Children []struct {
			FirstName string  `json:"first_name"`
			Avatar    *string `json:"avatar"`
		} `json:"children"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Children, 2)
	// Sorted by first_name (ListByFamily orders alphabetically)
	assert.Equal(t, "Alice", resp.Children[0].FirstName)
	require.NotNil(t, resp.Children[0].Avatar)
	assert.Equal(t, "🦊", *resp.Children[0].Avatar)

	assert.Equal(t, "Bob", resp.Children[1].FirstName)
	assert.Nil(t, resp.Children[1].Avatar)
}

func TestHandleListFamilyChildren_NonexistentSlug(t *testing.T) {
	h, _, _ := newTestHandlers(t)

	req := httptest.NewRequest("GET", "/api/families/nonexistent/children", nil)
	req.SetPathValue("slug", "nonexistent")
	rr := httptest.NewRecorder()

	h.HandleListFamilyChildren(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Children []interface{} `json:"children"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Children)
}

func TestHandleListFamilyChildren_NoSensitiveFields(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("doe-family")
	require.NoError(t, err)
	_, err = childStore.Create(fam.ID, "Charlie", "password789", nil)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/families/doe-family/children", nil)
	req.SetPathValue("slug", "doe-family")
	rr := httptest.NewRecorder()

	h.HandleListFamilyChildren(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse as raw JSON to check for unexpected fields
	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	require.NoError(t, err)

	children := raw["children"]
	require.Len(t, children, 1)

	child := children[0]
	// Only first_name and avatar should be present
	assert.Contains(t, child, "first_name")
	assert.Contains(t, child, "avatar")
	assert.NotContains(t, child, "id")
	assert.NotContains(t, child, "balance_cents")
	assert.NotContains(t, child, "is_locked")
	assert.NotContains(t, child, "password_hash")
	assert.NotContains(t, child, "family_id")
}

// T006: HandleUpdateTheme tests
func TestHandleUpdateTheme_ValidTheme(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	child, err := childStore.Create(fam.ID, "Alice", "password123", nil)
	require.NoError(t, err)

	body := strings.NewReader(`{"theme":"piggybank"}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/theme", body)
	req = testutil.SetRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTheme(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Message string `json:"message"`
		Theme   string `json:"theme"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Theme updated", resp.Message)
	assert.Equal(t, "piggybank", resp.Theme)

	// Verify persisted
	updated, err := childStore.GetByID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Theme)
	assert.Equal(t, "piggybank", *updated.Theme)
}

func TestHandleUpdateTheme_InvalidTheme(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	child, err := childStore.Create(fam.ID, "Alice", "password123", nil)
	require.NoError(t, err)

	body := strings.NewReader(`{"theme":"darkmode"}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/theme", body)
	req = testutil.SetRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTheme(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUpdateTheme_NonChildUser(t *testing.T) {
	h, familyStore, _ := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	ps := store.NewParentStore(testutil.SetupTestDB(t))
	parent, err := ps.Create("g-123", "p@test.com", "Parent")
	require.NoError(t, err)

	body := strings.NewReader(`{"theme":"sparkle"}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/theme", body)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTheme(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleUpdateAvatar_SetAvatar(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	child, err := childStore.Create(fam.ID, "Alice", "password123", nil)
	require.NoError(t, err)

	body := strings.NewReader(`{"avatar":"🦊"}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/avatar", body)
	req = testutil.SetRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateAvatar(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Message string  `json:"message"`
		Avatar  *string `json:"avatar"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Avatar updated", resp.Message)
	require.NotNil(t, resp.Avatar)
	assert.Equal(t, "🦊", *resp.Avatar)

	// Verify persisted
	updated, err := childStore.GetByID(child.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Avatar)
	assert.Equal(t, "🦊", *updated.Avatar)
}

func TestHandleUpdateAvatar_ClearAvatar(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	avatar := "🐻"
	child, err := childStore.Create(fam.ID, "Alice", "password123", &avatar)
	require.NoError(t, err)

	body := strings.NewReader(`{"avatar":null}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/avatar", body)
	req = testutil.SetRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateAvatar(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify cleared
	updated, err := childStore.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Avatar)
}

func TestHandleUpdateAvatar_EmptyStringClearsAvatar(t *testing.T) {
	h, familyStore, childStore := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	avatar := "🐻"
	child, err := childStore.Create(fam.ID, "Alice", "password123", &avatar)
	require.NoError(t, err)

	body := strings.NewReader(`{"avatar":""}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/avatar", body)
	req = testutil.SetRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateAvatar(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify cleared
	updated, err := childStore.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Avatar)
}

func TestHandleUpdateAvatar_NonChildUser(t *testing.T) {
	h, familyStore, _ := newTestHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	ps := store.NewParentStore(testutil.SetupTestDB(t))
	parent, err := ps.Create("g-456", "p2@test.com", "Parent")
	require.NoError(t, err)

	body := strings.NewReader(`{"avatar":"🦊"}`)
	req := httptest.NewRequest("PUT", "/api/child/settings/avatar", body)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateAvatar(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// HandleDeleteAccount subscription guard tests

func setupDeleteAccountTest(t *testing.T) (*Handlers, *store.FamilyStore, *store.ParentStore, *store.Family, *store.Parent) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	familyStore := store.NewFamilyStore(db)
	parentStore := store.NewParentStore(db)
	childStore := store.NewChildStore(db)
	eventStore := store.NewAuthEventStore(db)
	h := NewHandlers(familyStore, parentStore, childStore, eventStore, []byte("test-key"))

	fam, err := familyStore.Create("del-test-family")
	require.NoError(t, err)

	parent, err := parentStore.Create("g-del-1", "del@test.com", "Del Parent")
	require.NoError(t, err)

	err = parentStore.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	return h, familyStore, parentStore, fam, parent
}

func TestHandleDeleteAccount_ActiveSubscriptionBlocks(t *testing.T) {
	h, familyStore, _, fam, parent := setupDeleteAccountTest(t)

	// Set up active, un-cancelled subscription
	err := familyStore.UpdateSubscriptionFromCheckout(fam.ID, "cus_123", "sub_123", "active", time.Now().Add(30*24*time.Hour))
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/api/account", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleDeleteAccount(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "cancel your subscription")
}

func TestHandleDeleteAccount_CancellingSubscriptionAllows(t *testing.T) {
	h, familyStore, _, fam, parent := setupDeleteAccountTest(t)

	// Set up active subscription that is cancelling at period end
	err := familyStore.UpdateSubscriptionFromCheckout(fam.ID, "cus_234", "sub_234", "active", time.Now().Add(30*24*time.Hour))
	require.NoError(t, err)
	err = familyStore.UpdateSubscriptionStatus("sub_234", "active", time.Now().Add(30*24*time.Hour), true)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/api/account", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleDeleteAccount(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandleDeleteAccount_NoSubscriptionAllows(t *testing.T) {
	h, _, _, fam, parent := setupDeleteAccountTest(t)

	// No subscription set up — free account
	req := httptest.NewRequest("DELETE", "/api/account", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleDeleteAccount(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandleCreateChild_MaxChildrenLimit(t *testing.T) {
	db := testutil.SetupTestDB(t)
	familyStore := store.NewFamilyStore(db)
	parentStore := store.NewParentStore(db)
	childStore := store.NewChildStore(db)
	eventStore := store.NewAuthEventStore(db)
	h := NewHandlers(familyStore, parentStore, childStore, eventStore, []byte("test-key"))

	fam, err := familyStore.Create("big-family")
	require.NoError(t, err)

	parent, err := parentStore.Create("g-limit-1", "limit@test.com", "Limit Parent")
	require.NoError(t, err)
	err = parentStore.SetFamilyID(parent.ID, fam.ID)
	require.NoError(t, err)

	// Create 20 children
	for i := 1; i <= 20; i++ {
		_, err := childStore.Create(fam.ID, fmt.Sprintf("Child%d", i), "password123", nil)
		require.NoError(t, err)
	}

	// 21st should be rejected
	body := strings.NewReader(`{"first_name":"Child21","password":"password123"}`)
	req := httptest.NewRequest("POST", "/api/children", body)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleCreateChild(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Limit reached", resp["error"])
}
