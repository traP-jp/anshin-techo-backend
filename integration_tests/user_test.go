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
			rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma","role":"manager"}]`)

			expectedStatus := `200 OK`
			expectedBody := ``
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})

		t.Run("invalid: name is blank", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", "Pugma", `[{"role":"manager"}]`)

			expectedStatus := `400 Bad Request`
			expectedBody := `{"error_message":"operation UsersPut: decode request: decode application/json: callback: invalid: traq_id (field required)"}`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})

		t.Run("invalid: role is blank", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma"}]`)

			expectedStatus := `400 Bad Request`
			expectedBody := `{"error_message":"operation UsersPut: decode request: decode application/json: callback: invalid: role (field required)"}`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})

		t.Run("invalid: role is invalid", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma","role":"not_a_role"}]`)

			expectedStatus := `400 Bad Request`
			expectedBody := `{"error_message":"operation UsersPut: decode request: validate: invalid: [0] (invalid: role (invalid value: not_a_role))"}`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})
	})

	t.Run("get users", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()
			rec := doRequest(t, "GET", "/users", "Pugma", "")

			expectedStatus := `200 OK`
			expectedBody := `[{"traq_id":"Pugma","role":"manager"}]`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
		})
	})
}
