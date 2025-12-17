// NOTE: go test -updateを実行することで、スナップショットを更新することができる

package integrationtests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestUser(t *testing.T) {
	t.Run("create an user", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", `[{"traq_id":"ramdos","role":"manager"}]`)

			expectedStatus := `200 OK`
			expectedBody := ``
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})

		t.Run("invalid: name is blank", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", `{"role":"manager"}`)

			expectedStatus := `401 Unauthorized`
			expectedBody := `{"error_message":"operation UsersPut: decode request: decode application/json: \"[\" expected: unexpected byte 123 '{' at 0"}`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})

		t.Run("invalid: role is blank", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", `{"traq_id":"ramdos"}`)

			expectedStatus := `400 Bad Request`
			expectedBody := `{"error_message":"operation UsersPut: decode request: decode application/json: \"[\" expected: unexpected byte 123 '{' at 0"}`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})

		t.Run("invalid: role is invalid", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", `{"traq_id":"ramdos","role":"not_a_role"}`)

			expectedStatus := `400 Bad Request`
			expectedBody := `{"error_message":"operation UsersPut: decode request: decode application/json: \"[\" expected: unexpected byte 123 '{' at 0"}`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})
	})

	t.Run("get users", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "GET", "/users", "")

			expectedStatus := `200 OK`
			expectedBody := `[{"traq_id":"ramdos","role":"manager"}]`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})
	})
}
