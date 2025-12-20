// NOTE: go test -updateを実行することで、スナップショットを更新することができる

package integrationtests

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestReview(t *testing.T) {
	t.Run("create reviews", func(t *testing.T) {
		var reviewPath string
		t.Run("prepare create reviews", func(t *testing.T) {
			t.Run("prepare: create users", func(t *testing.T) {
				rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma","role":"manager"},{"traq_id":"aruze_pino","role":"manager"},{"traq_id":"ramdos","role":"assistant"},{"traq_id":"Hokaze","role":"assistant"}]`)

				expectedStatus := `200 OK`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			var ticketID int
			t.Run("prepare: create a ticket", func(t *testing.T) {
				rec := doRequest(t, "POST", "/tickets", "Pugma", `{"title": "タイトル","description": "説明","status": "completed","assignee": "hoge","sub_assignees": ["fuga"],"stakeholders": ["piyo"],"due": "2025-12-17","tags": ["タグ"]}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":1,"title":"タイトル","description":"説明","assignee":"hoge","sub_assignees":["fuga"],"stakeholders":["piyo"],"status":"completed","tags":["タグ"],"due":"2025-12-17","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				ticketID = int(unmarshalResponse(t, rec)["id"].(float64))
			})
			var noteID int
			t.Run("prepare: create a note", func(t *testing.T) {
				rec := doRequest(t, "POST", "/tickets/"+fmt.Sprintf("%v", ticketID)+"/notes", "ramdos", `{"type": "outgoing","content": "毎々お世話になっております。","mention_notification": false}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":1,"ticket_id":1,"type":"outgoing","author":"ramdos","content":"毎々お世話になっております。","reviews":[],"created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
				noteID = int(unmarshalResponse(t, rec)["id"].(float64))
			})
			reviewPath = "/tickets/" + fmt.Sprintf("%v", ticketID) + "/notes/" + fmt.Sprintf("%v", noteID) + "/reviews"
		})
		t.Run("create reviews by manager", func(t *testing.T) {
			t.Run("manager can create weight=5 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Pugma", `{"type": "approve","weight": 5,"comment": "LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":1,"note_id":1,"reviewer":"Pugma","type":"approve","weight":5,"status":"active","comment":"LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})

			t.Run("manager cannot create weight=6 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Pugma", `{"type": "approve","weight": 6,"comment": "very LGTM"}`)

				expectedStatus := `400 Bad Request`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("manager cannot create weight=-1 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "aruze_pino", `{"type": "approve","weight": -1,"comment": "very little LGTM"}`)

				expectedStatus := `400 Bad Request`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("manager can create weight=0 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "aruze_pino", `{"type": "approve","weight": 0,"comment": "little LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":2,"note_id":1,"reviewer":"aruze_pino","type":"approve","weight":0,"status":"active","comment":"little LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})

	})

}
