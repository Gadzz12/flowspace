package handler

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"todo-app-test/internal/model"
	"todo-app-test/internal/repository"
)

func TestHandlerCollaborationAndInbox(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "handler_collab_test.json")
	repo := repository.NewJSONRepository(tempFile)

	tmpl := template.Must(template.New("dashboard.html").Parse("Dashboard: {{.Username}}, Active: {{.ActivePage}}"))
	template.Must(tmpl.New("login.html").Parse("Login Form"))
	template.Must(tmpl.New("register.html").Parse("Register Form"))
	template.Must(tmpl.New("detail_todo.html").Parse("Detail: {{.Title}}"))

	h := New(repo, tmpl)

	// Buat user 1 (User A)
	userA, _ := repo.CreateUser(model.User{Nama: "User A", Username: "usera", Email: "usera@example.com", Password: "pass"})
	// Buat user 2 (User B)
	userB, _ := repo.CreateUser(model.User{Nama: "User B", Username: "userb", Email: "userb@example.com", Password: "pass"})

	// User A membuat tugas
	repo.CreateTodo(model.Todo{UserID: userA.ID, Title: "Tugas Kolaborasi Penting"}, nil, "")
	todos, _ := repo.GetTodosForUser(userA.ID, userA.Email, "")
	todoID := todos[0].ID

	// 1. User A mengundang User B lewat POST /invitation/send
	form := url.Values{}
	form.Set("todo_id", strconv.Itoa(todoID))
	form.Set("invitee", "userb@example.com")
	form.Set("from_page", "team")

	req := httptest.NewRequest(http.MethodPost, "/invitation/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userA.ID)})
	rec := httptest.NewRecorder()

	h.SendInvitation(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("SendInvitation expected 303 redirect, got %d", rec.Code)
	}

	// 2. User B membuka inbox
	inboxMessages, err := repo.GetInboxMessages(userB.Email)
	if err != nil || len(inboxMessages) != 1 {
		t.Fatalf("User B harus menerima 1 pesan inbox: %v", err)
	}
	inviteID := inboxMessages[0].InviteID

	// 3. User B merespons dengan Terima (ACC) via POST /invitation/respond
	respForm := url.Values{}
	respForm.Set("invite_id", strconv.Itoa(inviteID))
	respForm.Set("action", "accept")

	req = httptest.NewRequest(http.MethodPost, "/invitation/respond", strings.NewReader(respForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userB.ID)})
	rec = httptest.NewRecorder()

	h.RespondInvitation(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("RespondInvitation expected 303 redirect, got %d", rec.Code)
	}

	// 4. Verifikasi User B sekarang menjadi kolaborator aktif
	todoB, found, err := repo.GetTodoByID(todoID, userB.ID, userB.Email)
	if !found || err != nil {
		t.Fatalf("User B harus bisa mengakses tugas setelah ACC: %v", err)
	}
	if len(todoB.Collaborators) == 0 || todoB.Collaborators[0] != userB.Email {
		t.Fatalf("User B harus terdaftar sebagai kolaborator: %+v", todoB.Collaborators)
	}
	if todoB.UserRole != "Member" {
		t.Fatalf("User B role should be Member, got %s", todoB.UserRole)
	}
	todoA, _, _ := repo.GetTodoByID(todoID, userA.ID, userA.Email)
	if todoA.UserRole != "Admin" {
		t.Fatalf("User A role should be Admin, got %s", todoA.UserRole)
	}

	// 5. User B menandai semua inbox dibaca via POST /inbox/read-all
	req = httptest.NewRequest(http.MethodPost, "/inbox/read-all", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userB.ID)})
	rec = httptest.NewRecorder()

	h.MarkAllInboxRead(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("MarkAllInboxRead expected 303 redirect, got %d", rec.Code)
	}

	unreadB, _ := repo.GetUnreadInboxCount(userB.Email)
	if unreadB != 0 {
		t.Fatalf("Unread count User B harus 0 setelah mark all read, got %d", unreadB)
	}

	// 6. Test AttachDocument via POST /todo/attach-document
	docForm := url.Values{}
	docForm.Set("todo_id", strconv.Itoa(todoID))
	docForm.Set("document", "Design_System.fig")
	docForm.Set("from_page", "documents")

	req = httptest.NewRequest(http.MethodPost, "/todo/attach-document", strings.NewReader(docForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userA.ID)})
	rec = httptest.NewRecorder()

	h.AttachDocument(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("AttachDocument expected 303 redirect, got %d", rec.Code)
	}

	// 7. Test /documents redirect
	req = httptest.NewRequest(http.MethodGet, "/documents", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userA.ID)})
	rec = httptest.NewRecorder()

	h.Documents(rec, req)
	if rec.Code != http.StatusSeeOther || !strings.Contains(rec.Header().Get("Location"), "page=documents") {
		t.Fatalf("Documents expected redirect to /?page=documents, got %s", rec.Header().Get("Location"))
	}

	// 8. Test /calendar redirect
	req = httptest.NewRequest(http.MethodGet, "/calendar", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userA.ID)})
	rec = httptest.NewRecorder()

	h.Calendar(rec, req)
	if rec.Code != http.StatusSeeOther || !strings.Contains(rec.Header().Get("Location"), "page=calendar") {
		t.Fatalf("Calendar expected redirect to /?page=calendar, got %s", rec.Header().Get("Location"))
	}

	// 9. Test Dashboard rendering with Calendar & Documents
	req = httptest.NewRequest(http.MethodGet, "/?page=calendar", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userA.ID)})
	rec = httptest.NewRecorder()

	h.Dashboard(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Dashboard expected 200 OK, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Active: calendar") {
		t.Fatalf("Dashboard should render active calendar, got: %s", body)
	}

	// 10. Test User B (Invited Collaborator) tries to Edit Project
	// GET /todo/edit?id=... should redirect to /todo/detail with error "hanya admin yang dapat merubah project"
	req = httptest.NewRequest(http.MethodGet, "/todo/edit?id="+strconv.Itoa(todoID), nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userB.ID)})
	rec = httptest.NewRecorder()
	h.EditTodo(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Collaborator EditTodo GET expected 303 redirect, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "hanya+admin+yang+dapat+merubah+project") {
		t.Fatalf("Collaborator EditTodo GET redirect location should contain error message, got %s", loc)
	}

	// POST /todo/edit?id=... should be rejected with 403 Forbidden
	editData := url.Values{}
	editData.Set("title", "Hacked Title")
	req = httptest.NewRequest(http.MethodPost, "/todo/edit?id="+strconv.Itoa(todoID), strings.NewReader(editData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userB.ID)})
	rec = httptest.NewRecorder()
	h.EditTodo(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Collaborator EditTodo POST expected 403 Forbidden, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "hanya admin yang dapat merubah project") {
		t.Fatalf("Collaborator EditTodo POST response should contain message, got: %s", rec.Body.String())
	}

	// 11. Test User B (Invited Collaborator) tries to Delete Project
	// POST /todo/delete?id=... should be rejected with 403 Forbidden
	req = httptest.NewRequest(http.MethodPost, "/todo/delete?id="+strconv.Itoa(todoID), nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: strconv.Itoa(userB.ID)})
	rec = httptest.NewRecorder()
	h.DeleteTodo(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Collaborator DeleteTodo POST expected 403 Forbidden, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "hanya admin yang dapat merubah project") {
		t.Fatalf("Collaborator DeleteTodo POST response should contain message, got: %s", rec.Body.String())
	}
}
