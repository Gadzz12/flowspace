package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"todo-app-test/internal/model"
)

// loginPageData menyimpan data yang ditampilkan pada halaman login
type loginPageData struct {
	Error string
}

// Login menangani proses autentikasi masuk pengguna
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		identifier := strings.TrimSpace(r.FormValue("identifier"))
		password := r.FormValue("password")

		user, exists := h.repo.GetUserByUsernameOrEmail(identifier)
		if exists && user.Password == password {
			cookie := &http.Cookie{
				Name:     "user_id",
				Value:    strconv.Itoa(user.ID),
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}

			// "Ingat saya" membuat sesi bertahan 30 hari.
			if r.FormValue("remember") != "" {
				cookie.MaxAge = 30 * 24 * 60 * 60
			}

			http.SetCookie(w, cookie)

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	data := loginPageData{}
	if r.URL.Query().Get("error") != "" {
		data.Error = "Username/email atau password salah. Silakan coba lagi."
	}

	if err := h.tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
		log.Println("Gagal render login.html:", err)
	}
}

// Register menangani pendaftaran akun baru
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		nama := strings.TrimSpace(r.FormValue("nama"))
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")

		if nama == "" || username == "" || password == "" || !emailIsValid(email) {
			http.Error(w, "Data registrasi tidak valid", http.StatusBadRequest)
			return
		}

		if h.repo.UserExists(username, email) {
			http.Error(w, "Username atau email sudah terdaftar", http.StatusBadRequest)
			return
		}

		newUser := model.User{
			Nama:     nama,
			Username: username,
			Email:    email,
			Password: password,
		}

		if _, err := h.repo.CreateUser(newUser); err != nil {
			http.Error(w, "Gagal menyimpan akun pengguna", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "register.html", nil); err != nil {
		log.Println("Gagal render register.html:", err)
	}
}

// Logout menghapus sesi login pengguna
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
