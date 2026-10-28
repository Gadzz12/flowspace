package repository

import "todo-app-test/internal/model"

// Repository adalah interface untuk akses data aplikasi Todo
type Repository interface {
	// User operations
	GetUserByID(id int) (model.User, bool)
	GetUserByUsernameOrEmail(identifier string) (model.User, bool)
	UserExists(username, email string) bool
	CreateUser(user model.User) (model.User, error)
	UpdateUserPreferences(userID int, avatar, language string) error

	// Todo operations
	GetTodosForUser(userID int, userEmail string, statusFilter string) ([]model.Todo, error)
	GetTodoByID(id int, currentUserID int, userEmail string) (model.Todo, bool, error)
	CreateTodo(todo model.Todo, subTitles []string, collaboratorEmail string) error
	UpdateTodo(id int, userID int, title, description, deadline, status string, priority ...string) error
	DeleteTodo(id int, userID int) error
	ToggleSubTodo(todoID int, subID int, userID int) error
	UpdateTodoStatus(todoID int, status string, userID int) error
	AddCollaborator(todoID int, ownerID int, email string) error
	AttachDocument(todoID int, userID int, document string) error

	// Collaboration & Inbox operations
	InviteCollaborator(todoID int, inviterID int, inviteeEmailOrUsername string) (model.Invitation, error)
	RespondInvitation(inviteID int, userID int, userEmail string, accept bool) error
	GetInvitationsForUser(userEmail string) ([]model.Invitation, error)
	GetPendingInvitationsForTodo(todoID int) ([]model.Invitation, error)
	GetInboxMessages(userEmail string) ([]model.InboxMessage, error)
	GetUnreadInboxCount(userEmail string) (int, error)
	MarkAllInboxAsRead(userEmail string) error
	GetTeamMembers(userEmail string) ([]model.TeamMember, error)
	GetUserSimpleTodos(userID int, userEmail string) ([]model.TodoSimple, error)

	// Stats calculation
	CalculateStats(todos []model.Todo) (countDone, countOngoing, countPending, total, donePercent, ongoingPercent, pendingPercent, ongoingEndPercent int)
}

// CalculateStats menghitung statistik progres tugas untuk dashboard
func CalculateStats(todos []model.Todo) (countDone, countOngoing, countPending, total, donePercent, ongoingPercent, pendingPercent, ongoingEndPercent int) {
	for _, todo := range todos {
		switch todo.Status {
		case "sudah":
			countDone++
		case "sedang":
			countOngoing++
		default:
			countPending++
		}
	}

	total = countDone + countOngoing + countPending
	if total > 0 {
		donePercent = countDone * 100 / total
		ongoingPercent = countOngoing * 100 / total
		if countPending == 0 {
			pendingPercent = 0
			if countOngoing > 0 && countDone > 0 {
				ongoingPercent = 100 - donePercent
			}
		} else {
			pendingPercent = 100 - donePercent - ongoingPercent
		}
		if countDone == 0 {
			donePercent = 0
		}
		if countOngoing == 0 {
			ongoingPercent = 0
			if countPending > 0 {
				pendingPercent = 100 - donePercent
			}
		}
		ongoingEndPercent = donePercent + ongoingPercent
	}
	return
}
