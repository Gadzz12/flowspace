package handler

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"todo-app-test/internal/model"
	"todo-app-test/internal/repository"
)

func TestHandlerAuthAndDashboard(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_handler_data.json")

	repo := repository.NewJSONRepository(tempFile)

	// Inisialisasi dummy template sederhana untuk testing
	tmpl := template.Must(template.New("dashboard.html").Parse("Dashboard: {{.Username}}"))
	template.Must(tmpl.New("login.html").Parse("Login Form"))
	template.Must(tmpl.New("register.html").Parse("Register Form"))
	template.Must(tmpl.New("settings.html").Parse("Settings: {{.Username}}, {{.Email}}"))
	template.Must(tmpl.New("edit_todo.html").Parse("Edit: {{.Todo.Title}}"))
	template.Must(tmpl.New("delete_todo.html").Parse("Delete: {{.Todo.Title}}"))

	h := New(repo, tmpl)

	// 1. Test Register
	form := url.Values{}
	form.Set("nama", "Budi Santoso")
	form.Set("username", "budi")
	form.Set("email", "budi@example.com")
	form.Set("password", "secret123")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.Register(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Register expected status 303, got %d", rec.Code)
	}

	// 2. Test Login sukses
	loginForm := url.Values{}
	loginForm.Set("identifier", "budi")
	loginForm.Set("password", "secret123")

	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Login expected status 303, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "user_id" {
			authCookie = c
			break
		}
	}
	if authCookie == nil || authCookie.Value != "1" {
		t.Fatalf("Cookie user_id tidak sesuai: %v", authCookie)
	}

	// 3. Test Dashboard tanpa login (harus redirect ke /login)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	h.Dashboard(rec, req)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
		t.Fatalf("Dashboard tanpa login harus redirect ke /login, got code %d to %s", rec.Code, rec.Header().Get("Location"))
	}

	// 4. Test Dashboard dengan cookie login
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.Dashboard(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Dashboard dengan login expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Dashboard: budi") {
		t.Fatalf("Body tidak memuat username budi: %s", rec.Body.String())
	}

	// 5. Test Settings tanpa login (harus redirect ke /login)
	req = httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec = httptest.NewRecorder()
	h.Settings(rec, req)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
		t.Fatalf("Settings tanpa login harus redirect ke /login, got code %d", rec.Code)
	}

	// 6. Test Settings dengan login (HTTP 200 dan render template)
	req = httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.Settings(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Settings dengan login expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Settings: budi, budi@example.com") {
		t.Fatalf("Body settings tidak memuat data yang sesuai: %s", rec.Body.String())
	}

	// 7. Test UpdateUserPreferences
	prefForm := url.Values{}
	prefForm.Set("avatar", "<svg>test-avatar</svg>")
	prefForm.Set("language", "zh")

	req = httptest.NewRequest(http.MethodPost, "/user/preferences", strings.NewReader(prefForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.UpdateUserPreferences(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateUserPreferences expected status 200, got %d", rec.Code)
	}
	savedUser, _ := repo.GetUserByID(1)
	if savedUser.Avatar != "<svg>test-avatar</svg>" || savedUser.Language != "zh" {
		t.Fatalf("Saved user preferences mismatch: %v", savedUser)
	}

	// 8. Test CreateTodo with Priority High
	createForm := url.Values{}
	createForm.Set("title", "Handler High Priority Task")
	createForm.Set("description", "Testing task priority")
	createForm.Set("deadline", "2026-12-31T23:59")
	createForm.Set("priority", "high")

	req = httptest.NewRequest(http.MethodPost, "/todo/create", strings.NewReader(createForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.CreateTodo(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("CreateTodo expected status 303, got %d", rec.Code)
	}

	userTodos, _ := repo.GetTodosForUser(1, "budi@example.com", "")
	if len(userTodos) == 0 || userTodos[0].Priority != "high" {
		t.Fatalf("Created todo priority expected 'high', got %v", userTodos)
	}

	// 9. Test EditTodo with Priority Low
	editForm := url.Values{}
	editForm.Set("title", "Handler Low Priority Task")
	editForm.Set("description", "Testing updated task priority")
	editForm.Set("deadline", "2026-12-31T23:59")
	editForm.Set("status", "sedang")
	editForm.Set("priority", "low")

	req = httptest.NewRequest(http.MethodPost, "/todo/edit?id=1", strings.NewReader(editForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.EditTodo(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("EditTodo expected status 303, got %d", rec.Code)
	}

	editedTodo, _, _ := repo.GetTodoByID(1, 1, "budi@example.com")
	if editedTodo.Priority != "low" {
		t.Fatalf("Edited todo priority expected 'low', got %q", editedTodo.Priority)
	}

	// 10. Test EditTodo GET (renders edit_todo.html with dashboard shell data)
	req = httptest.NewRequest(http.MethodGet, "/todo/edit?id=1", nil)
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.EditTodo(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("EditTodo GET expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Edit: Handler Low Priority Task") {
		t.Fatalf("EditTodo GET body mismatch: %s", rec.Body.String())
	}

	// 11. Test DeleteTodo GET (renders delete_todo.html confirmation page)
	req = httptest.NewRequest(http.MethodGet, "/todo/delete?id=1", nil)
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.DeleteTodo(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("DeleteTodo GET expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Delete: Handler Low Priority Task") {
		t.Fatalf("DeleteTodo GET body mismatch: %s", rec.Body.String())
	}

	// 12. Test DeleteTodo POST (deletes todo and redirects to /)
	req = httptest.NewRequest(http.MethodPost, "/todo/delete?id=1", nil)
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	h.DeleteTodo(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("DeleteTodo POST expected status 303, got %d", rec.Code)
	}
	_, found, _ := repo.GetTodoByID(1, 1, "budi@example.com")
	if found {
		t.Fatal("Todo ID 1 should be deleted after POST /todo/delete")
	}
}

func TestActualTemplatesParsing(t *testing.T) {
	tmpl, err := template.ParseGlob("../../templates/*.html")
	if err != nil {
		t.Fatalf("Gagal mem-parse templates/*.html: %v", err)
	}

	// Pastikan template dashboard, settings, detail, edit, create, delete dapat dieksekusi tanpa error sintaks
	if tmpl.Lookup("dashboard.html") == nil {
		t.Fatal("dashboard.html tidak ditemukan")
	}
	if tmpl.Lookup("settings.html") == nil {
		t.Fatal("settings.html tidak ditemukan")
	}
	if tmpl.Lookup("detail_todo.html") == nil {
		t.Fatal("detail_todo.html tidak ditemukan")
	}
	if tmpl.Lookup("create_todo.html") == nil {
		t.Fatal("create_todo.html tidak ditemukan")
	}
	if tmpl.Lookup("edit_todo.html") == nil {
		t.Fatal("edit_todo.html tidak ditemukan")
	}
	if tmpl.Lookup("delete_todo.html") == nil {
		t.Fatal("delete_todo.html tidak ditemukan")
	}

	// Test render dashboard.html dengan dummy data
	dashboardData := model.DashboardData{
		Username:          "skydev",
		Nama:              "Muhammad Yusril Ananta",
		Email:             "skydev@example.com",
		Initial:           "M",
		CountDone:         2,
		CountOngoing:      1,
		CountPending:      1,
		TotalVisible:      4,
		DonePercent:       50,
		OngoingPercent:    25,
		PendingPercent:    25,
		OngoingEndPercent: 75,
		OverdueCount:      1,
		PendingTodos: []model.TodoItem{
			{
				Todo: model.Todo{
					ID:       1,
					Title:    "Tugas Menunggu",
					Deadline: "2026-10-10 10:00",
				},
				DeadlineLabel: "10 Oct 2026 10:00",
				IsOverdue:     true,
			},
		},
		OngoingTodos: []model.TodoItem{
			{
				Todo: model.Todo{
					ID:       2,
					Title:    "Tugas Berjalan",
					Deadline: "2026-10-30 10:00",
				},
				DeadlineLabel: "30 Oct 2026 10:00",
				IsOverdue:     false,
			},
		},
	}

	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, "dashboard.html", dashboardData); err != nil {
		t.Fatalf("Gagal render dashboard.html: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Diagram Progres Tugas") {
		t.Fatal("dashboard.html tidak memuat 'Diagram Progres Tugas'")
	}
	if !strings.Contains(output, "Peringatan Deadline:") {
		t.Fatal("dashboard.html tidak memuat peringatan deadline banner")
	}
	if !strings.Contains(output, "⚠️ Terlewat") {
		t.Fatal("dashboard.html tidak memuat badge '⚠️ Terlewat'")
	}

	// Test render settings.html
	settingsData := model.SettingsData{
		Username:     "skydev",
		Nama:         "Muhammad Yusril Ananta",
		Email:        "skydev@example.com",
		Initial:      "M",
		JoinedLabel:  "27 Oct 2026",
		TotalVisible: 4,
		CountPending: 1,
		CountOngoing: 1,
		CountDone:    2,
	}

	buf.Reset()
	if err := tmpl.ExecuteTemplate(&buf, "settings.html", settingsData); err != nil {
		t.Fatalf("Gagal render settings.html: %v", err)
	}
	if !strings.Contains(buf.String(), "Pengaturan") {
		t.Fatal("settings.html tidak memuat judul Pengaturan")
	}
}
