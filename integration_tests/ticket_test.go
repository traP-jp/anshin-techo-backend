// NOTE: go test -updateを実行することで、スナップショットを更新することができる

package integrationtests

import (
	"strconv"
	"strings"
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
				expectedBody := `{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				ticketID1 = int(unmarshalResponse(t, rec)["id"].(float64))
			})
			t.Run("create a ticket by assistant", func(t *testing.T) {
				rec := doRequest(t, "POST", "/tickets", "ramdos", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
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
				expectedBody := `[{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("get all tickets by assistant", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets", "ramdos", ``)

				expectedStatus := `200 OK`
				expectedBody := `[{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("get all tickets by normal user", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets", "cp20", ``)

				expectedStatus := `200 OK`
				expectedBody := `[{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})
		t.Run("get tickets with filter and sort", func(t *testing.T) {
			body := `{"title":"フィルターテスト","description":"説明","status":"not_planned","assignee":"hoge2"}`
			rec := doRequest(t, "POST", "/tickets", "Pugma", body)
			t.Run("get tickets filtered by status", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets?status=completed", "Pugma", ``)
				expectedStatus := `200 OK`
				expectedBody := `[{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				assert.Assert(t, !strings.Contains(rec.Body.String(), `"title":"フィルターテスト"`))
			})
			t.Run("get tickets filtered by assignee", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets?assignee=hoge", "Pugma", ``)
				expectedStatus := `200 OK`
				expectedBody := `[{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				assert.Assert(t, !strings.Contains(rec.Body.String(), `"title":"フィルターテスト"`))
			})
			t.Run("get tickets sorted by created date", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets?sort=created_desc", "Pugma", ``)
				expectedStatus := `200 OK`
				expectedBody := `[{"id":[ID],"title":"フィルターテスト","description":"説明","assignee":"hoge2","sub_assignees":[],"stakeholders":[],"status":"not_planned","tags":[],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"},{"id":[ID],"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}]`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				assert.Assert(t, strings.Index(rec.Body.String(), `"title":"フィルターテスト"`) < strings.Index(rec.Body.String(), `"title":"タイトル"`))
			})
		})
		t.Run("update ticket", func (t *testing.T) {
			t.Run("update a ticket by manager", func(t *testing.T) {
				body := `{"title":"タイトル2","status":"waiting_review"}`
				rec := doRequest(t, "PATCH", "/tickets/"+strconv.Itoa(ticketID1), "Pugma", body)
				expectedStatus := `200 OK`
				expectedBody := `{"id":[ID],"title":"タイトル2","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"waiting_review","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("update a ticket by assistant", func(t *testing.T) {
				body := `{"title":"タイトル2","status":"waiting_review"}`
				rec := doRequest(t, "PATCH", "/tickets/"+strconv.Itoa(ticketID2), "ramdos", body)
				expectedStatus := `200 OK`
				expectedBody := `{"id":[ID],"title":"タイトル2","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"waiting_review","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("cannot update a ticket by normal user", func(t *testing.T) {
				body := `{"title":"タイトル2","status":"waiting_review"}`
				rec := doRequest(t, "PATCH", "/tickets/"+strconv.Itoa(ticketID2), "cp20", body)
				expectedStatus := `403 Forbidden`
				assert.Equal(t, rec.Result().Status, expectedStatus)
			})
		})
		t.Run("get ticket detail", func(t *testing.T) {
			t.Run("get a ticket detail by manager", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets/"+strconv.Itoa(ticketID1), "Pugma", ``)
				expectedStatus := `200 OK`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Assert(t, strings.Contains(rec.Body.String(), `"notes":[]`))
			})
			t.Run("get a ticket detail with non-existent ID by manager", func(t *testing.T) {
				rec := doRequest(t, "GET", "/tickets/99999", "Pugma", ``)
				expectedStatus := `404 Not Found`
				assert.Equal(t, rec.Result().Status, expectedStatus)
			})
		})
		t.Run("delete ticket", func(t *testing.T) {

			t.Run("delete a ticket by manager", func(t *testing.T) {
				recDelete := doRequest(t, "DELETE", "/tickets/"+strconv.Itoa(ticketID1), "Pugma", ``)

				expectedStatus := `204 No Content`
				expectedBody := ``
				assert.Equal(t, recDelete.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, recDelete.Body.String()), expectedBody)

				recGet := doRequest(t, "GET", "/tickets/"+strconv.Itoa(ticketID1), "Pugma", ``)
				expectedStatusGet := `404 Not Found`
				assert.Equal(t, recGet.Result().Status, expectedStatusGet)
			})
			t.Run("cannot delete a ticket by assistant", func(t *testing.T) {
				rec := doRequest(t, "DELETE", "/tickets/"+strconv.Itoa(ticketID2), "ramdos", ``)

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

	t.Run("get tickets with censored content", func(t *testing.T) {
		body := `{"title": "タイトル!!伏せ字!!","description": "説明!!伏せ字!!","status": "not_planned","assignee": "Pugma"}`
		recPost := doRequest(t, "POST", "/tickets", "Pugma", body)
		ticketID4 := int(unmarshalResponse(t, recPost)["id"].(float64))

		t.Run("check censoring for manager", func (t *testing.T) {
			rec := doRequest(t, "GET", "/tickets/"+strconv.Itoa(ticketID4), "Pugma", ``)
			expectedStatus := `200 OK`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Assert(t, strings.Contains(rec.Body.String(), `"title":"タイトル!!伏せ字!!"`))
			assert.Assert(t, strings.Contains(rec.Body.String(), `"description":"説明!!伏せ字!!"`))
		})

		t.Run("check censoring for assistant", func (t *testing.T) {
			rec := doRequest(t, "GET", "/tickets/"+strconv.Itoa(ticketID4), "ramdos", ``)
			expectedStatus := `200 OK`
			assert.Equal(t, rec.Result().Status, expectedStatus)
			assert.Assert(t, strings.Contains(rec.Body.String(), `"title":"タイトル!!■■■!!"`))
			assert.Assert(t, !strings.Contains(rec.Body.String(), `"title":"タイトル!!伏せ字!!"`))
			assert.Assert(t, strings.Contains(rec.Body.String(), `"description":"説明!!■■■!!"`))
			assert.Assert(t, !strings.Contains(rec.Body.String(), `"description":"説明!!伏せ字!!"`))
		})

	})

}
