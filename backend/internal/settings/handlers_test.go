package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bank-of-dad/internal/store"
	"bank-of-dad/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandlers(t *testing.T) (*Handlers, *store.FamilyStore) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	familyStore := store.NewFamilyStore(db)
	h := NewHandlers(familyStore)
	return h, familyStore
}

// --- GET /api/settings ---

func TestHandleGetSettings_ReturnsDefaultTimezone(t *testing.T) {
	h, fs := newTestHandlers(t)

	fam, err := fs.Create("settings-family")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/settings", nil)
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleGetSettings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", resp["timezone"])
}

func TestHandleGetSettings_NoFamilyID(t *testing.T) {
	h, _ := newTestHandlers(t)

	req := httptest.NewRequest("GET", "/api/settings", nil)
	req = testutil.SetRequestContext(req, "parent", 1, 0)
	rr := httptest.NewRecorder()

	h.HandleGetSettings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- PUT /api/settings/timezone ---

func TestHandleUpdateTimezone_ValidTimezone(t *testing.T) {
	h, fs := newTestHandlers(t)

	fam, err := fs.Create("tz-update-fam")
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{"timezone": "America/Chicago"})
	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader(body))
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Timezone updated", resp["message"])
	assert.Equal(t, "America/Chicago", resp["timezone"])

	// Verify it persisted
	tz, err := fs.GetTimezone(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "America/Chicago", tz)
}

func TestHandleUpdateTimezone_InvalidTimezone(t *testing.T) {
	h, fs := newTestHandlers(t)

	fam, err := fs.Create("tz-invalid-fam")
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{"timezone": "Fake/Timezone"})
	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader(body))
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid timezone", resp["error"])
}

func TestHandleUpdateTimezone_EmptyBody(t *testing.T) {
	h, fs := newTestHandlers(t)

	fam, err := fs.Create("tz-empty-fam")
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader([]byte("{}")))
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUpdateTimezone_NoFamilyID(t *testing.T) {
	h, _ := newTestHandlers(t)

	body, _ := json.Marshal(map[string]string{"timezone": "America/Chicago"})
	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader(body))
	req = testutil.SetRequestContext(req, "parent", 1, 0)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUpdateTimezone_MalformedJSON(t *testing.T) {
	h, fs := newTestHandlers(t)

	fam, err := fs.Create("tz-malformed-fam")
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader([]byte("not json")))
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
