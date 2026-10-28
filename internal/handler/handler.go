package handler

import (
	"html/template"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"todo-app-test/internal/repository"
)

// Handler mengelola HTTP request untuk aplikasi Todo
type Handler struct {
	repo repository.Repository
	tmpl *template.Template
}

// New membuat instance Handler baru
func New(repo repository.Repository, tmpl *template.Template) *Handler {
	return &Handler{
		repo: repo,
		tmpl: tmpl,
	}
}

// RegisterRoutes mendaftarkan seluruh endpoint aplikasi ke router
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/", h.Dashboard)
	mux.HandleFunc("/settings", h.Settings)
	mux.HandleFunc("/user/preferences", h.UpdateUserPreferences)
	mux.HandleFunc("/todo/create", h.CreateTodo)
	mux.HandleFunc("/todo/detail", h.DetailTodo)
	mux.HandleFunc("/todo/edit", h.EditTodo)
	mux.HandleFunc("/todo/delete", h.DeleteTodo)

	// Halaman & Fitur Workspace (Inbox, Team, Document, Calendar)
	mux.HandleFunc("/inbox", h.Inbox)
	mux.HandleFunc("/team", h.Team)
	mux.HandleFunc("/documents", h.Documents)
	mux.HandleFunc("/calendar", h.Calendar)
	mux.HandleFunc("/todo/attach-document", h.AttachDocument)
	mux.HandleFunc("/invitation/send", h.SendInvitation)
	mux.HandleFunc("/invitation/respond", h.RespondInvitation)
	mux.HandleFunc("/inbox/read-all", h.MarkAllInboxRead)

	// File statis (CSS, JS) untuk dashboard dan halaman lainnya
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
}

// getLoggedInUserID mengambil user ID dari cookie sesi
func (h *Handler) getLoggedInUserID(r *http.Request) int {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		return 0
	}

	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return 0
	}

	return id
}

// emailIsValid memeriksa kevalidan format alamat email
func emailIsValid(email string) bool {
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
