// NOTE: go test -updateを実行することで、スナップショットを更新することができる

package integrationtests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestConfig(t *testing.T) {
	truncateAllTables(t)

	t.Run("prepare users", func(t *testing.T) {
		rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma","role":"manager"},{"traq_id":"ramdos","role":"assistant"}]`)

		expectedStatus := `200 OK`
		expectedBody := ``
		assert.Equal(t, rec.Result().Status, expectedStatus)
		assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
	})

	t.Run("get default config as manager", func(t *testing.T) {
		rec := doRequest(t, "GET", "/config", "Pugma", "")

		expectedStatus := `200 OK`
		expectedBody := `{"reminder_interval":{"overdue_day":[],"notesent_hour":0},"revise_prompt":""}`
		assert.Equal(t, rec.Result().Status, expectedStatus)
		assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
	})

	t.Run("forbid non manager", func(t *testing.T) {
		rec := doRequest(t, "GET", "/config", "ramdos", "")

		expectedStatus := `403 Forbidden`
		expectedBody := ``
		assert.Equal(t, rec.Result().Status, expectedStatus)
		assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
	})

	t.Run("update config as manager", func(t *testing.T) {
		body := `{"reminder_interval":{"overdue_day":[1,3,7],"notesent_hour":12},"revise_prompt":"Please revise."}`
		rec := doRequest(t, "POST", "/config", "Pugma", body)

		expectedStatus := `200 OK`
		expectedBody := `{"reminder_interval":{"overdue_day":[1,3,7],"notesent_hour":12},"revise_prompt":"Please revise."}`
		assert.Equal(t, rec.Result().Status, expectedStatus)
		assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
	})

	t.Run("update config forbidden for assistant", func(t *testing.T) {
		body := `{"reminder_interval":{"overdue_day":[2],"notesent_hour":6},"revise_prompt":"no"}`
		rec := doRequest(t, "POST", "/config", "ramdos", body)

		expectedStatus := `403 Forbidden`
		expectedBody := ``
		assert.Equal(t, rec.Result().Status, expectedStatus)
		assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
	})

	t.Run("get updated config", func(t *testing.T) {
		rec := doRequest(t, "GET", "/config", "Pugma", "")

		expectedStatus := `200 OK`
		expectedBody := `{"reminder_interval":{"overdue_day":[1,3,7],"notesent_hour":12},"revise_prompt":"Please revise."}`
		assert.Equal(t, rec.Result().Status, expectedStatus)
		assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
	})
}
