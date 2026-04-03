package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bank-of-dad/internal/testutil"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandlers(t *testing.T) (*Handlers, *repositories.FamilyRepo, *repositories.ParentRepo) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	familyRepo := repositories.NewFamilyRepo(db)
	parentRepo := repositories.NewParentRepo(db)
	h := NewHandlers(familyRepo, parentRepo)
	return h, familyRepo, parentRepo
}

// --- GET /api/settings ---

func TestHandleGetSettings_ReturnsDefaultTimezone(t *testing.T) {
	h, fs, pr := newTestHandlers(t)

	fam, err := fs.Create("settings-family")
	require.NoError(t, err)

	p, err := pr.Create("google-settings", "settings@example.com", "Settings User")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(p.ID, fam.ID))

	req := httptest.NewRequest("GET", "/api/settings", nil)
	req = testutil.SetRequestContext(req, "parent", p.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleGetSettings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", resp["timezone"])

	// Verify notifications object is present with defaults
	notifications, ok := resp["notifications"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, notifications["notify_withdrawal_requests"])
	assert.Equal(t, true, notifications["notify_chore_completions"])
	assert.Equal(t, true, notifications["notify_decisions"])
}

func TestHandleGetSettings_NoFamilyID(t *testing.T) {
	h, _, _ := newTestHandlers(t)

	req := httptest.NewRequest("GET", "/api/settings", nil)
	req = testutil.SetRequestContext(req, "parent", 1, 0)
	rr := httptest.NewRecorder()

	h.HandleGetSettings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- PUT /api/settings/timezone ---

func TestHandleUpdateTimezone_ValidTimezone(t *testing.T) {
	h, fs, _ := newTestHandlers(t)

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
	h, fs, _ := newTestHandlers(t)

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
	h, fs, _ := newTestHandlers(t)

	fam, err := fs.Create("tz-empty-fam")
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader([]byte("{}")))
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUpdateTimezone_NoFamilyID(t *testing.T) {
	h, _, _ := newTestHandlers(t)

	body, _ := json.Marshal(map[string]string{"timezone": "America/Chicago"})
	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader(body))
	req = testutil.SetRequestContext(req, "parent", 1, 0)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUpdateTimezone_MalformedJSON(t *testing.T) {
	h, fs, _ := newTestHandlers(t)

	fam, err := fs.Create("tz-malformed-fam")
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/settings/timezone", bytes.NewReader([]byte("not json")))
	req = testutil.SetRequestContext(req, "parent", 1, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleUpdateTimezone(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
