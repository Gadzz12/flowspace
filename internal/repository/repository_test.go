package repository

import (
	"os"
	"path/filepath"
	"testing"

	"todo-app-test/internal/model"
)

func TestJSONRepositoryOperations(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_data.json")

	repo := NewJSONRepository(tempFile)

	// Test 1: Buat user baru
	newUser := model.User{
		Nama:     "Tester User",
		Username: "tester",
		Email:    "tester@example.com",
		Password: "password123",
	}

	createdUser, err := repo.CreateUser(newUser)
	if err != nil {
		t.Fatalf("CreateUser gagal: %v", err)
	}
	if createdUser.ID != 1 {
		t.Fatalf("Expected ID 1, got %d", createdUser.ID)
	}

	// Test 2: UserExists
	if !repo.UserExists("tester", "random@example.com") {
		t.Fatal("User tester seharusnya terdeteksi ada")
	}
	if repo.UserExists("unknown", "unknown@example.com") {
		t.Fatal("User unknown seharusnya tidak terdeteksi")
	}

	// Test 3: GetUserByUsernameOrEmail
	user, ok := repo.GetUserByUsernameOrEmail("tester")
	if !ok || user.ID != 1 {
		t.Fatal("Gagal mengambil user via username")
	}

	user, ok = repo.GetUserByUsernameOrEmail("tester@example.com")
	if !ok || user.ID != 1 {
		t.Fatal("Gagal mengambil user via email")
	}

	// Test 4: CreateTodo dengan sub-task
	todo := model.Todo{
		UserID:      1,
		Title:       "Test Todo 1",
		Description: "Testing Deskripsi",
		Deadline:    "2026-12-31T23:59",
	}
	subTitles := []string{"Subtask 1", "Subtask 2"}
	err = repo.CreateTodo(todo, subTitles, "")
	if err != nil {
		t.Fatalf("CreateTodo gagal: %v", err)
	}

	// Test 5: GetTodosForUser
	todos, err := repo.GetTodosForUser(1, "tester@example.com", "")
	if err != nil || len(todos) != 1 {
		t.Fatalf("GetTodosForUser gagal: %v", err)
	}
	if len(todos[0].SubTodos) != 2 {
		t.Fatalf("Expected 2 sub-todos, got %d", len(todos[0].SubTodos))
	}
	if todos[0].Progress != 0 {
		t.Fatalf("Expected progress 0, got %d", todos[0].Progress)
	}

	// Test 6: Toggle sub-task
	subID := todos[0].SubTodos[0].ID
	err = repo.ToggleSubTodo(todos[0].ID, subID, 1)
	if err != nil {
		t.Fatalf("ToggleSubTodo gagal: %v", err)
	}

	updatedTodo, found, err := repo.GetTodoByID(todos[0].ID, 1, "tester@example.com")
	if !found || err != nil {
		t.Fatalf("GetTodoByID gagal: %v", err)
	}
	if updatedTodo.Progress != 50 {
		t.Fatalf("Expected progress 50%%, got %d%%", updatedTodo.Progress)
	}
	if updatedTodo.Status != "sedang" {
		t.Fatalf("Expected status 'sedang', got %q", updatedTodo.Status)
	}

	// Test 7: CalculateStats
	countDone, countOngoing, countPending, total, donePercent, ongoingPercent, pendingPercent, ongoingEndPercent := repo.CalculateStats([]model.Todo{updatedTodo})
	if total != 1 || countOngoing != 1 || countDone != 0 || countPending != 0 {
		t.Fatalf("Stats tidak sesuai: total=%d, ongoing=%d", total, countOngoing)
	}
	if ongoingPercent != 100 || ongoingEndPercent != 100 || donePercent != 0 || pendingPercent != 0 {
		t.Fatalf("Persentase tidak sesuai")
	}

	// Test 8: UpdateUserPreferences
	err = repo.UpdateUserPreferences(1, "<svg>avatar</svg>", "ru")
	if err != nil {
		t.Fatalf("UpdateUserPreferences gagal: %v", err)
	}
	u, ok := repo.GetUserByID(1)
	if !ok || u.Avatar != "<svg>avatar</svg>" || u.Language != "ru" {
		t.Fatalf("UpdateUserPreferences tidak tersimpan dengan benar: avatar=%q, lang=%q", u.Avatar, u.Language)
	}

	// Test 9: Create and Update Priority Level
	prioTodo := model.Todo{
		UserID:   1,
		Title:    "High Priority Task",
		Priority: "high",
		Deadline: "2026-12-31T23:59",
	}
	if err := repo.CreateTodo(prioTodo, nil, ""); err != nil {
		t.Fatalf("CreateTodo with priority high failed: %v", err)
	}
	allUserTodos, _ := repo.GetTodosForUser(1, "tester@example.com", "")
	targetID := allUserTodos[len(allUserTodos)-1].ID
	prioFetched, found, err := repo.GetTodoByID(targetID, 1, "tester@example.com")
	if !found || err != nil || prioFetched.Priority != "high" {
		t.Fatalf("Expected priority 'high', got %q", prioFetched.Priority)
	}
	if err := repo.UpdateTodo(targetID, 1, "Low Priority Task", "", "2026-12-31T23:59", "belum", "low"); err != nil {
		t.Fatalf("UpdateTodo with priority low failed: %v", err)
	}
	prioFetched, found, err = repo.GetTodoByID(targetID, 1, "tester@example.com")
	if !found || err != nil || prioFetched.Priority != "low" {
		t.Fatalf("Expected updated priority 'low', got %q", prioFetched.Priority)
	}

	// Cleanup verified by TempDir
	_ = os.Remove(tempFile)
}
