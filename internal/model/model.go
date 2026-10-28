package model

import "time"

// User merepresentasikan data pengguna
type User struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Nama      string    `json:"nama" gorm:"size:255;not null"`
	Username  string    `json:"username" gorm:"size:100;uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	Password  string    `json:"password" gorm:"size:255;not null"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"autoCreateTime"`
	Avatar    string    `json:"avatar,omitempty" gorm:"type:text"`
	Language  string    `json:"language,omitempty" gorm:"size:10;default:'en'"`
}

// SubTodo merepresentasikan item checklist dari suatu Todo
type SubTodo struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TodoID      int    `json:"todo_id" gorm:"index;not null"`
	Title       string `json:"title" gorm:"size:255;not null"`
	IsCompleted bool   `json:"is_completed" gorm:"default:false"`
}

// Collaborator merepresentasikan kolaborator yang ditambahkan ke Todo
type Collaborator struct {
	ID     int    `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	TodoID int    `json:"todo_id" gorm:"index;not null"`
	Email  string `json:"email" gorm:"size:255;not null"`
	Role   string `json:"role,omitempty" gorm:"size:50;default:'Member'"`
}

// Invitation merepresentasikan undangan kolaborasi antar pengguna
type Invitation struct {
	ID           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TodoID       int    `json:"todo_id" gorm:"index;not null"`
	TodoTitle    string `json:"todo_title" gorm:"size:255;not null"`
	InviterID    int    `json:"inviter_id" gorm:"index;not null"`
	InviterName  string `json:"inviter_name" gorm:"size:255;not null"`
	InviterEmail string `json:"inviter_email" gorm:"size:255;not null"`
	InviteeEmail string `json:"invitee_email" gorm:"size:255;index;not null"`
	Status       string `json:"status" gorm:"size:30;default:'pending'"` // "pending", "accepted", "rejected"
	CreatedAt    string `json:"created_at" gorm:"size:50"`
	UpdatedAt    string `json:"updated_at,omitempty" gorm:"size:50"`
}

// InboxMessage merepresentasikan notifikasi atau pesan di inbox pengguna
type InboxMessage struct {
	ID        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int    `json:"user_id,omitempty" gorm:"index"`
	UserEmail string `json:"user_email" gorm:"size:255;index;not null"`
	Sender    string `json:"sender" gorm:"size:255;not null"`
	Title     string `json:"title" gorm:"size:255;not null"`
	Message   string `json:"message" gorm:"type:text;not null"`
	Type      string `json:"type" gorm:"size:50;default:'invite'"` // "invite", "info", "accepted", "rejected"
	InviteID  int    `json:"invite_id,omitempty" gorm:"index"`
	TodoID    int    `json:"todo_id,omitempty" gorm:"index"`
	IsRead    bool   `json:"is_read" gorm:"default:false"`
	CreatedAt string `json:"created_at" gorm:"size:50"`
}

// TeamMember merepresentasikan anggota tim untuk tampilan halaman Team
type TeamMember struct {
	ID       int      `json:"id"`
	Nama     string   `json:"nama"`
	Email    string   `json:"email"`
	Initial  string   `json:"initial"`
	Role     string   `json:"role"`     // "Pemilik", "Kolaborator", "Menunggu ACC"
	Projects []string `json:"projects"` // Daftar judul tugas yang dikerjakan bersama
	Status   string   `json:"status"`   // "active", "pending"
}

// TodoSimple adalah representasi ringkas Todo untuk dropdown pemilihan tugas saat invite
type TodoSimple struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// DocumentItem merepresentasikan berkas/dokumen yang terlampir pada Todo
type DocumentItem struct {
	TodoID    int    `json:"todo_id"`
	TodoTitle string `json:"todo_title"`
	FileName  string `json:"file_name"`
	Ext       string `json:"ext"`
	TypeClass string `json:"type_class"`
	Date      string `json:"date"`
}

// CalendarEvent merepresentasikan jadwal deadline tugas untuk tampilan Calendar
type CalendarEvent struct {
	TodoID      int    `json:"todo_id"`
	Title       string `json:"title"`
	Date        string `json:"date"` // Format: YYYY-MM-DD
	Time        string `json:"time,omitempty"`
	Status      string `json:"status"`
	StatusClass string `json:"status_class"`
	StatusLabel string `json:"status_label"`
	Progress    int    `json:"progress"`
	Priority    string `json:"priority,omitempty"`
	IsDeadline  bool   `json:"is_deadline"`
}

// Todo merepresentasikan data utama tugas
type Todo struct {
	ID                 int            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID             int            `json:"user_id" gorm:"index;not null"`
	Title              string         `json:"title" gorm:"size:255;not null"`
	Description        string         `json:"description" gorm:"type:text"`
	CreatedAt          string         `json:"created_at" gorm:"size:50"`
	Deadline           string         `json:"deadline" gorm:"size:50"`
	CompletedAt        string         `json:"completed_at" gorm:"size:50"`
	Status             string         `json:"status" gorm:"size:20;default:'belum'"`
	Progress           int            `json:"progress" gorm:"default:0"`
	Priority           string         `json:"priority" gorm:"size:20;default:'medium'"`
	Document           string         `json:"document,omitempty" gorm:"size:255"`
	DocumentExt        string         `json:"document_ext,omitempty" gorm:"-"`
	DocumentTypeClass  string         `json:"document_type_class,omitempty" gorm:"-"`
	SubTodos           []SubTodo      `json:"sub_todos" gorm:"foreignKey:TodoID;constraint:OnDelete:CASCADE"`
	Collaborators      []string       `json:"collaborators" gorm:"-"`
	CollabRecords      []Collaborator `json:"-" gorm:"foreignKey:TodoID;constraint:OnDelete:CASCADE"`
	PendingInvitations []Invitation   `json:"pending_invitations,omitempty" gorm:"-"`
	FlashMessage       string         `json:"flash_message,omitempty" gorm:"-"`
	FlashError         string         `json:"flash_error,omitempty" gorm:"-"`
	IsOwner            bool           `json:"-" gorm:"-"`
	OwnerName          string         `json:"owner_name,omitempty" gorm:"-"`
	OwnerEmail         string         `json:"owner_email,omitempty" gorm:"-"`
	UserRole           string         `json:"user_role,omitempty" gorm:"-"` // "Admin" atau "Member"
	IsOverdue          bool           `json:"is_overdue,omitempty" gorm:"-"`
}

// DataStore struktur JSON untuk file data.json
type DataStore struct {
	Users         []User         `json:"users"`
	Todos         []Todo         `json:"todos"`
	SubTodos      []SubTodo      `json:"sub_todos"`
	Collaborators []Collaborator `json:"collaborators"`
	Invitations   []Invitation   `json:"invitations"`
	InboxMessages []InboxMessage `json:"inbox_messages"`
}

// TodoItem adalah data tugas yang sudah dilengkapi label tampilan
// untuk kebutuhan dashboard (format tanggal, prioritas, dan inisial tim).
type TodoItem struct {
	Todo
	DeadlineLabel        string
	CreatedLabel         string
	Priority             string
	PriorityClass        string
	CollaboratorInitials []string
	IsOverdue            bool
}

// DashboardData data yang dikirim ke template dashboard.html
type DashboardData struct {
	Username          string
	Nama              string
	Email             string
	Initial           string
	Filter            string
	CountDone         int
	CountOngoing      int
	CountPending      int
	TotalVisible      int
	DonePercent       int
	OngoingPercent    int
	PendingPercent    int
	OngoingEndPercent int
	TeamCount         int
	UpcomingDeadlines int
	OverdueCount      int
	PendingTodos      []TodoItem
	OngoingTodos      []TodoItem
	DoneTodos         []TodoItem
	// Fitur Kolaborasi & Tampilan Workspace (Inbox, Team, Document, Calendar)
	ActivePage         string
	UnreadInboxCount   int
	InboxMessages      []InboxMessage
	Invitations        []Invitation
	TeamMembers        []TeamMember
	UserTodosSimple    []TodoSimple
	Documents          []DocumentItem
	CalendarEvents     []CalendarEvent
	CalendarEventsJSON string
	Avatar             string
	Language           string
	FlashMessage       string
	FlashError         string
	Todo               Todo
}

// SettingsData data yang dikirim ke template settings.html
type SettingsData struct {
	Username     string
	Nama         string
	Email        string
	Initial      string
	Avatar       string
	Language     string
	JoinedLabel  string
	TotalVisible int
	CountPending int
	CountOngoing int
	CountDone    int
}
