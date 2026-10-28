package handler

import (
	"log"
	"net/http"
	"strings"

	"todo-app-test/internal/model"
)

// Settings menangani tampilan halaman pengaturan profil dan preferensi workspace
func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
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

	allTodos, err := h.repo.GetTodosForUser(userID, user.Email, "")
	if err != nil {
		allTodos = []model.Todo{}
	}

	countDone, countOngoing, countPending, total, _, _, _, _ := h.repo.CalculateStats(allTodos)

	initial := ""
	if user.Nama != "" {
		initial = strings.ToUpper(string(user.Nama[0]))
	}

	joinedLabel := "-"
	if !user.CreatedAt.IsZero() {
		joinedLabel = user.CreatedAt.Format("02 Jan 2006")
	}

	lang := user.Language
	if lang == "" {
		lang = "en"
	}

	data := model.SettingsData{
		Username:     user.Username,
		Nama:         user.Nama,
		Email:        user.Email,
		Initial:      initial,
		Avatar:       user.Avatar,
		Language:     lang,
		JoinedLabel:  joinedLabel,
		TotalVisible: total,
		CountPending: countPending,
		CountOngoing: countOngoing,
		CountDone:    countDone,
	}

	if err := h.tmpl.ExecuteTemplate(w, "settings.html", data); err != nil {
		log.Println("Gagal render settings.html:", err)
	}
}

// UpdateUserPreferences menangani pembaruan foto avatar profil dan preferensi bahasa
func (h *Handler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := h.getLoggedInUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	avatar := strings.TrimSpace(r.FormValue("avatar"))
	language := strings.TrimSpace(r.FormValue("language"))

	if err := h.repo.UpdateUserPreferences(userID, avatar, language); err != nil {
		log.Println("UpdateUserPreferences error:", err)
	}

	if language != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "app_lang",
			Value:    language,
			Path:     "/",
			HttpOnly: false,
		})
	}

	redirectURL := r.FormValue("redirect")
	if redirectURL != "" {
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
