package integrationtests

import (
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"gotest.tools/v3/assert"
)

var (
	uuidRegexp = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	timeRegexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z`)
	idRegexp   = regexp.MustCompile(`id":\d+`)
)

func escapeSnapshot(t *testing.T, s string) string {
	t.Helper()

	s = strings.Trim(s, "\n")
	s = uuidRegexp.ReplaceAllString(s, "[UUID]")
	s = timeRegexp.ReplaceAllString(s, "[TIME]")
	s = idRegexp.ReplaceAllString(s, `id":[ID]`)

	return s
}

func truncateAllTables(t *testing.T) {
	t.Helper()

	stmts := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE note_review_assignees",
		"TRUNCATE TABLE reviews",
		"TRUNCATE TABLE notes",
		"TRUNCATE TABLE ticket_sub_assignees",
		"TRUNCATE TABLE ticket_stakeholders",
		"TRUNCATE TABLE ticket_tags",
		"TRUNCATE TABLE tickets",
		"TRUNCATE TABLE users",
		"SET FOREIGN_KEY_CHECKS = 1",
	}

	for _, stmt := range stmts {
		_, err := globalDB.Exec(stmt)
		assert.NilError(t, err)
	}
}

func doRequest(t *testing.T, method, path string, user string, bodystr string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(bodystr))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Forwarded-User", user)
	rec := httptest.NewRecorder()

	globalServer.ServeHTTP(rec, req)

	return rec
}

func unmarshalResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	v := map[string]any{}
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &v))

	return v
}
