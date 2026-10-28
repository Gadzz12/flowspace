package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"todo-app-test/internal/model"
)

// wibLocation zona waktu Indonesia Barat (UTC+7 / Jam Indonesia)
var wibLocation = time.FixedZone("WIB", 7*3600)

// parseDateTime mencoba mengurai string tanggal dan waktu dalam zona waktu Indonesia (WIB)
func parseDateTime(val string) (time.Time, bool) {
	val = strings.TrimSpace(val)
	if val == "" || val == "-" {
		return time.Time{}, false
	}
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, f := range formats {
		if tm, err := time.ParseInLocation(f, val, wibLocation); err == nil {
			return tm.In(wibLocation), true
		}
	}
	return time.Time{}, false
}

// isDeadlineOverdue memeriksa apakah tanggal deadline telah lewat dari waktu sekarang (WIB)
func isDeadlineOverdue(deadlineStr string) bool {
	tm, ok := parseDateTime(deadlineStr)
	if !ok {
		return false
	}
	return time.Now().In(wibLocation).After(tm)
}

// Dashboard menampilkan daftar tugas dan grafik statistik
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	userID := h.getLoggedInUserID(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, exists := h.repo.GetUserByID(userID)
	if !exists {
		http.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	statusFilter := r.URL.Query().Get("status")

	// Ambil semua tugas untuk statistik lengkap
	allTodos, err := h.repo.GetTodosForUser(userID, user.Email, "")
	if err != nil {
		log.Println("Gagal mengambil daftar tugas:", err)
		http.Error(w, "Gagal mengambil data tugas", http.StatusInternalServerError)
		return
	}

	countDone, countOngoing, countPending, total, donePercent, ongoingPercent, pendingPercent, ongoingEndPercent := h.repo.CalculateStats(allTodos)

	// Konversi ke TodoItem dengan label tampilan dan kelompokkan per status
	var pendingTodos, ongoingTodos, doneTodos []model.TodoItem
	upcomingDeadlines := 0
	overdueCount := 0
	now := time.Now()

	for _, t := range allTodos {
		item := model.TodoItem{Todo: t}

		// Label tanggal dibuat
		if t.CreatedAt != "" {
			if tm, ok := parseDateTime(t.CreatedAt); ok {
				item.CreatedLabel = tm.Format("02 Jan 2006")
			} else {
				item.CreatedLabel = t.CreatedAt
			}
		} else {
			item.CreatedLabel = "-"
		}

		// Label deadline, hitung overdue, dan hitung deadline mendatang (7 hari)
		if t.Deadline != "" {
			if tm, ok := parseDateTime(t.Deadline); ok {
				item.DeadlineLabel = tm.Format("02 Jan 2006 15:04")
				if t.Status != "sudah" {
					if now.After(tm) {
						item.IsOverdue = true
						overdueCount++
					} else {
						diff := tm.Sub(now)
						if diff >= 0 && diff.Hours() <= 168 {
							upcomingDeadlines++
						}
					}
				}
			} else {
				item.DeadlineLabel = t.Deadline
			}
		} else {
			item.DeadlineLabel = "-"
		}

		// Prioritas tugas (High Priority, Medium Priority, Low Priority)
		prio := strings.ToLower(strings.TrimSpace(t.Priority))
		switch prio {
		case "high", "tinggi":
			item.Priority = "High Priority"
			item.PriorityClass = "high"
		case "low", "rendah":
			item.Priority = "Low Priority"
			item.PriorityClass = "low"
		default:
			item.Priority = "Medium Priority"
			item.PriorityClass = "medium"
		}

		// Inisial kolaborator
		for _, c := range t.Collaborators {
			if c != "" {
				item.CollaboratorInitials = append(item.CollaboratorInitials, strings.ToUpper(string(c[0])))
			}
		}

		switch t.Status {
		case "sudah":
			doneTodos = append(doneTodos, item)
		case "sedang":
			ongoingTodos = append(ongoingTodos, item)
		default:
			pendingTodos = append(pendingTodos, item)
		}
	}

	// Hitung jumlah anggota tim (user + kolaborator unik)
	collaboratorEmails := make(map[string]bool)
	for _, t := range allTodos {
		for _, c := range t.Collaborators {
			if c != "" && c != user.Email {
				collaboratorEmails[c] = true
			}
		}
	}
	teamCount := 1 + len(collaboratorEmails)

	// Inisial avatar pengguna
	initial := ""
	if user.Nama != "" {
		initial = strings.ToUpper(string(user.Nama[0]))
	}

	activePage := strings.TrimSpace(r.URL.Query().Get("page"))
	if activePage == "" {
		activePage = "dashboard"
	}

	unreadInbox, _ := h.repo.GetUnreadInboxCount(user.Email)
	inboxMessages, _ := h.repo.GetInboxMessages(user.Email)
	teamMembers, _ := h.repo.GetTeamMembers(user.Email)
	simpleTodos, _ := h.repo.GetUserSimpleTodos(userID, user.Email)
	flashMsg := r.URL.Query().Get("msg")
	flashErr := r.URL.Query().Get("error")

	if len(teamMembers) > 0 {
		teamCount = len(teamMembers)
	}

	// Kumpulkan dokumen dari seluruh tugas yang memiliki berkas lampiran
	var documents []model.DocumentItem
	for _, t := range allTodos {
		doc := strings.TrimSpace(t.Document)
		if doc != "" {
			ext := "FILE"
			dotIdx := strings.LastIndex(doc, ".")
			if dotIdx != -1 && dotIdx < len(doc)-1 {
				ext = strings.ToUpper(doc[dotIdx+1:])
			}
			typeClass := "other"
			switch ext {
			case "PDF":
				typeClass = "pdf"
			case "DOC", "DOCX":
				typeClass = "doc"
			case "XLS", "XLSX", "CSV":
				typeClass = "sheet"
			}
			dateStr := t.CreatedAt
			if tm, ok := parseDateTime(t.CreatedAt); ok {
				dateStr = tm.Format("02 Jan 2006")
			}
			documents = append(documents, model.DocumentItem{
				TodoID:    t.ID,
				TodoTitle: t.Title,
				FileName:  doc,
				Ext:       ext,
				TypeClass: typeClass,
				Date:      dateStr,
			})
		}
	}

	// Kumpulkan jadwal deadline untuk tampilan Calendar
	var calendarEvents []model.CalendarEvent
	for _, t := range allTodos {
		statusLabel := "Menunggu"
		switch t.Status {
		case "sudah":
			statusLabel = "Selesai"
		case "sedang":
			statusLabel = "Sedang Dikerjakan"
		}

		if t.Deadline != "" {
			if tm, ok := parseDateTime(t.Deadline); ok {
				calendarEvents = append(calendarEvents, model.CalendarEvent{
					TodoID:      t.ID,
					Title:       t.Title,
					Date:        tm.Format("2006-01-02"),
					Time:        tm.Format("15:04") + " WIB",
					Status:      t.Status,
					StatusClass: t.Status,
					StatusLabel: statusLabel,
					Progress:    t.Progress,
					Priority:    t.Priority,
					IsDeadline:  true,
				})
				continue
			}
		}
		// Fallback ke tanggal pembuatan jika deadline tidak ditentukan
		if t.CreatedAt != "" {
			if tm, ok := parseDateTime(t.CreatedAt); ok {
				calendarEvents = append(calendarEvents, model.CalendarEvent{
					TodoID:      t.ID,
					Title:       t.Title,
					Date:        tm.Format("2006-01-02"),
					Time:        tm.Format("15:04") + " WIB",
					Status:      t.Status,
					StatusClass: t.Status,
					StatusLabel: statusLabel,
					Progress:    t.Progress,
					Priority:    t.Priority,
					IsDeadline:  false,
				})
			}
		}
	}
	calendarJSONBytes, _ := json.Marshal(calendarEvents)

	pageData := model.DashboardData{
		Username:           user.Username,
		Nama:               user.Nama,
		Email:              user.Email,
		Initial:            initial,
		Filter:             statusFilter,
		CountDone:          countDone,
		CountOngoing:       countOngoing,
		CountPending:       countPending,
		TotalVisible:       total,
		DonePercent:        donePercent,
		OngoingPercent:     ongoingPercent,
		PendingPercent:     pendingPercent,
		OngoingEndPercent:  ongoingEndPercent,
		TeamCount:          teamCount,
		UpcomingDeadlines:  upcomingDeadlines,
		OverdueCount:       overdueCount,
		PendingTodos:       pendingTodos,
		OngoingTodos:       ongoingTodos,
		DoneTodos:          doneTodos,
		ActivePage:         activePage,
		UnreadInboxCount:   unreadInbox,
		InboxMessages:      inboxMessages,
		TeamMembers:        teamMembers,
		UserTodosSimple:    simpleTodos,
		Documents:          documents,
		CalendarEvents:     calendarEvents,
		CalendarEventsJSON: string(calendarJSONBytes),
		Avatar:             user.Avatar,
		Language:           user.Language,
		FlashMessage:       flashMsg,
		FlashError:         flashErr,
	}

	if err := h.tmpl.ExecuteTemplate(w, "dashboard.html", pageData); err != nil {
		log.Println("Template Error:", err)
	}
}

// CreateTodo menangani pembuatan tugas baru
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	userID := h.getLoggedInUserID(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseMultipartForm(10 << 20)

		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		deadline := strings.TrimSpace(r.FormValue("deadline"))
		docName := strings.TrimSpace(r.FormValue("document"))

		file, header, fErr := r.FormFile("document_file")
		if fErr == nil && header != nil {
			defer file.Close()
			docName = header.Filename
		}

		if title == "" {
			http.Error(w, "Judul tugas wajib diisi", http.StatusBadRequest)
			return
		}

		if deadline != "" {
			if tm, ok := parseDateTime(deadline); ok {
				if time.Now().After(tm) {
					http.Error(w, "Peringatan: Tanggal deadline tidak boleh melebihi/kurang dari tanggal dibuat (waktu telah lewat). Harap tentukan tanggal mendatang.", http.StatusBadRequest)
					return
				}
			}
		}

		priority := strings.ToLower(strings.TrimSpace(r.FormValue("priority")))
		if priority == "" {
			priority = "medium"
		}
		if priority != "high" && priority != "low" && priority != "medium" {
			switch priority {
			case "tinggi":
				priority = "high"
			case "rendah":
				priority = "low"
			default:
				priority = "medium"
			}
		}

		newTodo := model.Todo{
			UserID:      userID,
			Title:       title,
			Description: description,
			Deadline:    deadline,
			Document:    docName,
			Priority:    priority,
		}

		subTitles := r.Form["sub_titles"]
		collabEmail := r.FormValue("collaborator_email")

		if err := h.repo.CreateTodo(newTodo, subTitles, collabEmail); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, exists := h.repo.GetUserByID(userID)
	initial := "U"
	if exists {
		if len(user.Nama) > 0 {
			initial = strings.ToUpper(string([]rune(user.Nama)[0]))
		} else if len(user.Username) > 0 {
			initial = strings.ToUpper(string([]rune(user.Username)[0]))
		}
	}

	allTodos, _ := h.repo.GetTodosForUser(userID, user.Email, "")
	countDone, countOngoing, countPending, total, _, _, _, _ := h.repo.CalculateStats(allTodos)

	unreadInbox, _ := h.repo.GetUnreadInboxCount(user.Email)

	pageData := model.DashboardData{
		Username:         user.Username,
		Nama:             user.Nama,
		Email:            user.Email,
		Initial:          initial,
		Avatar:           user.Avatar,
		Language:         user.Language,
		CountPending:     countPending,
		CountOngoing:     countOngoing,
		CountDone:        countDone,
		TotalVisible:     total,
		UnreadInboxCount: unreadInbox,
		ActivePage:       "create_todo",
	}

	if err := h.tmpl.ExecuteTemplate(w, "create_todo.html", pageData); err != nil {
		log.Println("Gagal render create_todo.html:", err)
	}
}

// DetailTodo menampilkan informasi rincian tugas dan checklist sub-task
func (h *Handler) DetailTodo(w http.ResponseWriter, r *http.Request) {
	userID := h.getLoggedInUserID(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	todoID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || todoID <= 0 {
		http.Error(w, "ID tugas tidak valid", http.StatusBadRequest)
		return
	}

	user, exists := h.repo.GetUserByID(userID)
	if !exists {
		http.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	todo, found, err := h.repo.GetTodoByID(todoID, userID, user.Email)
	if !found {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Akses ditolak", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if !todo.IsOwner {
			http.Error(w, "hanya admin yang dapat merubah project", http.StatusForbidden)
			return
		}

		_ = r.ParseMultipartForm(10 << 20)

		switch r.FormValue("action") {
		case "add_collaborator":
			email := strings.TrimSpace(r.FormValue("collaborator_email"))
			if err := h.repo.AddCollaborator(todoID, userID, email); err != nil {
				http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&error=%s", todoID, url.QueryEscape(err.Error())), http.StatusSeeOther)
				return
			}
			msg := fmt.Sprintf("Undangan kolaborasi berhasil dikirim ke %s. Menunggu persetujuan (ACC) via inbox.", email)
			http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&msg=%s", todoID, url.QueryEscape(msg)), http.StatusSeeOther)
			return

		case "attach_document":
			doc := strings.TrimSpace(r.FormValue("document"))
			file, header, fErr := r.FormFile("document_file")
			if fErr == nil && header != nil {
				defer file.Close()
				doc = header.Filename
			}
			if doc == "" {
				http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&error=%s", todoID, url.QueryEscape("Nama dokumen tidak boleh kosong")), http.StatusSeeOther)
				return
			}
			if err := h.repo.AttachDocument(todoID, userID, doc); err != nil {
				http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&error=%s", todoID, url.QueryEscape(err.Error())), http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&msg=%s", todoID, url.QueryEscape("Dokumen berhasil dilampirkan ke tugas")), http.StatusSeeOther)
			return

		case "toggle_sub":
			subID, err := strconv.Atoi(r.FormValue("sub_id"))
			if err != nil {
				http.Error(w, "ID sub-task tidak valid", http.StatusBadRequest)
				return
			}
			if err := h.repo.ToggleSubTodo(todoID, subID, userID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		case "update_status":
			status := r.FormValue("status")
			if err := h.repo.UpdateTodoStatus(todoID, status, userID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, "Aksi tidak valid", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d", todoID), http.StatusSeeOther)
		return
	}

	if todo.Status != "sudah" && isDeadlineOverdue(todo.Deadline) {
		todo.IsOverdue = true
	}

	todo.FlashMessage = r.URL.Query().Get("msg")
	todo.FlashError = r.URL.Query().Get("error")

	if todo.Document != "" {
		ext := "FILE"
		dotIdx := strings.LastIndex(todo.Document, ".")
		if dotIdx != -1 && dotIdx < len(todo.Document)-1 {
			ext = strings.ToUpper(todo.Document[dotIdx+1:])
		}
		typeClass := "other"
		switch ext {
		case "PDF":
			typeClass = "pdf"
		case "DOC", "DOCX":
			typeClass = "doc"
		case "XLS", "XLSX", "CSV":
			typeClass = "sheet"
		}
		todo.DocumentExt = ext
		todo.DocumentTypeClass = typeClass
	}

	if err := h.tmpl.ExecuteTemplate(w, "detail_todo.html", todo); err != nil {
		log.Println("Gagal render detail_todo.html:", err)
	}
}

// EditTodo menangani perubahan data tugas
func (h *Handler) EditTodo(w http.ResponseWriter, r *http.Request) {
	userID := h.getLoggedInUserID(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	todoID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || todoID <= 0 {
		http.Error(w, "ID tugas tidak valid", http.StatusBadRequest)
		return
	}

	user, exists := h.repo.GetUserByID(userID)
	if !exists {
		http.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	todo, found, err := h.repo.GetTodoByID(todoID, userID, user.Email)
	if !found {
		http.NotFound(w, r)
		return
	}
	if !todo.IsOwner {
		if r.Method == http.MethodGet {
			http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&error=%s", todoID, url.QueryEscape("hanya admin yang dapat merubah project")), http.StatusSeeOther)
			return
		}
		http.Error(w, "hanya admin yang dapat merubah project", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" {
			http.Error(w, "Judul tugas wajib diisi", http.StatusBadRequest)
			return
		}

		description := strings.TrimSpace(r.FormValue("description"))
		deadline := strings.TrimSpace(r.FormValue("deadline"))
		status := r.FormValue("status")
		priority := strings.ToLower(strings.TrimSpace(r.FormValue("priority")))
		if priority == "" {
			priority = todo.Priority
		}
		if priority != "high" && priority != "low" && priority != "medium" {
			switch priority {
			case "tinggi":
				priority = "high"
			case "rendah":
				priority = "low"
			default:
				priority = "medium"
			}
		}

		if deadline != todo.Deadline && status != "sudah" && deadline != "" {
			if tm, ok := parseDateTime(deadline); ok {
				if time.Now().After(tm) {
					http.Error(w, "Peringatan: Perubahan tanggal deadline tidak boleh di masa lampau. Harap tentukan tanggal mendatang.", http.StatusBadRequest)
					return
				}
			}
		}

		if err := h.repo.UpdateTodo(todoID, userID, title, description, deadline, status, priority); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d", todoID), http.StatusSeeOther)
		return
	}

	initial := "U"
	if len(user.Nama) > 0 {
		initial = strings.ToUpper(string([]rune(user.Nama)[0]))
	} else if len(user.Username) > 0 {
		initial = strings.ToUpper(string([]rune(user.Username)[0]))
	}

	allTodos, _ := h.repo.GetTodosForUser(userID, user.Email, "")
	countDone, countOngoing, countPending, total, _, _, _, _ := h.repo.CalculateStats(allTodos)
	unreadInbox, _ := h.repo.GetUnreadInboxCount(user.Email)

	pageData := model.DashboardData{
		Username:         user.Username,
		Nama:             user.Nama,
		Email:            user.Email,
		Initial:          initial,
		Avatar:           user.Avatar,
		Language:         user.Language,
		CountPending:     countPending,
		CountOngoing:     countOngoing,
		CountDone:        countDone,
		TotalVisible:     total,
		UnreadInboxCount: unreadInbox,
		ActivePage:       "edit_todo",
		Todo:             todo,
	}

	if err := h.tmpl.ExecuteTemplate(w, "edit_todo.html", pageData); err != nil {
		log.Println("Gagal render edit_todo.html:", err)
	}
}

// DeleteTodo menangani penghapusan tugas beserta sub-task dan kolaboratornya
func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	userID := h.getLoggedInUserID(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	todoID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || todoID <= 0 {
		http.Error(w, "ID tugas tidak valid", http.StatusBadRequest)
		return
	}

	user, exists := h.repo.GetUserByID(userID)
	if !exists {
		http.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	todo, found, err := h.repo.GetTodoByID(todoID, userID, user.Email)
	if !found {
		http.NotFound(w, r)
		return
	}
	if !todo.IsOwner {
		if r.Method == http.MethodGet {
			http.Redirect(w, r, fmt.Sprintf("/todo/detail?id=%d&error=%s", todoID, url.QueryEscape("hanya admin yang dapat merubah project")), http.StatusSeeOther)
			return
		}
		http.Error(w, "hanya admin yang dapat merubah project", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		initial := "U"
		if len(user.Nama) > 0 {
			initial = strings.ToUpper(string([]rune(user.Nama)[0]))
		} else if len(user.Username) > 0 {
			initial = strings.ToUpper(string([]rune(user.Username)[0]))
		}

		allTodos, _ := h.repo.GetTodosForUser(userID, user.Email, "")
		countDone, countOngoing, countPending, total, _, _, _, _ := h.repo.CalculateStats(allTodos)
		unreadInbox, _ := h.repo.GetUnreadInboxCount(user.Email)

		pageData := model.DashboardData{
			Username:         user.Username,
			Nama:             user.Nama,
			Email:            user.Email,
			Initial:          initial,
			Avatar:           user.Avatar,
			Language:         user.Language,
			CountPending:     countPending,
			CountOngoing:     countOngoing,
			CountDone:        countDone,
			TotalVisible:     total,
			UnreadInboxCount: unreadInbox,
			ActivePage:       "delete_todo",
			Todo:             todo,
		}

		if err := h.tmpl.ExecuteTemplate(w, "delete_todo.html", pageData); err != nil {
			log.Println("Gagal render delete_todo.html:", err)
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	if err := h.repo.DeleteTodo(todoID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
