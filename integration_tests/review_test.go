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
				rec := doRequest(t, "PUT", "/users", "Pugma", `[{"traq_id":"Pugma","role":"manager"},{"traq_id":"aruze_pino","role":"manager"},{"traq_id":"ramdos","role":"assistant"},{"traq_id":"Hokaze","role":"assistant"},{"traq_id":"gUuUnya","role":"assistant"},{"traq_id":"Akira_256","role":"assistant"},{"traq_id":"Synori","role":"assistant"}]`)

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
		t.Run("create reviews by assistant", func(t *testing.T) {

			t.Run("assistant cannot create weight=5 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Hokaze", `{"type": "approve","weight": 5,"comment": "very LGTM"}`)

				expectedStatus := `400 Bad Request`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant can create weight=4 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Hokaze", `{"type": "approve","weight": 4,"comment": "LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":3,"note_id":1,"reviewer":"Hokaze","type":"approve","weight":4,"status":"active","comment":"LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant cannot create weight=-1 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "gUuUnya", `{"type": "approve","weight": -1,"comment": "very little LGTM"}`)

				expectedStatus := `400 Bad Request`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant can create weight=0 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "gUuUnya", `{"type": "approve","weight": 0,"comment": "little LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":4,"note_id":1,"reviewer":"gUuUnya","type":"approve","weight":0,"status":"active","comment":"little LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant can create comment review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Akira_256", `{"type": "comment","weight": 0,"comment": "comment"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":5,"note_id":1,"reviewer":"Akira_256","type":"comment","weight":0,"status":"active","comment":"comment","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant can create cr review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Synori", `{"type": "change_request","weight": 0,"comment": "not LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":6,"note_id":1,"reviewer":"Synori","type":"change_request","weight":0,"status":"active","comment":"not LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant cannot create duplicate review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Synori", `{"type": "approve","weight": 3,"comment": "LGTM"}`)

				expectedStatus := `409 Conflict`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})
		t.Run("create reviews by normal users", func(t *testing.T) {

			t.Run("assistant cannot create weight=1 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "cp20", `{"type": "approve","weight": 1,"comment": "very LGTM"}`)

				expectedStatus := `400 Bad Request`
				expectedBody := ``
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("assistant can create weight=0 review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "kenken", `{"type": "approve","weight": 0,"comment": "LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":7,"note_id":1,"reviewer":"kenken","type":"approve","weight":0,"status":"active","comment":"LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})
		t.Run("invalid create review", func(t *testing.T) {
			t.Run("self review", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "ramdos", `{"type": "approve","weight": 0,"comment": "little LGTM"}`)

				expectedStatus := `201 Created`
				expectedBody := `{"id":8,"note_id":1,"reviewer":"ramdos","type":"approve","weight":0,"status":"active","comment":"little LGTM","created_at":"[TIME]","updated_at":"[TIME]"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
			t.Run("invalid type", func(t *testing.T) {
				rec := doRequest(t, "POST", reviewPath, "Synori", `{"type": "LGTM","weight": 3,"comment": "LGTM"}`)

				expectedStatus := `400 Bad Request`
				expectedBody := `{"error_message":"operation CreateReview: decode request: validate: invalid: type (invalid value: LGTM)"}`
				assert.Equal(t, rec.Result().Status, expectedStatus)
				assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
			})
		})

	})

}
