package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"todo-app-test/internal/model"

	"gorm.io/gorm"
)

// GORMRepository implementasi Repository memakai GORM (PostgreSQL / Neon / Supabase)
type GORMRepository struct {
	db *gorm.DB
}

// NewGORMRepository membuat instance repository berbasis GORM
func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) GetUserByID(id int) (model.User, bool) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return model.User{}, false
	}
	return user, true
}

func (r *GORMRepository) GetUserByUsernameOrEmail(identifier string) (model.User, bool) {
	var user model.User
	identifier = strings.TrimSpace(identifier)
	err := r.db.Where("username = ? OR LOWER(email) = LOWER(?)", identifier, identifier).First(&user).Error
	if err != nil {
		return model.User{}, false
	}
	return user, true
}

func (r *GORMRepository) UserExists(username, email string) bool {
	var count int64
	r.db.Model(&model.User{}).Where("LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)", strings.TrimSpace(username), strings.TrimSpace(email)).Count(&count)
	return count > 0
}

func (r *GORMRepository) CreateUser(user model.User) (model.User, error) {
	err := r.db.Create(&user).Error
	return user, err
}

func (r *GORMRepository) GetTodosForUser(userID int, userEmail string, statusFilter string) ([]model.Todo, error) {
	var todos []model.Todo

	query := r.db.Model(&model.Todo{}).
		Where("user_id = ? OR id IN (SELECT todo_id FROM collaborators WHERE LOWER(email) = LOWER(?))", userID, strings.TrimSpace(userEmail))

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Order("id asc").Find(&todos).Error; err != nil {
		return nil, err
	}

	for i := range todos {
		todos[i].IsOwner = (todos[i].UserID == userID)

		var subs []model.SubTodo
		r.db.Where("todo_id = ?", todos[i].ID).Order("id asc").Find(&subs)
		todos[i].SubTodos = subs

		totalSub := len(subs)
		completed := 0
		for _, s := range subs {
			if s.IsCompleted {
				completed++
			}
		}

		if totalSub > 0 {
			todos[i].Progress = completed * 100 / totalSub
		} else if todos[i].Status == "sudah" {
			todos[i].Progress = 100
		} else {
			todos[i].Progress = 0
		}

		var collabEmails []string
		r.db.Model(&model.Collaborator{}).Where("todo_id = ?", todos[i].ID).Pluck("email", &collabEmails)
		todos[i].Collaborators = collabEmails
	}

	return todos, nil
}

func (r *GORMRepository) GetTodoByID(id int, currentUserID int, userEmail string) (model.Todo, bool, error) {
	var todo model.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Todo{}, false, nil
		}
		return model.Todo{}, false, err
	}

	isOwner := (todo.UserID == currentUserID)

	var collabCount int64
	r.db.Model(&model.Collaborator{}).Where("todo_id = ? AND LOWER(email) = LOWER(?)", id, strings.TrimSpace(userEmail)).Count(&collabCount)
	isCollab := collabCount > 0

	if !isOwner && !isCollab {
		return model.Todo{}, false, errors.New("akses ditolak")
	}

	todo.IsOwner = isOwner
	if isOwner {
		todo.UserRole = "Admin"
	} else {
		todo.UserRole = "Member"
	}
	var ownerUser model.User
	if err := r.db.First(&ownerUser, todo.UserID).Error; err == nil {
		todo.OwnerName = ownerUser.Nama
		todo.OwnerEmail = ownerUser.Email
	}

	var subs []model.SubTodo
	r.db.Where("todo_id = ?", id).Order("id asc").Find(&subs)
	todo.SubTodos = subs

	totalSub := len(subs)
	completed := 0
	for _, s := range subs {
		if s.IsCompleted {
			completed++
		}
	}

	if totalSub > 0 {
		todo.Progress = completed * 100 / totalSub
	} else if todo.Status == "sudah" {
		todo.Progress = 100
	} else {
		todo.Progress = 0
	}

	var collabEmails []string
	r.db.Model(&model.Collaborator{}).Where("todo_id = ?", id).Pluck("email", &collabEmails)
	todo.Collaborators = collabEmails

	var pendingInvs []model.Invitation
	r.db.Where("todo_id = ? AND status = ?", id, "pending").Find(&pendingInvs)
	todo.PendingInvitations = pendingInvs

	return todo, true, nil
}

func (r *GORMRepository) CreateTodo(todo model.Todo, subTitles []string, collaboratorEmail string) error {
	todo.CreatedAt = time.Now().Format("2006-01-02 15:04")
	todo.Status = "belum"
	if strings.TrimSpace(todo.Priority) == "" {
		todo.Priority = "medium"
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		cleanTodo := todo
		cleanTodo.SubTodos = nil
		cleanTodo.Collaborators = nil
		cleanTodo.CollabRecords = nil
		if err := tx.Create(&cleanTodo).Error; err != nil {
			return err
		}

		for _, subTitle := range subTitles {
			subTitle = strings.TrimSpace(subTitle)
			if subTitle == "" {
				continue
			}
			sub := model.SubTodo{
				TodoID:      cleanTodo.ID,
				Title:       subTitle,
				IsCompleted: false,
			}
			if err := tx.Create(&sub).Error; err != nil {
				return err
			}
		}

		collaboratorEmail = strings.TrimSpace(collaboratorEmail)
		if collaboratorEmail != "" {
			if err := r.addCollaboratorTx(tx, cleanTodo.ID, cleanTodo.UserID, collaboratorEmail); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GORMRepository) UpdateTodo(id int, userID int, title, description, deadline, status string, priority ...string) error {
	var todo model.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	if todo.UserID != userID {
		return errors.New("hanya admin yang dapat merubah project")
	}

	todo.Title = title
	todo.Description = description
	todo.Deadline = deadline
	if status == "belum" || status == "sedang" || status == "sudah" {
		todo.Status = status
	}

	if len(priority) > 0 && strings.TrimSpace(priority[0]) != "" {
		todo.Priority = strings.TrimSpace(priority[0])
	}

	if todo.Status == "sudah" {
		todo.CompletedAt = time.Now().Format("2006-01-02 15:04")
	} else {
		todo.CompletedAt = ""
	}

	return r.db.Save(&todo).Error
}

func (r *GORMRepository) DeleteTodo(id int, userID int) error {
	var todo model.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	if todo.UserID != userID {
		return errors.New("hanya admin yang dapat merubah project")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("todo_id = ?", id).Delete(&model.SubTodo{})
		tx.Where("todo_id = ?", id).Delete(&model.Collaborator{})
		return tx.Delete(&todo).Error
	})
}

func (r *GORMRepository) ToggleSubTodo(todoID int, subID int, userID int) error {
	var todo model.Todo
	if err := r.db.First(&todo, todoID).Error; err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	if todo.UserID != userID {
		return errors.New("hanya pemilik yang dapat mengubah tugas")
	}

	var sub model.SubTodo
	if err := r.db.Where("id = ? AND todo_id = ?", subID, todoID).First(&sub).Error; err != nil {
		return errors.New("sub-task tidak ditemukan")
	}

	sub.IsCompleted = !sub.IsCompleted
	if err := r.db.Save(&sub).Error; err != nil {
		return err
	}

	var subs []model.SubTodo
	r.db.Where("todo_id = ?", todoID).Find(&subs)

	total := len(subs)
	completed := 0
	for _, s := range subs {
		if s.IsCompleted {
			completed++
		}
	}

	if total > 0 {
		todo.Progress = completed * 100 / total
	}

	if total > 0 && completed == total {
		todo.Status = "sudah"
		todo.CompletedAt = time.Now().Format("2006-01-02 15:04")
	} else if completed > 0 {
		todo.Status = "sedang"
		todo.CompletedAt = ""
	} else {
		todo.Status = "belum"
		todo.CompletedAt = ""
	}

	return r.db.Save(&todo).Error
}

func (r *GORMRepository) UpdateTodoStatus(todoID int, status string, userID int) error {
	if status != "belum" && status != "sedang" && status != "sudah" {
		return errors.New("status tidak valid")
	}

	var todo model.Todo
	if err := r.db.First(&todo, todoID).Error; err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	if todo.UserID != userID {
		return errors.New("hanya pemilik yang dapat mengubah status tugas")
	}

	todo.Status = status
	if status == "sudah" {
		todo.Progress = 100
		todo.CompletedAt = time.Now().Format("2006-01-02 15:04")
	} else {
		todo.CompletedAt = ""
	}

	return r.db.Save(&todo).Error
}

func (r *GORMRepository) AttachDocument(todoID int, userID int, document string) error {
	var todo model.Todo
	if err := r.db.First(&todo, todoID).Error; err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	var user model.User
	r.db.First(&user, userID)

	var collabCount int64
	if user.Email != "" {
		r.db.Model(&model.Collaborator{}).Where("todo_id = ? AND LOWER(email) = LOWER(?)", todoID, user.Email).Count(&collabCount)
	}

	if todo.UserID != userID && collabCount == 0 {
		return errors.New("hanya pemilik atau kolaborator tugas yang dapat melampirkan dokumen")
	}

	return r.db.Model(&todo).Update("document", strings.TrimSpace(document)).Error
}

func (r *GORMRepository) AddCollaborator(todoID int, ownerID int, email string) error {
	_, err := r.InviteCollaborator(todoID, ownerID, email)
	return err
}

func (r *GORMRepository) addCollaboratorTx(tx *gorm.DB, todoID int, ownerID int, email string) error {
	_, err := r.inviteCollaboratorTx(tx, todoID, ownerID, email)
	return err
}

func (r *GORMRepository) InviteCollaborator(todoID int, inviterID int, inviteeEmailOrUsername string) (model.Invitation, error) {
	var inv model.Invitation
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		inv, err = r.inviteCollaboratorTx(tx, todoID, inviterID, inviteeEmailOrUsername)
		return err
	})
	return inv, err
}

func (r *GORMRepository) inviteCollaboratorTx(tx *gorm.DB, todoID int, inviterID int, inviteeInput string) (model.Invitation, error) {
	inviteeInput = strings.TrimSpace(inviteeInput)
	if inviteeInput == "" {
		return model.Invitation{}, errors.New("email atau username kolaborator wajib diisi")
	}

	var todo model.Todo
	if err := tx.First(&todo, todoID).Error; err != nil {
		return model.Invitation{}, errors.New("tugas tidak ditemukan")
	}

	var inviter model.User
	if err := tx.First(&inviter, inviterID).Error; err != nil {
		return model.Invitation{}, errors.New("pengguna pengundang tidak valid")
	}

	isOwner := todo.UserID == inviterID
	var collabCount int64
	tx.Model(&model.Collaborator{}).Where("todo_id = ? AND LOWER(email) = LOWER(?)", todoID, inviter.Email).Count(&collabCount)
	if !isOwner && collabCount == 0 {
		return model.Invitation{}, errors.New("hanya pemilik atau kolaborator tugas yang dapat mengundang anggota baru")
	}

	var invitee model.User
	if err := tx.Where("LOWER(email) = LOWER(?) OR LOWER(username) = LOWER(?)", inviteeInput, inviteeInput).First(&invitee).Error; err != nil {
		return model.Invitation{}, errors.New("pengguna belum terdaftar. Pastikan email/username sudah benar dan terdaftar")
	}

	if invitee.ID == inviterID || strings.EqualFold(invitee.Email, inviter.Email) {
		return model.Invitation{}, errors.New("anda tidak dapat mengundang diri sendiri")
	}

	if invitee.ID == todo.UserID {
		return model.Invitation{}, errors.New("pengguna tersebut adalah pemilik tugas ini")
	}

	var activeCollabCount int64
	tx.Model(&model.Collaborator{}).Where("todo_id = ? AND LOWER(email) = LOWER(?)", todoID, invitee.Email).Count(&activeCollabCount)
	if activeCollabCount > 0 {
		return model.Invitation{}, errors.New("pengguna tersebut sudah menjadi kolaborator aktif pada tugas ini")
	}

	var pendingCount int64
	tx.Model(&model.Invitation{}).Where("todo_id = ? AND LOWER(invitee_email) = LOWER(?) AND status = ?", todoID, invitee.Email, "pending").Count(&pendingCount)
	if pendingCount > 0 {
		return model.Invitation{}, errors.New("undangan kolaborasi untuk pengguna ini sudah dikirim dan sedang menunggu persetujuan (ACC)")
	}

	now := time.Now().Format("2006-01-02 15:04")
	invitation := model.Invitation{
		TodoID:       todoID,
		TodoTitle:    todo.Title,
		InviterID:    inviter.ID,
		InviterName:  inviter.Nama,
		InviterEmail: inviter.Email,
		InviteeEmail: invitee.Email,
		Status:       "pending",
		CreatedAt:    now,
	}
	if err := tx.Create(&invitation).Error; err != nil {
		return model.Invitation{}, err
	}

	inboxMsg := model.InboxMessage{
		UserID:    invitee.ID,
		UserEmail: invitee.Email,
		Sender:    inviter.Nama,
		Title:     "Undangan Kolaborasi",
		Message:   fmt.Sprintf("%s mengundang Anda untuk berkolaborasi pada tugas '%s'. Menunggu persetujuan Anda.", inviter.Nama, todo.Title),
		Type:      "invite",
		InviteID:  invitation.ID,
		TodoID:    todoID,
		IsRead:    false,
		CreatedAt: now,
	}
	if err := tx.Create(&inboxMsg).Error; err != nil {
		return model.Invitation{}, err
	}

	return invitation, nil
}

func (r *GORMRepository) RespondInvitation(inviteID int, userID int, userEmail string, accept bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var inv model.Invitation
		if err := tx.First(&inv, inviteID).Error; err != nil {
			return errors.New("undangan tidak ditemukan")
		}

		if !strings.EqualFold(inv.InviteeEmail, userEmail) {
			return errors.New("anda tidak memiliki hak akses untuk merespons undangan ini")
		}

		if inv.Status != "pending" {
			return fmt.Errorf("undangan sudah di-%s sebelumnya", inv.Status)
		}

		var responder model.User
		responderName := userEmail
		if err := tx.First(&responder, userID).Error; err == nil && responder.Nama != "" {
			responderName = responder.Nama
		}

		now := time.Now().Format("2006-01-02 15:04")
		inv.UpdatedAt = now

		tx.Model(&model.InboxMessage{}).Where("invite_id = ? AND LOWER(user_email) = LOWER(?)", inviteID, userEmail).Update("is_read", true)

		if accept {
			inv.Status = "accepted"
			if err := tx.Save(&inv).Error; err != nil {
				return err
			}

			var count int64
			tx.Model(&model.Collaborator{}).Where("todo_id = ? AND LOWER(email) = LOWER(?)", inv.TodoID, inv.InviteeEmail).Count(&count)
			if count == 0 {
				collab := model.Collaborator{
					TodoID: inv.TodoID,
					Email:  inv.InviteeEmail,
				}
				if err := tx.Create(&collab).Error; err != nil {
					return err
				}
			}

			tx.Create(&model.InboxMessage{
				UserID:    inv.InviterID,
				UserEmail: inv.InviterEmail,
				Sender:    responderName,
				Title:     "Undangan Kolaborasi Diterima (ACC)",
				Message:   fmt.Sprintf("%s telah menerima (ACC) undangan kolaborasi untuk tugas '%s'.", responderName, inv.TodoTitle),
				Type:      "accepted",
				InviteID:  inv.ID,
				TodoID:    inv.TodoID,
				IsRead:    false,
				CreatedAt: now,
			})
		} else {
			inv.Status = "rejected"
			if err := tx.Save(&inv).Error; err != nil {
				return err
			}

			tx.Create(&model.InboxMessage{
				UserID:    inv.InviterID,
				UserEmail: inv.InviterEmail,
				Sender:    responderName,
				Title:     "Undangan Kolaborasi Ditolak",
				Message:   fmt.Sprintf("%s menolak undangan kolaborasi untuk tugas '%s'.", responderName, inv.TodoTitle),
				Type:      "rejected",
				InviteID:  inv.ID,
				TodoID:    inv.TodoID,
				IsRead:    false,
				CreatedAt: now,
			})
		}

		return nil
	})
}

func (r *GORMRepository) GetInvitationsForUser(userEmail string) ([]model.Invitation, error) {
	var result []model.Invitation
	err := r.db.Where("LOWER(invitee_email) = LOWER(?) OR LOWER(inviter_email) = LOWER(?)", userEmail, userEmail).Order("id desc").Find(&result).Error
	return result, err
}

func (r *GORMRepository) GetPendingInvitationsForTodo(todoID int) ([]model.Invitation, error) {
	var result []model.Invitation
	err := r.db.Where("todo_id = ? AND status = ?", todoID, "pending").Order("id asc").Find(&result).Error
	return result, err
}

func (r *GORMRepository) GetInboxMessages(userEmail string) ([]model.InboxMessage, error) {
	var result []model.InboxMessage
	err := r.db.Where("LOWER(user_email) = LOWER(?)", userEmail).Order("id desc").Find(&result).Error
	if err != nil {
		return nil, err
	}

	for i := range result {
		if result[i].Type == "invite" && result[i].InviteID > 0 {
			var inv model.Invitation
			if r.db.First(&inv, result[i].InviteID).Error == nil && inv.Status != "pending" {
				result[i].Type = inv.Status
			}
		}
	}
	return result, nil
}

func (r *GORMRepository) GetUnreadInboxCount(userEmail string) (int, error) {
	var count int64
	err := r.db.Model(&model.InboxMessage{}).Where("LOWER(user_email) = LOWER(?) AND is_read = false", userEmail).Count(&count).Error
	return int(count), err
}

func (r *GORMRepository) MarkAllInboxAsRead(userEmail string) error {
	return r.db.Model(&model.InboxMessage{}).Where("LOWER(user_email) = LOWER(?)", userEmail).Update("is_read", true).Error
}

func (r *GORMRepository) GetTeamMembers(userEmail string) ([]model.TeamMember, error) {
	var currentUser model.User
	r.db.Where("LOWER(email) = LOWER(?)", userEmail).First(&currentUser)

	memberMap := make(map[string]*model.TeamMember)
	if currentUser.Email != "" {
		initial := "U"
		if len(currentUser.Nama) > 0 {
			initial = strings.ToUpper(string(currentUser.Nama[0]))
		}
		memberMap[strings.ToLower(currentUser.Email)] = &model.TeamMember{
			ID:       currentUser.ID,
			Nama:     currentUser.Nama,
			Email:    currentUser.Email,
			Initial:  initial,
			Role:     "Admin",
			Status:   "active",
			Projects: []string{},
		}
	}

	var allTodos []model.Todo
	r.db.Find(&allTodos)

	var allUsers []model.User
	r.db.Find(&allUsers)

	userLookup := make(map[string]model.User)
	for _, u := range allUsers {
		userLookup[strings.ToLower(u.Email)] = u
	}

	var allCollabs []model.Collaborator
	r.db.Find(&allCollabs)

	var allInvs []model.Invitation
	r.db.Where("status = ?", "pending").Find(&allInvs)

	for _, t := range allTodos {
		isOwner := currentUser.ID > 0 && t.UserID == currentUser.ID

		var isCollab bool
		for _, c := range allCollabs {
			if c.TodoID == t.ID && strings.EqualFold(c.Email, userEmail) {
				isCollab = true
				break
			}
		}

		if isOwner {
			if m, ok := memberMap[strings.ToLower(currentUser.Email)]; ok {
				m.Projects = append(m.Projects, t.Title)
			}
			for _, c := range allCollabs {
				if c.TodoID == t.ID && !strings.EqualFold(c.Email, userEmail) {
					emailLower := strings.ToLower(c.Email)
					if _, exists := memberMap[emailLower]; !exists {
						uName := c.Email
						uID := 0
						if u, ok := userLookup[emailLower]; ok {
							uName = u.Nama
							uID = u.ID
						}
						init := "C"
						if len(uName) > 0 {
							init = strings.ToUpper(string(uName[0]))
						}
						memberMap[emailLower] = &model.TeamMember{
							ID:       uID,
							Nama:     uName,
							Email:    c.Email,
							Initial:  init,
							Role:     "Member",
							Status:   "active",
							Projects: []string{t.Title},
						}
					} else {
						memberMap[emailLower].Projects = append(memberMap[emailLower].Projects, t.Title)
					}
				}
			}
			for _, inv := range allInvs {
				if inv.TodoID == t.ID {
					emailLower := strings.ToLower(inv.InviteeEmail)
					if _, exists := memberMap[emailLower]; !exists {
						uName := inv.InviteeEmail
						uID := 0
						if u, ok := userLookup[emailLower]; ok {
							uName = u.Nama
							uID = u.ID
						}
						init := "P"
						if len(uName) > 0 {
							init = strings.ToUpper(string(uName[0]))
						}
						memberMap[emailLower] = &model.TeamMember{
							ID:       uID,
							Nama:     uName,
							Email:    inv.InviteeEmail,
							Initial:  init,
							Role:     "Member (Pending)",
							Status:   "pending",
							Projects: []string{t.Title},
						}
					}
				}
			}
		} else if isCollab {
			for _, u := range allUsers {
				if u.ID == t.UserID && !strings.EqualFold(u.Email, userEmail) {
					emailLower := strings.ToLower(u.Email)
					if _, exists := memberMap[emailLower]; !exists {
						init := "P"
						if len(u.Nama) > 0 {
							init = strings.ToUpper(string(u.Nama[0]))
						}
						memberMap[emailLower] = &model.TeamMember{
							ID:       u.ID,
							Nama:     u.Nama,
							Email:    u.Email,
							Initial:  init,
							Role:     "Admin",
							Status:   "active",
							Projects: []string{t.Title},
						}
					} else {
						memberMap[emailLower].Projects = append(memberMap[emailLower].Projects, t.Title)
					}
					break
				}
			}
		}
	}

	for _, u := range allUsers {
		emailLower := strings.ToLower(u.Email)
		if _, exists := memberMap[emailLower]; !exists {
			init := "U"
			if len(u.Nama) > 0 {
				init = strings.ToUpper(string(u.Nama[0]))
			}
			memberMap[emailLower] = &model.TeamMember{
				ID:       u.ID,
				Nama:     u.Nama,
				Email:    u.Email,
				Initial:  init,
				Role:     "Member",
				Status:   "active",
				Projects: []string{"Belum ada proyek"},
			}
		}
	}

	var members []model.TeamMember
	if self, ok := memberMap[strings.ToLower(userEmail)]; ok {
		members = append(members, *self)
		delete(memberMap, strings.ToLower(userEmail))
	}
	for _, m := range memberMap {
		members = append(members, *m)
	}

	return members, nil
}

func (r *GORMRepository) GetUserSimpleTodos(userID int, userEmail string) ([]model.TodoSimple, error) {
	var todos []model.Todo
	err := r.db.Where("user_id = ? OR id IN (SELECT todo_id FROM collaborators WHERE LOWER(email) = LOWER(?))", userID, strings.TrimSpace(userEmail)).Order("id desc").Find(&todos).Error
	if err != nil {
		return nil, err
	}

	var result []model.TodoSimple
	for _, t := range todos {
		result = append(result, model.TodoSimple{
			ID:    t.ID,
			Title: t.Title,
		})
	}
	return result, nil
}

func (r *GORMRepository) UpdateUserPreferences(userID int, avatar, language string) error {
	updates := map[string]interface{}{}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if language != "" {
		updates["language"] = language
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *GORMRepository) CalculateStats(todos []model.Todo) (countDone, countOngoing, countPending, total, donePercent, ongoingPercent, pendingPercent, ongoingEndPercent int) {
	return CalculateStats(todos)
}
