package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"

	"todo-app-test/internal/model"
)

// JSONRepository implementasi Repository memakai file JSON lokal
type JSONRepository struct {
	filePath string
	mu       sync.Mutex
}

// NewJSONRepository membuat instance repository berbasis JSON
func NewJSONRepository(filePath string) *JSONRepository {
	repo := &JSONRepository{filePath: filePath}
	repo.loadData()
	return repo
}

func (r *JSONRepository) loadData() model.DataStore {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			data := model.DataStore{
				Users:         []model.User{},
				Todos:         []model.Todo{},
				SubTodos:      []model.SubTodo{},
				Collaborators: []model.Collaborator{},
				Invitations:   []model.Invitation{},
				InboxMessages: []model.InboxMessage{},
			}
			r.saveDataLocked(data)
			return data
		}
		log.Println("Gagal membaca data:", err)
		return model.DataStore{}
	}

	var data model.DataStore
	if err := json.Unmarshal(file, &data); err != nil {
		log.Println("Format data tidak valid:", err)
		return model.DataStore{}
	}

	if data.Users == nil {
		data.Users = []model.User{}
	}
	if data.Todos == nil {
		data.Todos = []model.Todo{}
	}
	if data.SubTodos == nil {
		data.SubTodos = []model.SubTodo{}
	}
	if data.Collaborators == nil {
		data.Collaborators = []model.Collaborator{}
	}
	if data.Invitations == nil {
		data.Invitations = []model.Invitation{}
	}
	if data.InboxMessages == nil {
		data.InboxMessages = []model.InboxMessage{}
	}

	return data
}

func (r *JSONRepository) saveDataLocked(data model.DataStore) {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("Gagal mengubah data menjadi JSON:", err)
		return
	}

	if err := os.WriteFile(r.filePath, file, 0644); err != nil {
		log.Println("Gagal menyimpan data:", err)
	}
}

func (r *JSONRepository) saveData(data model.DataStore) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveDataLocked(data)
}

func (r *JSONRepository) GetUserByID(id int) (model.User, bool) {
	data := r.loadData()
	for _, user := range data.Users {
		if user.ID == id {
			return user, true
		}
	}
	return model.User{}, false
}

func (r *JSONRepository) GetUserByUsernameOrEmail(identifier string) (model.User, bool) {
	data := r.loadData()
	for _, user := range data.Users {
		if (user.Username == identifier || strings.EqualFold(user.Email, identifier)) {
			return user, true
		}
	}
	return model.User{}, false
}

func (r *JSONRepository) UserExists(username, email string) bool {
	data := r.loadData()
	for _, user := range data.Users {
		if strings.EqualFold(user.Username, username) || strings.EqualFold(user.Email, email) {
			return true
		}
	}
	return false
}

func (r *JSONRepository) CreateUser(user model.User) (model.User, error) {
	data := r.loadData()

	newID := 1
	for _, u := range data.Users {
		if u.ID >= newID {
			newID = u.ID + 1
		}
	}
	user.ID = newID

	data.Users = append(data.Users, user)
	r.saveData(data)
	return user, nil
}

func (r *JSONRepository) isTodoCollaborator(data model.DataStore, todo model.Todo, email string) bool {
	email = strings.TrimSpace(email)
	for _, c := range data.Collaborators {
		if c.TodoID == todo.ID && strings.EqualFold(strings.TrimSpace(c.Email), email) {
			return true
		}
	}
	for _, c := range todo.Collaborators {
		if strings.EqualFold(strings.TrimSpace(c), email) {
			return true
		}
	}
	return false
}

func (r *JSONRepository) GetTodosForUser(userID int, userEmail string, statusFilter string) ([]model.Todo, error) {
	data := r.loadData()
	var userTodos []model.Todo

	for _, todo := range data.Todos {
		isOwner := todo.UserID == userID
		isCollab := r.isTodoCollaborator(data, todo, userEmail)

		if !isOwner && !isCollab {
			continue
		}

		if statusFilter != "" && todo.Status != statusFilter {
			continue
		}

		var subs []model.SubTodo
		for _, sub := range data.SubTodos {
			if sub.TodoID == todo.ID {
				subs = append(subs, sub)
			}
		}
		todo.SubTodos = subs
		todo.IsOwner = isOwner

		totalSub := len(subs)
		completed := 0
		for _, sub := range subs {
			if sub.IsCompleted {
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

		var collaborators []string
		for _, c := range data.Collaborators {
			if c.TodoID == todo.ID {
				collaborators = append(collaborators, c.Email)
			}
		}
		if len(collaborators) == 0 {
			collaborators = todo.Collaborators
		}
		todo.Collaborators = collaborators

		userTodos = append(userTodos, todo)
	}

	return userTodos, nil
}

func (r *JSONRepository) GetTodoByID(id int, currentUserID int, userEmail string) (model.Todo, bool, error) {
	data := r.loadData()
	var foundTodo model.Todo
	found := false

	for _, t := range data.Todos {
		if t.ID == id {
			foundTodo = t
			found = true
			break
		}
	}

	if !found {
		return model.Todo{}, false, nil
	}

	isOwner := foundTodo.UserID == currentUserID
	isCollab := r.isTodoCollaborator(data, foundTodo, userEmail)
	if !isOwner && !isCollab {
		return model.Todo{}, false, errors.New("akses ditolak")
	}

	foundTodo.IsOwner = isOwner
	if isOwner {
		foundTodo.UserRole = "Admin"
	} else {
		foundTodo.UserRole = "Member"
	}
	for _, u := range data.Users {
		if u.ID == foundTodo.UserID {
			foundTodo.OwnerName = u.Nama
			foundTodo.OwnerEmail = u.Email
			break
		}
	}
	foundTodo.SubTodos = []model.SubTodo{}

	for _, sub := range data.SubTodos {
		if sub.TodoID == id {
			foundTodo.SubTodos = append(foundTodo.SubTodos, sub)
		}
	}

	total, completed := len(foundTodo.SubTodos), 0
	for _, sub := range foundTodo.SubTodos {
		if sub.IsCompleted {
			completed++
		}
	}

	if total > 0 {
		foundTodo.Progress = completed * 100 / total
	} else if foundTodo.Status == "sudah" {
		foundTodo.Progress = 100
	} else {
		foundTodo.Progress = 0
	}

	foundTodo.Collaborators = []string{}
	for _, c := range data.Collaborators {
		if c.TodoID == id {
			foundTodo.Collaborators = append(foundTodo.Collaborators, c.Email)
		}
	}

	foundTodo.PendingInvitations = []model.Invitation{}
	for _, inv := range data.Invitations {
		if inv.TodoID == id && inv.Status == "pending" {
			foundTodo.PendingInvitations = append(foundTodo.PendingInvitations, inv)
		}
	}

	return foundTodo, true, nil
}

func (r *JSONRepository) CreateTodo(todo model.Todo, subTitles []string, collaboratorEmail string) error {
	data := r.loadData()

	newID := 1
	for _, t := range data.Todos {
		if t.ID >= newID {
			newID = t.ID + 1
		}
	}
	todo.ID = newID
	todo.CreatedAt = time.Now().Format("2006-01-02 15:04")
	todo.Status = "belum"
	if strings.TrimSpace(todo.Priority) == "" {
		todo.Priority = "medium"
	}

	data.Todos = append(data.Todos, todo)

	subID := 1
	for _, sub := range data.SubTodos {
		if sub.ID >= subID {
			subID = sub.ID + 1
		}
	}

	for _, subTitle := range subTitles {
		subTitle = strings.TrimSpace(subTitle)
		if subTitle == "" {
			continue
		}
		data.SubTodos = append(data.SubTodos, model.SubTodo{
			ID:     subID,
			TodoID: newID,
			Title:  subTitle,
		})
		subID++
	}

	collaboratorEmail = strings.TrimSpace(collaboratorEmail)
	if collaboratorEmail != "" {
		if err := r.addCollaboratorInternal(&data, newID, todo.UserID, collaboratorEmail); err != nil {
			return err
		}
	}

	r.saveData(data)
	return nil
}

func (r *JSONRepository) UpdateTodo(id int, userID int, title, description, deadline, status string, priority ...string) error {
	data := r.loadData()
	index := -1

	for i, todo := range data.Todos {
		if todo.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("tugas tidak ditemukan")
	}

	if data.Todos[index].UserID != userID {
		return errors.New("hanya admin yang dapat merubah project")
	}

	data.Todos[index].Title = title
	data.Todos[index].Description = description
	data.Todos[index].Deadline = deadline

	if status == "belum" || status == "sedang" || status == "sudah" {
		data.Todos[index].Status = status
	}

	if len(priority) > 0 && strings.TrimSpace(priority[0]) != "" {
		data.Todos[index].Priority = strings.TrimSpace(priority[0])
	}

	if data.Todos[index].Status == "sudah" {
		data.Todos[index].CompletedAt = time.Now().Format("2006-01-02 15:04")
	} else {
		data.Todos[index].CompletedAt = ""
	}

	r.saveData(data)
	return nil
}

func (r *JSONRepository) DeleteTodo(id int, userID int) error {
	data := r.loadData()
	found := false

	for _, todo := range data.Todos {
		if todo.ID == id {
			found = true
			if todo.UserID != userID {
				return errors.New("hanya admin yang dapat merubah project")
			}
			break
		}
	}

	if !found {
		return errors.New("tugas tidak ditemukan")
	}

	todos := []model.Todo{}
	for _, todo := range data.Todos {
		if todo.ID != id {
			todos = append(todos, todo)
		}
	}
	data.Todos = todos

	subs := []model.SubTodo{}
	for _, sub := range data.SubTodos {
		if sub.TodoID != id {
			subs = append(subs, sub)
		}
	}
	data.SubTodos = subs

	collabs := []model.Collaborator{}
	for _, c := range data.Collaborators {
		if c.TodoID != id {
			collabs = append(collabs, c)
		}
	}
	data.Collaborators = collabs

	r.saveData(data)
	return nil
}

func (r *JSONRepository) ToggleSubTodo(todoID int, subID int, userID int) error {
	data := r.loadData()

	foundTodo := false
	for _, t := range data.Todos {
		if t.ID == todoID {
			if t.UserID != userID {
				return errors.New("hanya pemilik yang dapat mengubah tugas")
			}
			foundTodo = true
			break
		}
	}
	if !foundTodo {
		return errors.New("tugas tidak ditemukan")
	}

	foundSub := false
	for i := range data.SubTodos {
		if data.SubTodos[i].ID == subID && data.SubTodos[i].TodoID == todoID {
			data.SubTodos[i].IsCompleted = !data.SubTodos[i].IsCompleted
			foundSub = true
			break
		}
	}
	if !foundSub {
		return errors.New("sub-task tidak ditemukan")
	}

	total, completed := 0, 0
	for _, sub := range data.SubTodos {
		if sub.TodoID == todoID {
			total++
			if sub.IsCompleted {
				completed++
			}
		}
	}

	for i := range data.Todos {
		if data.Todos[i].ID == todoID {
			if total > 0 {
				data.Todos[i].Progress = completed * 100 / total
			}

			if total > 0 && completed == total {
				data.Todos[i].Status = "sudah"
				data.Todos[i].CompletedAt = time.Now().Format("2006-01-02 15:04")
			} else if completed > 0 {
				data.Todos[i].Status = "sedang"
				data.Todos[i].CompletedAt = ""
			} else {
				data.Todos[i].Status = "belum"
				data.Todos[i].CompletedAt = ""
			}
			break
		}
	}

	r.saveData(data)
	return nil
}

func (r *JSONRepository) UpdateTodoStatus(todoID int, status string, userID int) error {
	if status != "belum" && status != "sedang" && status != "sudah" {
		return errors.New("status tidak valid")
	}

	data := r.loadData()
	found := false
	for i := range data.Todos {
		if data.Todos[i].ID == todoID {
			if data.Todos[i].UserID != userID {
				return errors.New("hanya pemilik yang dapat mengubah status tugas")
			}
			data.Todos[i].Status = status
			if status == "sudah" {
				data.Todos[i].Progress = 100
				data.Todos[i].CompletedAt = time.Now().Format("2006-01-02 15:04")
			} else {
				data.Todos[i].CompletedAt = ""
			}
			found = true
			break
		}
	}

	if !found {
		return errors.New("tugas tidak ditemukan")
	}

	r.saveData(data)
	return nil
}

func (r *JSONRepository) AttachDocument(todoID int, userID int, document string) error {
	data := r.loadData()
	found := false
	for i := range data.Todos {
		if data.Todos[i].ID == todoID {
			user, ok := r.getUserByIDFromData(data, userID)
			isCollab := ok && r.isTodoCollaborator(data, data.Todos[i], user.Email)
			if data.Todos[i].UserID != userID && !isCollab {
				return errors.New("hanya pemilik atau kolaborator tugas yang dapat melampirkan dokumen")
			}
			data.Todos[i].Document = strings.TrimSpace(document)
			found = true
			break
		}
	}
	if !found {
		return errors.New("tugas tidak ditemukan")
	}
	r.saveData(data)
	return nil
}

func (r *JSONRepository) AddCollaborator(todoID int, ownerID int, email string) error {
	_, err := r.InviteCollaborator(todoID, ownerID, email)
	return err
}

func (r *JSONRepository) addCollaboratorInternal(data *model.DataStore, todoID, ownerID int, email string) error {
	_, err := r.inviteCollaboratorInternal(data, todoID, ownerID, email)
	return err
}

func (r *JSONRepository) InviteCollaborator(todoID int, inviterID int, inviteeEmailOrUsername string) (model.Invitation, error) {
	data := r.loadData()
	inv, err := r.inviteCollaboratorInternal(&data, todoID, inviterID, inviteeEmailOrUsername)
	if err != nil {
		return model.Invitation{}, err
	}
	r.saveData(data)
	return inv, nil
}

func (r *JSONRepository) getUserByIDFromData(data model.DataStore, id int) (model.User, bool) {
	for _, u := range data.Users {
		if u.ID == id {
			return u, true
		}
	}
	return model.User{}, false
}

func (r *JSONRepository) inviteCollaboratorInternal(data *model.DataStore, todoID, inviterID int, inviteeInput string) (model.Invitation, error) {
	inviteeInput = strings.TrimSpace(inviteeInput)
	if inviteeInput == "" {
		return model.Invitation{}, errors.New("email atau username kolaborator wajib diisi")
	}

	todoIndex := -1
	for i := range data.Todos {
		if data.Todos[i].ID == todoID {
			todoIndex = i
			break
		}
	}
	if todoIndex == -1 {
		return model.Invitation{}, errors.New("tugas tidak ditemukan")
	}

	todo := data.Todos[todoIndex]

	inviter, ok := r.getUserByIDFromData(*data, inviterID)
	if !ok {
		return model.Invitation{}, errors.New("pengguna pengundang tidak valid")
	}

	isOwner := todo.UserID == inviterID
	isCollab := r.isTodoCollaborator(*data, todo, inviter.Email)
	if !isOwner && !isCollab {
		return model.Invitation{}, errors.New("hanya pemilik atau kolaborator tugas yang dapat mengundang anggota baru")
	}

	var invitee *model.User
	for i := range data.Users {
		u := &data.Users[i]
		if strings.EqualFold(strings.TrimSpace(u.Email), inviteeInput) || strings.EqualFold(strings.TrimSpace(u.Username), inviteeInput) {
			invitee = u
			break
		}
	}

	if invitee == nil {
		return model.Invitation{}, errors.New("pengguna belum terdaftar. Pastikan email/username sudah benar dan terdaftar")
	}

	if invitee.ID == inviterID || strings.EqualFold(invitee.Email, inviter.Email) {
		return model.Invitation{}, errors.New("anda tidak dapat mengundang diri sendiri")
	}

	if invitee.ID == todo.UserID {
		return model.Invitation{}, errors.New("pengguna tersebut adalah pemilik tugas ini")
	}

	if r.isTodoCollaborator(*data, todo, invitee.Email) {
		return model.Invitation{}, errors.New("pengguna tersebut sudah menjadi kolaborator aktif pada tugas ini")
	}

	for _, inv := range data.Invitations {
		if inv.TodoID == todoID && strings.EqualFold(inv.InviteeEmail, invitee.Email) && inv.Status == "pending" {
			return model.Invitation{}, errors.New("undangan kolaborasi untuk pengguna ini sudah dikirim dan sedang menunggu persetujuan (ACC)")
		}
	}

	newInvID := 1
	for _, inv := range data.Invitations {
		if inv.ID >= newInvID {
			newInvID = inv.ID + 1
		}
	}

	now := time.Now().Format("2006-01-02 15:04")
	invitation := model.Invitation{
		ID:           newInvID,
		TodoID:       todoID,
		TodoTitle:    todo.Title,
		InviterID:    inviter.ID,
		InviterName:  inviter.Nama,
		InviterEmail: inviter.Email,
		InviteeEmail: invitee.Email,
		Status:       "pending",
		CreatedAt:    now,
	}
	data.Invitations = append(data.Invitations, invitation)

	newMsgID := 1
	for _, msg := range data.InboxMessages {
		if msg.ID >= newMsgID {
			newMsgID = msg.ID + 1
		}
	}

	inboxMsg := model.InboxMessage{
		ID:        newMsgID,
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
	data.InboxMessages = append(data.InboxMessages, inboxMsg)

	return invitation, nil
}

func (r *JSONRepository) RespondInvitation(inviteID int, userID int, userEmail string, accept bool) error {
	data := r.loadData()

	invIndex := -1
	for i := range data.Invitations {
		if data.Invitations[i].ID == inviteID {
			invIndex = i
			break
		}
	}
	if invIndex == -1 {
		return errors.New("undangan tidak ditemukan")
	}

	inv := &data.Invitations[invIndex]
	if !strings.EqualFold(inv.InviteeEmail, userEmail) {
		return errors.New("anda tidak memiliki hak akses untuk merespons undangan ini")
	}

	if inv.Status != "pending" {
		return fmt.Errorf("undangan sudah di-%s sebelumnya", inv.Status)
	}

	responder, _ := r.getUserByIDFromData(data, userID)
	responderName := responder.Nama
	if responderName == "" {
		responderName = userEmail
	}

	now := time.Now().Format("2006-01-02 15:04")
	inv.UpdatedAt = now

	// Tandai pesan inbox terkait sebagai sudah dibaca
	for i := range data.InboxMessages {
		if data.InboxMessages[i].InviteID == inviteID && strings.EqualFold(data.InboxMessages[i].UserEmail, userEmail) {
			data.InboxMessages[i].IsRead = true
		}
	}

	newMsgID := 1
	for _, msg := range data.InboxMessages {
		if msg.ID >= newMsgID {
			newMsgID = msg.ID + 1
		}
	}

	if accept {
		inv.Status = "accepted"

		exists := false
		for _, c := range data.Collaborators {
			if c.TodoID == inv.TodoID && strings.EqualFold(c.Email, inv.InviteeEmail) {
				exists = true
				break
			}
		}
		if !exists {
			data.Collaborators = append(data.Collaborators, model.Collaborator{
				TodoID: inv.TodoID,
				Email:  inv.InviteeEmail,
			})
		}

		data.InboxMessages = append(data.InboxMessages, model.InboxMessage{
			ID:        newMsgID,
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

		data.InboxMessages = append(data.InboxMessages, model.InboxMessage{
			ID:        newMsgID,
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

	r.saveData(data)
	return nil
}

func (r *JSONRepository) GetInvitationsForUser(userEmail string) ([]model.Invitation, error) {
	data := r.loadData()
	var result []model.Invitation
	for _, inv := range data.Invitations {
		if strings.EqualFold(inv.InviteeEmail, userEmail) || strings.EqualFold(inv.InviterEmail, userEmail) {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *JSONRepository) GetPendingInvitationsForTodo(todoID int) ([]model.Invitation, error) {
	data := r.loadData()
	var result []model.Invitation
	for _, inv := range data.Invitations {
		if inv.TodoID == todoID && inv.Status == "pending" {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *JSONRepository) GetInboxMessages(userEmail string) ([]model.InboxMessage, error) {
	data := r.loadData()
	var result []model.InboxMessage
	for i := len(data.InboxMessages) - 1; i >= 0; i-- {
		msg := data.InboxMessages[i]
		if strings.EqualFold(msg.UserEmail, userEmail) {
			if msg.Type == "invite" && msg.InviteID > 0 {
				for _, inv := range data.Invitations {
					if inv.ID == msg.InviteID {
						if inv.Status != "pending" {
							msg.Type = inv.Status
						}
						break
					}
				}
			}
			result = append(result, msg)
		}
	}
	return result, nil
}

func (r *JSONRepository) GetUnreadInboxCount(userEmail string) (int, error) {
	data := r.loadData()
	count := 0
	for _, msg := range data.InboxMessages {
		if strings.EqualFold(msg.UserEmail, userEmail) && !msg.IsRead {
			count++
		}
	}
	return count, nil
}

func (r *JSONRepository) MarkAllInboxAsRead(userEmail string) error {
	data := r.loadData()
	for i := range data.InboxMessages {
		if strings.EqualFold(data.InboxMessages[i].UserEmail, userEmail) {
			data.InboxMessages[i].IsRead = true
		}
	}
	r.saveData(data)
	return nil
}

func (r *JSONRepository) GetTeamMembers(userEmail string) ([]model.TeamMember, error) {
	data := r.loadData()

	var currentUser model.User
	for _, u := range data.Users {
		if strings.EqualFold(u.Email, userEmail) {
			currentUser = u
			break
		}
	}

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

	for _, t := range data.Todos {
		isOwner := currentUser.ID > 0 && t.UserID == currentUser.ID
		isCollab := r.isTodoCollaborator(data, t, userEmail)

		if isOwner {
			if m, ok := memberMap[strings.ToLower(currentUser.Email)]; ok {
				m.Projects = append(m.Projects, t.Title)
			}
			for _, c := range data.Collaborators {
				if c.TodoID == t.ID && !strings.EqualFold(c.Email, userEmail) {
					emailLower := strings.ToLower(c.Email)
					if _, exists := memberMap[emailLower]; !exists {
						var uName string
						var uID int
						for _, u := range data.Users {
							if strings.EqualFold(u.Email, c.Email) {
								uName = u.Nama
								uID = u.ID
								break
							}
						}
						if uName == "" {
							uName = c.Email
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
			for _, inv := range data.Invitations {
				if inv.TodoID == t.ID && inv.Status == "pending" {
					emailLower := strings.ToLower(inv.InviteeEmail)
					if _, exists := memberMap[emailLower]; !exists {
						var uName string
						var uID int
						for _, u := range data.Users {
							if strings.EqualFold(u.Email, inv.InviteeEmail) {
								uName = u.Nama
								uID = u.ID
								break
							}
						}
						if uName == "" {
							uName = inv.InviteeEmail
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
			var ownerUser model.User
			for _, u := range data.Users {
				if u.ID == t.UserID {
					ownerUser = u
					break
				}
			}
			if ownerUser.Email != "" && !strings.EqualFold(ownerUser.Email, userEmail) {
				emailLower := strings.ToLower(ownerUser.Email)
				if _, exists := memberMap[emailLower]; !exists {
					init := "P"
					if len(ownerUser.Nama) > 0 {
						init = strings.ToUpper(string(ownerUser.Nama[0]))
					}
					memberMap[emailLower] = &model.TeamMember{
						ID:       ownerUser.ID,
						Nama:     ownerUser.Nama,
						Email:    ownerUser.Email,
						Initial:  init,
						Role:     "Admin",
						Status:   "active",
						Projects: []string{t.Title},
					}
				} else {
					memberMap[emailLower].Projects = append(memberMap[emailLower].Projects, t.Title)
				}
			}
		}
	}

	for _, u := range data.Users {
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

func (r *JSONRepository) GetUserSimpleTodos(userID int, userEmail string) ([]model.TodoSimple, error) {
	data := r.loadData()
	var result []model.TodoSimple
	for _, t := range data.Todos {
		if t.UserID == userID || r.isTodoCollaborator(data, t, userEmail) {
			result = append(result, model.TodoSimple{
				ID:    t.ID,
				Title: t.Title,
			})
		}
	}
	return result, nil
}

func (r *JSONRepository) UpdateUserPreferences(userID int, avatar, language string) error {
	data := r.loadData()
	found := false
	for i, u := range data.Users {
		if u.ID == userID {
			if avatar != "" {
				data.Users[i].Avatar = avatar
			}
			if language != "" {
				data.Users[i].Language = language
			}
			found = true
			break
		}
	}
	if !found {
		return errors.New("user not found")
	}
	r.saveData(data)
	return nil
}

func (r *JSONRepository) CalculateStats(todos []model.Todo) (countDone, countOngoing, countPending, total, donePercent, ongoingPercent, pendingPercent, ongoingEndPercent int) {
	return CalculateStats(todos)
}

func emailIsValid(email string) bool {
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
