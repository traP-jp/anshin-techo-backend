// NOTE: go test -updateを実行することで、スナップショットを更新することができる

package integrationtests

import (
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

func TestTicket(t *testing.T) {
	truncateAllTables(t)
	
	t.Run("create ticket", func(t *testing.T) {
		t.Run("prepare create tickets", func(t *testing.T) {
			t.Run("prepare: create users", func(t *testing.T) {
				rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma","role":"manager"},{"traq_id":"aruze_pino","role":"manager"},{"traq_id":"ramdos","role":"assistant"},{"traq_id":"Hokaze","role":"assistant"},{"traq_id":"gUuUnya","role":"assistant"},{"traq_id":"Akira_256","role":"assistant"},{"traq_id":"Synori","role":"assistant"}]`)

				expectedStatus := `200 OK`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})
		var ticketID1 int
		var ticketID2 int
		t.Run("create ticket", func(t *testing.T) {

			t.Run("create a ticket by manager", func(t *testing.T) {
				rec := doRequest(t, "POST", "/tickets", "Pugma", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":2,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				ticketID1 = int(unmarshalResponse(t, rec)["id"].(float64))
			})
			t.Run("create a ticket by assistant", func(t *testing.T) {
				rec := doRequest(t, "POST", "/tickets", "ramdos", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":3,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				ticketID2 = int(unmarshalResponse(t, rec)["id"].(float64))
			})
			t.Run("cannot create a ticket by normal user", func(t *testing.T) {
				rec := doRequest(t, "POST", "/tickets", "cp20", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `403 Forbidden`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})
		t.Run("get all tickets", func(t *testing.T) {
			t.Run("get all tickets by manager", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets", "Pugma", ``)

				expectedStatus := `200 OK`
				expectedBody := `[{"id":1,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":2,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":3,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("get all tickets by assistant", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets", "ramdos", ``)

				expectedStatus := `200 OK`
				expectedBody := `[{"id":1,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":2,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":3,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("get all tickets by normal user", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets", "cp20", ``)

				expectedStatus := `200 OK`
				expectedBody := `[{"id":1,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":2,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":3,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})
		t.Run("delete ticket", func(t *testing.T) {

			t.Run("delete a ticket by manager", func(t *testing.T) {
				rec := doRequest(t, "DELETE", "/tickets/"+strconv.Itoa(ticketID1), "Pugma", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `204 No Content`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("cannot delete a ticket by assistant", func(t *testing.T) {
				rec := doRequest(t, "DELETE", "/tickets/"+strconv.Itoa(ticketID2), "ramdos", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `403 Forbidden`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("cannot delete a ticket by normal user", func(t *testing.T) {
				rec := doRequest(t, "DELETE", "/tickets/"+strconv.Itoa(ticketID2), "cp20", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `403 Forbidden`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})

	})

}
