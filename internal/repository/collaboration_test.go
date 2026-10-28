package repository

import (
	"path/filepath"
	"strings"
	"testing"

	"todo-app-test/internal/model"
)

func TestCollaborationInviteAndAcceptWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "collab_test_data.json")

	repo := NewJSONRepository(tempFile)

	// Buat 2 user: Alice (Pemilik tugas) dan Bob (Kolaborator)
	alice, err := repo.CreateUser(model.User{
		Nama:     "Alice Smith",
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Gagal membuat user Alice: %v", err)
	}

	bob, err := repo.CreateUser(model.User{
		Nama:     "Bob Jones",
		Username: "bob",
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Gagal membuat user Bob: %v", err)
	}

	// Alice membuat tugas baru
	todo := model.Todo{
		UserID:      alice.ID,
		Title:       "Proyek Kolaborasi Tim",
		Description: "Tugas bersama Alice dan Bob",
		Deadline:    "2026-12-31T23:59",
	}
	err = repo.CreateTodo(todo, []string{"Checklist 1"}, "")
	if err != nil {
		t.Fatalf("Gagal membuat tugas: %v", err)
	}

	aliceTodos, err := repo.GetTodosForUser(alice.ID, alice.Email, "")
	if err != nil || len(aliceTodos) == 0 {
		t.Fatalf("Gagal mengambil tugas Alice: %v", err)
	}
	todoID := aliceTodos[0].ID

	// 1. Sebelum diundang: Bob tidak memiliki akses ke tugas ini
	_, found, err := repo.GetTodoByID(todoID, bob.ID, bob.Email)
	if err == nil && found {
		t.Fatal("Bob seharusnya belum memiliki akses ke tugas Alice")
	}

	// 2. Alice mengundang Bob ke kolaborasi
	inv, err := repo.InviteCollaborator(todoID, alice.ID, bob.Email)
	if err != nil {
		t.Fatalf("Gagal mengundang Bob: %v", err)
	}
	if inv.Status != "pending" {
		t.Fatalf("Expected status pending, got %s", inv.Status)
	}

	// Cek bahwa undangan pending muncul pada tugas
	pendingInvs, err := repo.GetPendingInvitationsForTodo(todoID)
	if err != nil || len(pendingInvs) != 1 {
		t.Fatalf("Expected 1 pending invitation, got %d", len(pendingInvs))
	}
	if pendingInvs[0].InviteeEmail != bob.Email {
		t.Fatalf("Invitee email mismatch: %s vs %s", pendingInvs[0].InviteeEmail, bob.Email)
	}

	// 3. Bob belum ACC: Bob masih TIDAK boleh memiliki akses ke tugas
	_, _, err = repo.GetTodoByID(todoID, bob.ID, bob.Email)
	if err == nil {
		t.Fatal("Sebelum Bob ACC undangan, Bob tidak boleh dapat mengakses tugas")
	}

	// Cek bahwa pesan muncul di Inbox Bob dan unread count bertambah
	bobUnread, err := repo.GetUnreadInboxCount(bob.Email)
	if err != nil || bobUnread != 1 {
		t.Fatalf("Expected Bob unread count 1, got %d", bobUnread)
	}

	bobMessages, err := repo.GetInboxMessages(bob.Email)
	if err != nil || len(bobMessages) != 1 {
		t.Fatalf("Expected 1 inbox message for Bob, got %d", len(bobMessages))
	}
	if bobMessages[0].Type != "invite" || bobMessages[0].InviteID != inv.ID {
		t.Fatalf("Pesan inbox Bob tidak sesuai: %+v", bobMessages[0])
	}

	// 4. Bob menerima (ACC) undangan melalui Inbox
	err = repo.RespondInvitation(inv.ID, bob.ID, bob.Email, true)
	if err != nil {
		t.Fatalf("Bob gagal ACC undangan: %v", err)
	}

	// 5. Setelah Bob ACC: Bob resmi menjadi kolaborator aktif!
	todoDetail, found, err := repo.GetTodoByID(todoID, bob.ID, bob.Email)
	if err != nil || !found {
		t.Fatalf("Setelah ACC, Bob harus bisa mengakses tugas: %v", err)
	}
	if len(todoDetail.Collaborators) != 1 || todoDetail.Collaborators[0] != bob.Email {
		t.Fatalf("Kolaborator pada tugas harus mencakup Bob: %+v", todoDetail.Collaborators)
	}

	// Tugas Alice kini harus muncul di dashboard Bob
	bobTodos, err := repo.GetTodosForUser(bob.ID, bob.Email, "")
	if err != nil || len(bobTodos) != 1 {
		t.Fatalf("Tugas kolaborasi harus muncul di daftar tugas Bob: %d tugas ditemukan", len(bobTodos))
	}

	// Alice menerima pesan konfirmasi di inbox-nya bahwa Bob telah menerima (ACC) undangan
	aliceMessages, err := repo.GetInboxMessages(alice.Email)
	if err != nil || len(aliceMessages) == 0 {
		t.Fatal("Alice harus menerima notifikasi konfirmasi di Inbox")
	}
	if aliceMessages[0].Type != "accepted" {
		t.Fatalf("Tipe pesan Alice harus 'accepted', got %s", aliceMessages[0].Type)
	}

	// Team members view harus memuat Alice (Admin) dan Bob (Member)
	team, err := repo.GetTeamMembers(alice.Email)
	if err != nil || len(team) < 2 {
		t.Fatalf("Expected minimal 2 team members, got %d", len(team))
	}
	var aliceFound, bobFound bool
	for _, m := range team {
		if strings.EqualFold(m.Email, alice.Email) && m.Role == "Admin" {
			aliceFound = true
		}
		if strings.EqualFold(m.Email, bob.Email) && m.Role == "Member" {
			bobFound = true
		}
	}
	if !aliceFound || !bobFound {
		t.Fatalf("Alice should have role Admin and Bob should have role Member: %+v", team)
	}
}

func TestCollaborationInviteRejectWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "collab_reject_test_data.json")

	repo := NewJSONRepository(tempFile)

	alice, _ := repo.CreateUser(model.User{Nama: "Alice", Username: "alice2", Email: "alice2@example.com", Password: "123"})
	charlie, _ := repo.CreateUser(model.User{Nama: "Charlie", Username: "charlie", Email: "charlie@example.com", Password: "123"})

	todo := model.Todo{UserID: alice.ID, Title: "Proyek Alpha", Deadline: "2026-12-31T23:59"}
	repo.CreateTodo(todo, nil, "")
	aliceTodos, _ := repo.GetTodosForUser(alice.ID, alice.Email, "")
	todoID := aliceTodos[0].ID

	// Alice mengundang Charlie via username
	inv, err := repo.InviteCollaborator(todoID, alice.ID, "charlie")
	if err != nil {
		t.Fatalf("Gagal mengundang Charlie via username: %v", err)
	}

	// Charlie menolak undangan
	err = repo.RespondInvitation(inv.ID, charlie.ID, charlie.Email, false)
	if err != nil {
		t.Fatalf("Gagal menolak undangan: %v", err)
	}

	// Pastikan Charlie TIDAK menjadi kolaborator
	_, _, err = repo.GetTodoByID(todoID, charlie.ID, charlie.Email)
	if err == nil {
		t.Fatal("Charlie tidak boleh memiliki akses ke tugas setelah menolak undangan")
	}

	// Alice menerima pesan penolakan
	aliceMessages, _ := repo.GetInboxMessages(alice.Email)
	if len(aliceMessages) == 0 || aliceMessages[0].Type != "rejected" {
		t.Fatalf("Alice harus menerima notifikasi penolakan (type=rejected)")
	}
}

func TestAttachDocumentAndGetTodos(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "doc_test_data.json")

	repo := NewJSONRepository(tempFile)
	user, _ := repo.CreateUser(model.User{Nama: "David", Username: "david", Email: "david@example.com", Password: "123"})

	// 1. Buat todo tanpa dokumen
	todo := model.Todo{UserID: user.ID, Title: "Tugas Tanpa Dokumen"}
	err := repo.CreateTodo(todo, nil, "")
	if err != nil {
		t.Fatalf("Gagal membuat todo: %v", err)
	}

	todos, _ := repo.GetTodosForUser(user.ID, user.Email, "")
	if len(todos) != 1 || todos[0].Document != "" {
		t.Fatalf("Document harus kosong jika tidak diinput: %v", todos[0].Document)
	}

	// 2. Lampirkan dokumen ke todo
	err = repo.AttachDocument(todos[0].ID, user.ID, "Project_Brief.pdf")
	if err != nil {
		t.Fatalf("Gagal melampirkan dokumen: %v", err)
	}

	// 3. Verifikasi dokumen terlampir
	updatedTodo, found, err := repo.GetTodoByID(todos[0].ID, user.ID, user.Email)
	if !found || err != nil {
		t.Fatalf("Gagal mengambil todo: %v", err)
	}
	if updatedTodo.Document != "Project_Brief.pdf" {
		t.Fatalf("Expected document 'Project_Brief.pdf', got '%s'", updatedTodo.Document)
	}
}
