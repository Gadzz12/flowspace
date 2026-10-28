/* ============================================================
   Workspace UI — theme, sidebar, task filters
   ============================================================ */

(function () {
    "use strict";

    var root = document.documentElement;

    /* ---------------- Theme ---------------- */

    var THEME_KEY = "todo-theme";
    var media = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;

    function readMode() {
        try {
            return localStorage.getItem(THEME_KEY) || "system";
        } catch (err) {
            return "system";
        }
    }

    function resolveMode(mode) {
        if (mode === "dark" || mode === "light") {
            return mode;
        }
        return media && media.matches ? "dark" : "light";
    }

    function applyMode(mode) {
        var theme = resolveMode(mode);
        root.setAttribute("data-theme", theme);
        root.style.colorScheme = theme;

        document.querySelectorAll("[data-theme-toggle]").forEach(function (btn) {
            var label = theme === "dark" ? "Aktifkan tema terang" : "Aktifkan tema gelap";
            btn.setAttribute("aria-label", label);
            btn.setAttribute("title", label);
            btn.setAttribute("data-next-theme", theme === "dark" ? "light" : "dark");
        });

        document.querySelectorAll("[data-theme-option]").forEach(function (input) {
            var active = input.value === mode;
            input.checked = active;
            var card = input.closest(".theme-option");
            if (card) {
                card.classList.toggle("is-active", active);
            }
        });

        document.querySelectorAll("[data-theme-label]").forEach(function (el) {
            el.textContent = theme === "dark" ? "Gelap" : "Terang";
        });
    }

    function setMode(mode) {
        try {
            localStorage.setItem(THEME_KEY, mode);
        } catch (err) {
            /* localStorage can be unavailable; theme still applies for this page. */
        }
        applyMode(mode);
    }

    document.querySelectorAll("[data-theme-toggle]").forEach(function (btn) {
        btn.addEventListener("click", function () {
            var current = root.getAttribute("data-theme") === "dark" ? "dark" : "light";
            setMode(current === "dark" ? "light" : "dark");
        });
    });

    document.querySelectorAll("[data-theme-option]").forEach(function (input) {
        input.addEventListener("change", function () {
            if (input.checked) {
                setMode(input.value);
            }
        });
    });

    var onSystemThemeChange = function () {
        if (readMode() === "system") {
            applyMode("system");
        }
    };

    if (media) {
        if (media.addEventListener) {
            media.addEventListener("change", onSystemThemeChange);
        } else if (media.addListener) {
            media.addListener(onSystemThemeChange);
        }
    }

    applyMode(readMode());

    /* ---------------- Sidebar ---------------- */

    var SIDEBAR_KEY = "todo-sidebar";

    var collapseBtn = document.getElementById("sidebarCollapse");
    if (collapseBtn) {
        collapseBtn.addEventListener("click", function () {
            var collapsed = root.classList.toggle("sidebar-collapsed");
            collapseBtn.setAttribute("aria-expanded", collapsed ? "false" : "true");
            try {
                localStorage.setItem(SIDEBAR_KEY, collapsed ? "collapsed" : "open");
            } catch (err) {
                /* ignore */
            }
        });
    }

    var menuBtn = document.getElementById("menuToggle");
    if (menuBtn) {
        menuBtn.addEventListener("click", function () {
            root.classList.toggle("sidebar-open");
        });
    }

    var scrim = document.getElementById("sidebarScrim");
    if (scrim) {
        scrim.addEventListener("click", function () {
            root.classList.remove("sidebar-open");
        });
    }

    document.querySelectorAll(".sidebar .nav-item").forEach(function (link) {
        link.addEventListener("click", function () {
            root.classList.remove("sidebar-open");
        });
    });

    var mobileQuery = window.matchMedia ? window.matchMedia("(max-width: 1024px)") : null;
    if (mobileQuery) {
        var onViewportChange = function (event) {
            if (!event.matches) {
                root.classList.remove("sidebar-open");
            }
        };
        if (mobileQuery.addEventListener) {
            mobileQuery.addEventListener("change", onViewportChange);
        } else if (mobileQuery.addListener) {
            mobileQuery.addListener(onViewportChange);
        }
    }

    /* ---------------- Collapsible task groups ---------------- */

    document.querySelectorAll(".group-toggle").forEach(function (btn) {
        btn.addEventListener("click", function () {
            var group = btn.closest(".group");
            if (!group) {
                return;
            }
            var collapsed = group.classList.toggle("collapsed");
            btn.setAttribute("aria-expanded", collapsed ? "false" : "true");
        });
    });

    /* ---------------- Task search + priority filter ---------------- */

    var searchInputs = Array.prototype.slice.call(document.querySelectorAll("[data-task-search]"));
    var prioritySelect = document.getElementById("priorityFilter");
    var query = "";

    function applyFilters() {
        var priority = prioritySelect ? prioritySelect.value : "";
        var needle = query.trim().toLowerCase();

        document.querySelectorAll(".group").forEach(function (group) {
            var visible = 0;

            group.querySelectorAll("tr[data-search]").forEach(function (row) {
                var haystack = (row.getAttribute("data-search") || "").toLowerCase();
                var matchesText = !needle || haystack.indexOf(needle) !== -1;
                var rowPriority = (row.getAttribute("data-priority") || "").toLowerCase();
                var matchesPriority = !priority || rowPriority === priority ||
                    (priority === "high" && (rowPriority === "tinggi" || rowPriority === "high")) ||
                    (priority === "medium" && (rowPriority === "sedang" || rowPriority === "medium")) ||
                    (priority === "low" && (rowPriority === "rendah" || rowPriority === "low")) ||
                    (priority === "tinggi" && (rowPriority === "tinggi" || rowPriority === "high")) ||
                    (priority === "sedang" && (rowPriority === "sedang" || rowPriority === "medium")) ||
                    (priority === "rendah" && (rowPriority === "rendah" || rowPriority === "low"));
                var show = matchesText && matchesPriority;

                row.hidden = !show;
                if (show) {
                    visible += 1;
                }
            });

            var empty = group.querySelector(".js-group-empty");
            if (empty) {
                empty.hidden = visible > 0;
            }
        });
    }

    searchInputs.forEach(function (input) {
        input.addEventListener("input", function () {
            query = input.value;
            searchInputs.forEach(function (other) {
                if (other !== input) {
                    other.value = input.value;
                }
            });
            applyFilters();
        });

        input.addEventListener("keydown", function (event) {
            if (event.key === "Escape") {
                input.value = "";
                query = "";
                searchInputs.forEach(function (other) {
                    other.value = "";
                });
                applyFilters();
                input.blur();
            }
        });
    });

    if (prioritySelect) {
        prioritySelect.addEventListener("change", applyFilters);
    }

    if (searchInputs.length > 0) {
        applyFilters();
    }

    document.addEventListener("keydown", function (event) {
        if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
            if (searchInputs.length === 0) {
                return;
            }
            event.preventDefault();
            searchInputs[0].focus();
            searchInputs[0].select();
        }
    });

    /* ---------------- Page views: Dashboard, Inbox, Team, Documents, Calendar ---------------- */
    var pageViews = {
        dashboard: document.getElementById("page-dashboard"),
        inbox: document.getElementById("page-inbox"),
        team: document.getElementById("page-team"),
        documents: document.getElementById("page-documents"),
        calendar: document.getElementById("page-calendar")
    };

    var breadcrumbTitle = document.getElementById("breadcrumbTitle");
    var titles = {
        dashboard: "Dashboard",
        inbox: "Inbox",
        team: "Team Members",
        documents: "Documents",
        calendar: "Calendar"
    };

    /* ============================================================
       Internationalization (i18n) — English (default), ID, RU, ZH
       ============================================================ */
    var i18n = {
        en: {
            main_menu: "MAIN MENU",
            nav_dashboard: "Dashboard",
            nav_inbox: "Inbox",
            nav_team: "Team",
            nav_my_tasks: "My Tasks",
            workspace_label: "WORKSPACE",
            nav_documents: "Documents",
            nav_calendar: "Calendar",
            project_status: "PROJECT STATUS",
            status_pending: "Waiting",
            status_ongoing: "In Progress",
            status_completed: "Completed",
            nav_settings: "Settings",
            nav_logout: "Logout",
            btn_new_task: "+ New Task",
            btn_invite_member: "＋ Invite Member",
            btn_today: "Today",
            search_placeholder: "Search tasks (Cmd + K or Ctrl + K)...",
            filter_all: "All Priorities",
            priority_high: "High Priority",
            priority_medium: "Medium Priority",
            priority_low: "Low Priority",
            priority_done: "Completed",
            welcome_title: "Hello {name}, manage your projects.",
            welcome_sub: "Track, analyze, and organize tasks so team work is more focused.",
            diagram_title: "Task Progress Diagram",
            diagram_sub: "Status distribution of all tasks in workspace",
            total_tasks: "Total Tasks",
            stat_pending_title: "WAITING",
            stat_pending_sub: "Tasks waiting to start",
            stat_ongoing_title: "ON PROGRESS",
            stat_ongoing_sub: "Tasks in progress",
            stat_completed_title: "COMPLETED",
            stat_completed_sub: "Tasks completed",
            stat_deadline_title: "UPCOMING DEADLINES",
            stat_deadline_sub: "In the next 7 days",
            th_name: "Task Name",
            th_desc: "Description",
            th_deadline: "Deadline",
            th_team: "Team",
            th_priority: "Priority",
            th_actions: "Actions",
            only_you: "Only you",
            collaborators: "collaborators",
            created_on: "Created",
            overdue_badge: "⚠️ Overdue",
            cal_title: "Calendar & Schedule",
            cal_time_label: "CURRENT TIME (WIB):",
            legend_holiday: "Sunday / Red Date",
            legend_note: "⏱ Task Time: 24-Hour Format (WIB)",
            inbox_title: "Inbox & Notifications",
            inbox_sub: "Workspace collaboration invitations and message notifications.",
            inbox_empty_title: "Inbox is Empty",
            inbox_empty_sub: "No incoming messages or invitations yet.",
            team_title: "Team Members",
            team_sub: "People collaborating on workspace tasks and projects.",
            team_empty_title: "No Team Members Yet",
            team_empty_sub: "Click '+ Invite Member' to invite collaborators to your tasks.",
            docs_title: "Documents",
            docs_sub: "Reference files and documents attached to your tasks.",
            docs_empty_title: "No Documents Yet",
            docs_empty_sub: "Documents only appear if a task has an attached document. If none, this section remains empty.",
            btn_attach_doc: "＋ Attach Document",
            settings_title: "Settings",
            settings_sub: "Manage your account, appearance, profile avatar, and workspace preferences.",
            card_profile: "Account Profile",
            card_profile_sub: "Account information connected to this workspace.",
            card_avatar: "Profile Avatar",
            card_avatar_sub: "Customize your profile photo: gender, skin tone, shirt color, and accessories.",
            card_language: "Workspace Language / Bahasa",
            card_language_sub: "Choose your preferred workspace language (English is default).",
            card_appearance: "Appearance",
            card_appearance_sub: "Choose the theme that suits your eyes best.",
            card_security: "Session & Security",
            card_security_sub: "Manage your account session on this device.",
            badge_active: "Active Account",
            badge_custom: "Custom Avatar",
            info_fullname: "Full name",
            info_username: "Username",
            info_email: "Email",
            info_joined: "Joined",
            avatar_preview_title: "Live Avatar Preview",
            avatar_preview_desc: "This avatar will represent you in the dashboard, team members, and sidebar.",
            avatar_gender: "Gender / Character",
            avatar_skin: "Skin Tone (Warna Kulit)",
            avatar_shirt: "Shirt Color (Warna Baju)",
            avatar_hair: "Hair Color (Warna Rambut)",
            avatar_glasses: "Accessories / Glasses",
            avatar_bg: "Background Color",
            acc_none: "None",
            acc_glasses: "👓 Glasses",
            acc_sunglasses: "🕶 Sunglasses",
            btn_randomize: "🎲 Randomize",
            btn_save_avatar: "💾 Save Avatar",
            theme_light: "Light",
            theme_light_sub: "Bright interface with clean white background.",
            theme_dark: "Dark",
            theme_dark_sub: "Comfortable for working in low light.",
            theme_system: "Follow system",
            theme_system_sub: "Automatically matches your OS preference.",
            logout_title: "Sign out of account",
            logout_sub: "End your active session. You will need to log back in to continue.",
            sidebar_collab_title: "Collaborative Workspace",
            sidebar_collab_desc: "Invite team members and wait for approval (acc) via inbox.",
            main_menu_breadcrumb: "Main Menu",
            deadline_warning_title: "Deadline Warning: {count} tasks have passed their deadline!",
            deadline_warning_desc: "There are unfinished tasks that have exceeded their due date. Please check your tasks below.",
            done_tasks: "tasks completed",
            ongoing_tasks: "tasks in progress",
            pending_tasks: "tasks not completed",
            no_tasks_diagram: "No tasks yet",
            stat_done_sub: "{percent}% of total tasks",
            stat_total_tasks: "Total Tasks",
            stat_total_sub: "All active tasks",
            stat_team_members: "Team Members",
            stat_team_sub: "People in workspace",
            project_overview_title: "Project Overview",
            project_overview_sub: "Track work at every stage.",
            group_empty: "No tasks in this section yet.",
            filter_showing: "Currently showing status filter",
            filter_show_all: "Show all tasks",
            inbox_comm: "COMMUNICATION",
            btn_mark_all_read: "Mark all as read",
            invite_pending_acc: "Waiting for Acceptance (Pending ACC)",
            btn_acc: "✓ Accept (ACC)",
            btn_reject: "✕ Reject",
            invite_accepted: "✓ Invitation Accepted (ACC'd)",
            invite_rejected: "✕ Invitation Rejected",
            btn_view_task: "View Task →",
            modal_invite_title: "Invite team member",
            modal_invite_sub: "Invite colleagues to collaborate on Group 1 workspace tasks.",
            modal_invite_choose_task: "Choose Task / Project",
            modal_invite_select_placeholder: "-- Select Task for Collaboration --",
            modal_invite_user: "User Email or Username",
            modal_invite_user_ph: "name@example.com or registered username",
            modal_invite_flow_note: "Collaboration Flow: Invitations will arrive in your colleague's Inbox. They must click Accept (ACC) before accessing the task.",
            btn_send_invite: "＋ Send Invitation",
            btn_cancel: "Cancel",
            modal_doc_title: "Attach Document to Task",
            modal_doc_sub: "Select a task and attach a document.",
            modal_doc_choose_task: "Choose Task / Project",
            modal_doc_select_ph: "-- Select Task --",
            modal_doc_name: "Document Name",
            modal_doc_name_ph: "e.g., Project_Brief.pdf, Report.docx",
            modal_doc_or_upload: "Or Upload File",
            btn_save_doc: "Save Document",
            cal_select_month: "Select Month",
            cal_select_year: "Select Year",
            cal_day_agenda: "DAILY AGENDA",
            cal_day_sub: "List of tasks scheduled on this date (WIB).",
            btn_close: "Close",
            badge_timezone: "Timezone: WIB (UTC+7 / Asia/Jakarta)",
            months: ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
            days: ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]
        },
        id: {
            main_menu: "MENU UTAMA",
            nav_dashboard: "Dashboard",
            nav_inbox: "Inbox",
            nav_team: "Tim",
            nav_my_tasks: "Tugas Saya",
            workspace_label: "WORKSPACE",
            nav_documents: "Dokumen",
            nav_calendar: "Kalender",
            project_status: "STATUS PROYEK",
            status_pending: "Menunggu",
            status_ongoing: "Sedang Dikerjakan",
            status_completed: "Selesai",
            nav_settings: "Pengaturan",
            nav_logout: "Keluar",
            btn_new_task: "+ Tugas Baru",
            btn_invite_member: "＋ Undang Anggota",
            btn_today: "Hari Ini",
            search_placeholder: "Cari tugas (Cmd + K atau Ctrl + K)...",
            filter_all: "Semua Prioritas",
            priority_high: "Prioritas Tinggi",
            priority_medium: "Prioritas Sedang",
            priority_low: "Prioritas Rendah",
            priority_done: "Selesai",
            welcome_title: "Halo {name}, kelola proyekmu.",
            welcome_sub: "Pantau, analisis, dan rapikan tugas agar pekerjaan tim lebih terarah.",
            diagram_title: "Diagram Progres Tugas",
            diagram_sub: "Distribusi status penyelesaian seluruh tugas dalam workspace",
            total_tasks: "Total Tugas",
            stat_pending_title: "MENUNGGU",
            stat_pending_sub: "Tugas belum dimulai",
            stat_ongoing_title: "SEDANG BERJALAN",
            stat_ongoing_sub: "Tugas sedang dikerjakan",
            stat_completed_title: "SELESAI",
            stat_completed_sub: "Tugas telah diselesaikan",
            stat_deadline_title: "DEADLINE MENDATANG",
            stat_deadline_sub: "Dalam 7 hari ke depan",
            th_name: "Nama Tugas",
            th_desc: "Deskripsi",
            th_deadline: "Deadline",
            th_team: "Tim",
            th_priority: "Prioritas",
            th_actions: "Aksi",
            only_you: "Hanya kamu",
            collaborators: "kolaborator",
            created_on: "Dibuat",
            overdue_badge: "⚠️ Terlewat",
            cal_title: "Kalender & Jadwal Tugas",
            cal_time_label: "WAKTU SAAT INI (WIB):",
            legend_holiday: "Minggu / Tanggal Merah",
            legend_note: "⏱ Waktu Tugas: Format 24 Jam (WIB)",
            inbox_title: "Inbox & Notifikasi",
            inbox_sub: "Undangan kolaborasi dan pesan notifikasi workspace.",
            inbox_empty_title: "Kotak Masuk Kosong",
            inbox_empty_sub: "Belum ada pesan atau undangan masuk.",
            team_title: "Anggota Tim",
            team_sub: "Orang-orang yang berkolaborasi dalam tugas workspace.",
            team_empty_title: "Belum Ada Anggota Tim",
            team_empty_sub: "Klik '+ Undang Anggota' untuk mengundang rekan ke tugas Anda.",
            docs_title: "Dokumen",
            docs_sub: "Berkas referensi dan dokumen yang terlampir pada tugas Anda.",
            docs_empty_title: "Belum Ada Dokumen",
            docs_empty_sub: "Dokumen hanya muncul jika ada tugas yang dilampirkan berkas. Jika tidak ada, bagian ini kosong.",
            btn_attach_doc: "＋ Lampirkan Dokumen",
            settings_title: "Pengaturan",
            settings_sub: "Kelola akun, tampilan, avatar profil, dan preferensi workspace-mu.",
            card_profile: "Profil Akun",
            card_profile_sub: "Informasi akun yang terhubung dengan workspace ini.",
            card_avatar: "Foto Profil / Avatar",
            card_avatar_sub: "Sesuaikan avatar: jenis kelamin, warna kulit, warna baju, dan aksesoris.",
            card_language: "Bahasa Workspace",
            card_language_sub: "Pilih bahasa yang Anda inginkan (Bahasa Inggris bawaan).",
            card_appearance: "Tampilan",
            card_appearance_sub: "Pilih tema yang paling nyaman untuk matamu.",
            card_security: "Sesi & Keamanan",
            card_security_sub: "Kelola sesi login akun di perangkat ini.",
            badge_active: "Akun Aktif",
            badge_custom: "Avatar Kustom",
            info_fullname: "Nama lengkap",
            info_username: "Username",
            info_email: "Email",
            info_joined: "Bergabung",
            avatar_preview_title: "Pratinjau Avatar Langsung",
            avatar_preview_desc: "Avatar ini akan ditampilkan di dashboard, anggota tim, dan navigasi.",
            avatar_gender: "Jenis Kelamin / Karakter",
            avatar_skin: "Warna Kulit (Skin Tone)",
            avatar_shirt: "Warna Baju (Shirt Color)",
            avatar_hair: "Warna Rambut (Hair Color)",
            avatar_glasses: "Aksesoris / Kacamata",
            avatar_bg: "Warna Latar",
            acc_none: "Tanpa Kacamata",
            acc_glasses: "👓 Kacamata",
            acc_sunglasses: "🕶 Kacamata Hitam",
            btn_randomize: "🎲 Acak",
            btn_save_avatar: "💾 Simpan Avatar",
            theme_light: "Terang",
            theme_light_sub: "Tampilan cerah dengan latar putih bersih.",
            theme_dark: "Gelap",
            theme_dark_sub: "Nyaman digunakan di ruangan redup.",
            theme_system: "Ikuti sistem",
            theme_system_sub: "Menyesuaikan otomatis dengan sistem operasi.",
            logout_title: "Keluar dari akun",
            logout_sub: "Akhiri sesi aktif di perangkat ini. Anda perlu masuk kembali nanti.",
            sidebar_collab_title: "Workspace Kolaborasi",
            sidebar_collab_desc: "Undang anggota tim dan tunggu persetujuan (acc) melalui inbox.",
            main_menu_breadcrumb: "Menu Utama",
            deadline_warning_title: "Peringatan Deadline: {count} tugas telah melebihi batas waktu!",
            deadline_warning_desc: "Terdapat tugas yang belum selesai dan telah melewati batas tenggat waktu. Silakan periksa status tugas Anda di bawah ini.",
            done_tasks: "tugas selesai",
            ongoing_tasks: "tugas berjalan",
            pending_tasks: "tugas belum selesai",
            no_tasks_diagram: "Belum ada tugas",
            stat_done_sub: "{percent}% dari total tugas",
            stat_total_tasks: "Total Tugas",
            stat_total_sub: "Semua tugas aktif",
            stat_team_members: "Anggota Tim",
            stat_team_sub: "Orang di workspace",
            project_overview_title: "Ringkasan Proyek",
            project_overview_sub: "Lacak pekerjaan di setiap tahap.",
            group_empty: "Belum ada tugas di bagian ini.",
            filter_showing: "Sedang menampilkan filter status",
            filter_show_all: "Tampilkan semua tugas",
            inbox_comm: "KOMUNIKASI",
            btn_mark_all_read: "Tandai semua dibaca",
            invite_pending_acc: "Menunggu Persetujuan (Pending ACC)",
            btn_acc: "✓ Terima (ACC)",
            btn_reject: "✕ Tolak",
            invite_accepted: "✓ Undangan Diterima (Telah ACC)",
            invite_rejected: "✕ Undangan Ditolak",
            btn_view_task: "Lihat Tugas →",
            modal_invite_title: "Undang anggota tim",
            modal_invite_sub: "Undang rekan untuk berkolaborasi pada tugas workspace Kelompok 1.",
            modal_invite_choose_task: "Pilih Tugas / Proyek",
            modal_invite_select_placeholder: "-- Pilih Tugas untuk Kolaborasi --",
            modal_invite_user: "Email atau Username Pengguna",
            modal_invite_user_ph: "name@example.com atau username terdaftar",
            modal_invite_flow_note: "Alur Kolaborasi: Undangan kolaborasi akan masuk ke Inbox rekan Anda. Rekan Anda harus mengklik Terima (ACC) terlebih dahulu sebelum tugas dapat diakses bersama.",
            btn_send_invite: "＋ Kirim Undangan",
            btn_cancel: "Batal",
            modal_doc_title: "Lampirkan Dokumen ke Tugas",
            modal_doc_sub: "Pilih tugas dan tentukan dokumen yang ingin dilampirkan.",
            modal_doc_choose_task: "Pilih Tugas / Proyek",
            modal_doc_select_ph: "-- Pilih Tugas --",
            modal_doc_name: "Nama Dokumen",
            modal_doc_name_ph: "Contoh: Project_Brief.pdf, Laporan.docx",
            modal_doc_or_upload: "Atau Unggah Berkas",
            btn_save_doc: "Simpan Dokumen",
            cal_select_month: "Pilih Bulan",
            cal_select_year: "Pilih Tahun",
            cal_day_agenda: "AGENDA HARIAN",
            cal_day_sub: "Daftar tugas yang terjadwal pada tanggal ini (WIB).",
            btn_close: "Tutup",
            badge_timezone: "Zona Waktu: WIB (UTC+7 / Asia/Jakarta)",
            months: ["Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"],
            days: ["Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"]
        },
        ru: {
            main_menu: "ГЛАВНОЕ МЕНЮ",
            nav_dashboard: "Панель",
            nav_inbox: "Входящие",
            nav_team: "Команда",
            nav_my_tasks: "Мои задачи",
            workspace_label: "РАБОЧАЯ ЗОНА",
            nav_documents: "Документы",
            nav_calendar: "Календарь",
            project_status: "СТАТУС ПРОЕКТА",
            status_pending: "В ожидании",
            status_ongoing: "В процессе",
            status_completed: "Завершено",
            nav_settings: "Настройки",
            nav_logout: "Выйти",
            btn_new_task: "+ Новая задача",
            btn_invite_member: "＋ Пригласить",
            btn_today: "Сегодня",
            search_placeholder: "Поиск задач (Cmd + K или Ctrl + K)...",
            filter_all: "Все приоритеты",
            priority_high: "Высокий приоритет",
            priority_medium: "Средний приоритет",
            priority_low: "Низкий приоритет",
            priority_done: "Завершено",
            welcome_title: "Здравствуйте, {name}, управляйте проектами.",
            welcome_sub: "Отслеживайте, анализируйте и организуйте задачи для эффективной работы команды.",
            diagram_title: "Диаграмма прогресса задач",
            diagram_sub: "Распределение статусов выполнения всех задач",
            total_tasks: "Всего задач",
            stat_pending_title: "ОЖИДАНИЕ",
            stat_pending_sub: "Задачи, готовые к началу",
            stat_ongoing_title: "В ПРОЦЕССЕ",
            stat_ongoing_sub: "Задачи в работе",
            stat_completed_title: "ЗАВЕРШЕНО",
            stat_completed_sub: "Выполненные задачи",
            stat_deadline_title: "СКОРО ДЕДЛАЙН",
            stat_deadline_sub: "В ближайшие 7 дней",
            th_name: "Задача",
            th_desc: "Описание",
            th_deadline: "Срок",
            th_team: "Команда",
            th_priority: "Приоритет",
            th_actions: "Действия",
            only_you: "Только вы",
            collaborators: "участников",
            created_on: "Создано",
            overdue_badge: "⚠️ Просрочено",
            cal_title: "Календарь и расписание",
            cal_time_label: "ТЕКУЩЕЕ ВРЕМЯ (WIB):",
            legend_holiday: "Воскресенье / Выходной",
            legend_note: "⏱ Время: 24-часовой формат (WIB)",
            inbox_title: "Входящие и уведомления",
            inbox_sub: "Приглашения к сотрудничеству и уведомления.",
            inbox_empty_title: "Входящие пусты",
            inbox_empty_sub: "Новых сообщений или приглашений пока нет.",
            team_title: "Участники команды",
            team_sub: "Люди, участвующие в совместной работе над задачами.",
            team_empty_title: "Участников пока нет",
            team_empty_sub: "Нажмите '+ Пригласить', чтобы добавить участников.",
            docs_title: "Документы",
            docs_sub: "Файлы и материалы, прикрепленные к задачам.",
            docs_empty_title: "Документов пока нет",
            docs_empty_sub: "Документы появятся здесь, если прикрепить их к задачам.",
            btn_attach_doc: "＋ Прикрепить документ",
            settings_title: "Настройки",
            settings_sub: "Управление аккаунтом, темой, аватаром и языком.",
            card_profile: "Профиль аккаунта",
            card_profile_sub: "Информация об аккаунте в этой рабочей зоне.",
            card_avatar: "Аватар профиля",
            card_avatar_sub: "Настройте свой аватар: пол, цвет кожи, цвет одежды и стиль.",
            card_language: "Язык интерфейса",
            card_language_sub: "Выберите язык приложения (по умолчанию английский).",
            card_appearance: "Оформление",
            card_appearance_sub: "Выберите тему для удобной работы.",
            card_security: "Безопасность и сеанс",
            card_security_sub: "Управление сессией входа на этом устройстве.",
            badge_active: "Активен",
            badge_custom: "Кастомный аватар",
            info_fullname: "Полное имя",
            info_username: "Имя пользователя",
            info_email: "Эл. почта",
            info_joined: "Дата регистрации",
            avatar_preview_title: "Предпросмотр аватара",
            avatar_preview_desc: "Этот аватар будет отображаться в шапке, боковой панели и списке команды.",
            avatar_gender: "Пол / Персонаж",
            avatar_skin: "Цвет кожи",
            avatar_shirt: "Цвет одежды",
            avatar_hair: "Цвет волос",
            avatar_glasses: "Очки и аксессуары",
            avatar_bg: "Цвет фона",
            acc_none: "Без очков",
            acc_glasses: "👓 Очки",
            acc_sunglasses: "🕶 Солнечные очки",
            btn_randomize: "🎲 Случайно",
            btn_save_avatar: "💾 Сохранить аватар",
            theme_light: "Светлая",
            theme_light_sub: "Светлый интерфейс с белым фоном.",
            theme_dark: "Темная",
            theme_dark_sub: "Комфортно при слабом освещении.",
            theme_system: "Системная",
            theme_system_sub: "Соответствует настройкам системы.",
            logout_title: "Выйти из системы",
            logout_sub: "Завершить сеанс на этом устройстве.",
            sidebar_collab_title: "Совместная работа",
            sidebar_collab_desc: "Приглашайте участников и ждите подтверждения во входящих.",
            main_menu_breadcrumb: "Главное меню",
            deadline_warning_title: "Внимание: {count} задач просрочили дедлайн!",
            deadline_warning_desc: "Есть незавершенные задачи с истекшим сроком. Пожалуйста, проверьте их ниже.",
            done_tasks: "задач завершено",
            ongoing_tasks: "задач в процессе",
            pending_tasks: "задач не начато",
            no_tasks_diagram: "Задач пока нет",
            stat_done_sub: "{percent}% от общего числа",
            stat_total_tasks: "Всего задач",
            stat_total_sub: "Все активные задачи",
            stat_team_members: "Участники команды",
            stat_team_sub: "Человек в рабочей зоне",
            project_overview_title: "Обзор проекта",
            project_overview_sub: "Отслеживайте работу на каждом этапе.",
            group_empty: "В этом разделе пока нет задач.",
            filter_showing: "Отображается фильтр по статусу",
            filter_show_all: "Показать все задачи",
            inbox_comm: "КОММУНИКАЦИЯ",
            btn_mark_all_read: "Отметить все прочитанными",
            invite_pending_acc: "Ожидает подтверждения (ACC)",
            btn_acc: "✓ Принять (ACC)",
            btn_reject: "✕ Отклонить",
            invite_accepted: "✓ Приглашение принято (ACC)",
            invite_rejected: "✕ Приглашение отклонено",
            btn_view_task: "Открыть задачу →",
            modal_invite_title: "Пригласить участника команды",
            modal_invite_sub: "Пригласите коллег для совместной работы над задачами рабочей зоны.",
            modal_invite_choose_task: "Выберите задачу / проект",
            modal_invite_select_placeholder: "-- Выберите задачу для сотрудничества --",
            modal_invite_user: "Эл. почта или имя пользователя",
            modal_invite_user_ph: "name@example.com или имя пользователя",
            modal_invite_flow_note: "Процесс: Приглашение поступит во входящие коллеги. Ему необходимо нажать 'Принять (ACC)' для доступа.",
            btn_send_invite: "＋ Отправить приглашение",
            btn_cancel: "Отмена",
            modal_doc_title: "Прикрепить документ к задаче",
            modal_doc_sub: "Выберите задачу и прикрепите документ.",
            modal_doc_choose_task: "Выберите задачу / проект",
            modal_doc_select_ph: "-- Выберите задачу --",
            modal_doc_name: "Название документа",
            modal_doc_name_ph: "например, Project_Brief.pdf, Report.docx",
            modal_doc_or_upload: "Или загрузите файл",
            btn_save_doc: "Сохранить документ",
            cal_select_month: "Выберите месяц",
            cal_select_year: "Выберите год",
            cal_day_agenda: "РАСПИСАНИЕ НА ДЕНЬ",
            cal_day_sub: "Список задач, запланированных на эту дату (WIB).",
            btn_close: "Закрыть",
            badge_timezone: "Часовой пояс: WIB (UTC+7 / Джакарта)",
            months: ["Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"],
            days: ["Воскресенье", "Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота"]
        },
        zh: {
            main_menu: "主菜单",
            nav_dashboard: "仪表盘",
            nav_inbox: "收件箱",
            nav_team: "团队",
            nav_my_tasks: "我的任务",
            workspace_label: "工作空间",
            nav_documents: "文档",
            nav_calendar: "日历",
            project_status: "项目状态",
            status_pending: "待处理",
            status_ongoing: "进行中",
            status_completed: "已完成",
            nav_settings: "设置",
            nav_logout: "退出登录",
            btn_new_task: "+ 新建任务",
            btn_invite_member: "＋ 邀请成员",
            btn_today: "今天",
            search_placeholder: "搜索任务 (Cmd + K 或 Ctrl + K)...",
            filter_all: "所有优先级",
            priority_high: "高优先级",
            priority_medium: "中优先级",
            priority_low: "低优先级",
            priority_done: "已完成",
            welcome_title: "你好 {name}，管理你的项目。",
            welcome_sub: "跟踪、分析和整理任务，让团队协作更专注高效。",
            diagram_title: "任务进度图表",
            diagram_sub: "工作区所有任务的状态分布情况",
            total_tasks: "全部任务",
            stat_pending_title: "待处理",
            stat_pending_sub: "等待开始的任务",
            stat_ongoing_title: "进行中",
            stat_ongoing_sub: "正在处理的任务",
            stat_completed_title: "已完成",
            stat_completed_sub: "已完成的任务",
            stat_deadline_title: "临近截止",
            stat_deadline_sub: "未来 7 天内",
            th_name: "任务名称",
            th_desc: "描述",
            th_deadline: "截止时间",
            th_team: "团队",
            th_priority: "优先级",
            th_actions: "操作",
            only_you: "仅你一人",
            collaborators: "位协作者",
            created_on: "创建于",
            overdue_badge: "⚠️ 已逾期",
            cal_title: "日历与任务安排",
            cal_time_label: "当前时间 (WIB):",
            legend_holiday: "周日 / 法定节假日",
            legend_note: "⏱ 任务时间：24小时制 (WIB)",
            inbox_title: "收件箱与通知",
            inbox_sub: "协作邀请与工作区通知消息。",
            inbox_empty_title: "收件箱为空",
            inbox_empty_sub: "暂无新消息或协作邀请。",
            team_title: "团队成员",
            team_sub: "参与工作区任务协作的成员。",
            team_empty_title: "暂无团队成员",
            team_empty_sub: "点击“+ 邀请成员”将同事添加到任务中。",
            docs_title: "工作区文档",
            docs_sub: "任务附带的参考资料与文件。",
            docs_empty_title: "暂无文档",
            docs_empty_sub: "只有在任务附加文件后，此区域才会显示文档。",
            btn_attach_doc: "＋ 附加文档",
            settings_title: "设置",
            settings_sub: "管理您的账户、外观、头像和工作区偏好设置。",
            card_profile: "账户资料",
            card_profile_sub: "连接到此工作区的账户信息。",
            card_avatar: "个人头像",
            card_avatar_sub: "自定义您的头像：性别、肤色、服装颜色与配件风格。",
            card_language: "工作区语言",
            card_language_sub: "选择您偏好的语言（默认为英文）。",
            card_appearance: "外观主题",
            card_appearance_sub: "选择最适合您的界面主题。",
            card_security: "登录与安全",
            card_security_sub: "管理此设备上的当前会话。",
            badge_active: "已激活",
            badge_custom: "自定义头像",
            info_fullname: "姓名",
            info_username: "用户名",
            info_email: "电子邮箱",
            info_joined: "注册时间",
            avatar_preview_title: "头像实时预览",
            avatar_preview_desc: "该头像将在仪表盘、侧边栏和团队列表中展示。",
            avatar_gender: "性别 / 角色",
            avatar_skin: "肤色",
            avatar_shirt: "服装颜色",
            avatar_hair: "发色",
            avatar_glasses: "配件 / 眼镜",
            avatar_bg: "背景颜色",
            acc_none: "无眼镜",
            acc_glasses: "👓 普通眼镜",
            acc_sunglasses: "🕶 太阳镜",
            btn_randomize: "🎲 随机生成",
            btn_save_avatar: "💾 保存头像",
            theme_light: "浅色",
            theme_light_sub: "清爽明亮的白色背景。",
            theme_dark: "深色",
            theme_dark_sub: "弱光环境下更舒适护眼。",
            theme_system: "跟随系统",
            theme_system_sub: "根据操作系统设置自动切换。",
            logout_title: "退出账户",
            logout_sub: "结束此设备上的会话，下次需要重新登录。",
            sidebar_collab_title: "协作工作区",
            sidebar_collab_desc: "邀请团队成员并通过收件箱等待批准。",
            main_menu_breadcrumb: "主菜单",
            deadline_warning_title: "截止日期提醒：{count} 个任务已超过截止时间！",
            deadline_warning_desc: "有未完成的任务已超过截止日期。请在下方检查您的任务状态。",
            done_tasks: "个任务已完成",
            ongoing_tasks: "个任务进行中",
            pending_tasks: "个任务待完成",
            no_tasks_diagram: "暂无任务",
            stat_done_sub: "占总任务的 {percent}%",
            stat_total_tasks: "全部任务",
            stat_total_sub: "所有活动任务",
            stat_team_members: "团队成员",
            stat_team_sub: "工作区成员",
            project_overview_title: "项目概览",
            project_overview_sub: "跟踪各个阶段的工作进度。",
            group_empty: "此部分暂无任务。",
            filter_showing: "正在显示状态筛选",
            filter_show_all: "显示全部任务",
            inbox_comm: "沟通",
            btn_mark_all_read: "全部标为已读",
            invite_pending_acc: "等待同意 (待确认)",
            btn_acc: "✓ 接受 (ACC)",
            btn_reject: "✕ 拒绝",
            invite_accepted: "✓ 邀请已接受",
            invite_rejected: "✕ 邀请已拒绝",
            btn_view_task: "查看任务 →",
            modal_invite_title: "邀请团队成员",
            modal_invite_sub: "邀请同事协作处理第1组工作区的任务。",
            modal_invite_choose_task: "选择任务 / 项目",
            modal_invite_select_placeholder: "-- 选择要协作的任务 --",
            modal_invite_user: "用户电子邮箱或用户名",
            modal_invite_user_ph: "name@example.com 或已注册用户名",
            modal_invite_flow_note: "协作流程：邀请将发送到同事的收件箱中。他们必须点击“接受 (ACC)”后才能共同访问任务。",
            btn_send_invite: "＋ 发送邀请",
            btn_cancel: "取消",
            modal_doc_title: "将文档附加到任务",
            modal_doc_sub: "选择任务并指定要附加的文档。",
            modal_doc_choose_task: "选择任务 / 项目",
            modal_doc_select_ph: "-- 选择任务 --",
            modal_doc_name: "文档名称",
            modal_doc_name_ph: "例如：Project_Brief.pdf, Report.docx",
            modal_doc_or_upload: "或上传文件",
            btn_save_doc: "保存文档",
            cal_select_month: "选择月份",
            cal_select_year: "选择年份",
            cal_day_agenda: "每日日程",
            cal_day_sub: "此日期安排的任务列表 (WIB)。",
            btn_close: "关闭",
            badge_timezone: "时区：WIB (UTC+7 / 雅加达)",
            months: ["一月", "二月", "三月", "四月", "五月", "六月", "七月", "八月", "九月", "十月", "十一月", "十二月"],
            days: ["星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"]
        }
    };

    var currentLang = "en";
    try {
        currentLang = localStorage.getItem("app_lang") || "en";
    } catch (e) {
        currentLang = "en";
    }

    function t(key) {
        var dict = i18n[currentLang] || i18n.en;
        return dict[key] || (i18n.en && i18n.en[key]) || key;
    }

    function setLanguage(lang) {
        if (!i18n[lang]) lang = "en";
        currentLang = lang;
        try {
            localStorage.setItem("app_lang", lang);
            document.cookie = "app_lang=" + lang + ";path=/;max-age=31536000";
        } catch (e) {}

        document.documentElement.lang = lang;

        // Update all data-i18n elements
        document.querySelectorAll("[data-i18n]").forEach(function (el) {
            var k = el.getAttribute("data-i18n");
            var val = t(k);
            if (val) {
                if (el.hasAttribute("data-name")) {
                    val = val.replace("{name}", el.getAttribute("data-name") || "");
                }
                if (el.hasAttribute("data-count")) {
                    val = val.replace("{count}", el.getAttribute("data-count") || "");
                }
                if (el.hasAttribute("data-pct")) {
                    val = val.replace("{percent}", el.getAttribute("data-pct") || "");
                }
                el.textContent = val;
            }
        });

        // Update placeholders
        document.querySelectorAll("[data-i18n-placeholder]").forEach(function (el) {
            var k = el.getAttribute("data-i18n-placeholder");
            var val = t(k);
            if (val) el.placeholder = val;
        });

        // Update titles
        document.querySelectorAll("[data-i18n-title]").forEach(function (el) {
            var k = el.getAttribute("data-i18n-title");
            var val = t(k);
            if (val) el.title = val;
        });

        // Sync dropdowns & radio buttons
        document.querySelectorAll("[data-lang-select]").forEach(function (sel) {
            sel.value = lang;
        });
        document.querySelectorAll("[data-lang-option]").forEach(function (radio) {
            radio.checked = (radio.value === lang);
        });

        var activeBadge = document.getElementById("activeLangBadge");
        if (activeBadge) {
            activeBadge.textContent = lang === "en" ? "EN (Default)" : lang.toUpperCase();
        }

        // Update breadcrumb titles
        if (breadcrumbTitle) {
            var activeP = "dashboard";
            for (var key in pageViews) {
                if (pageViews[key] && pageViews[key].classList.contains("active")) {
                    activeP = key;
                    break;
                }
            }
            var bMap = {
                dashboard: t("nav_dashboard"),
                inbox: t("nav_inbox"),
                team: t("nav_team"),
                documents: t("nav_documents"),
                calendar: t("nav_calendar")
            };
            if (bMap[activeP]) breadcrumbTitle.textContent = bMap[activeP];
        }

        // Re-render calendar to reflect month and day names
        renderCalendar();
    }

    // Attach listener for language selectors
    document.querySelectorAll("[data-lang-select]").forEach(function (sel) {
        sel.addEventListener("change", function () {
            setLanguage(this.value);
            // Optionally persist to server
            fetch("/user/preferences", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: "language=" + encodeURIComponent(this.value)
            }).catch(function () {});
        });
    });

    document.querySelectorAll("[data-lang-option]").forEach(function (radio) {
        radio.addEventListener("change", function () {
            if (this.checked) {
                setLanguage(this.value);
                fetch("/user/preferences", {
                    method: "POST",
                    headers: { "Content-Type": "application/x-www-form-urlencoded" },
                    body: "language=" + encodeURIComponent(this.value)
                }).catch(function () {});
            }
        });
    });

    /* ============================================================
       Profile Avatar SVG Generator & Builder
       ============================================================ */
    function generateAvatarSVG(opt) {
        var gender = opt.gender || "male";
        var skin = opt.skin || "#FFDFC4";
        var shirt = opt.shirt || "#2563EB";
        var hairColor = opt.hairColor || "#18181B";
        var bg = opt.bg || "#DBEAFE";
        var glasses = opt.glasses || "none";

        var svg = '<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">';
        // 1. Background
        svg += '<circle cx="50" cy="50" r="50" fill="' + bg + '"/>';

        // 2. Female back hair
        if (gender === "female") {
            svg += '<path d="M 23 45 C 21 72, 25 86, 33 96 L 67 96 C 75 86, 79 72, 77 45 C 75 22, 25 22, 23 45 Z" fill="' + hairColor + '"/>';
        }

        // 3. Shirt / Shoulders
        if (gender === "female") {
            svg += '<path d="M 20 100 C 20 76, 32 68, 50 68 C 68 68, 80 76, 80 100 Z" fill="' + shirt + '"/>';
            svg += '<path d="M 44 68 L 50 77 L 56 68 Z" fill="' + skin + '"/>';
        } else {
            svg += '<path d="M 18 100 C 18 78, 30 70, 50 70 C 70 70, 82 78, 82 100 Z" fill="' + shirt + '"/>';
            svg += '<path d="M 45 70 L 50 76 L 55 70 Z" fill="' + skin + '"/>';
        }

        // 4. Neck
        svg += '<rect x="44" y="60" width="12" height="12" rx="2" fill="' + skin + '"/>';
        svg += '<path d="M 44 62 C 47 66, 53 66, 56 62 Z" fill="rgba(0,0,0,0.08)"/>';

        // 5. Ears
        svg += '<circle cx="28" cy="47" r="4.5" fill="' + skin + '"/>';
        svg += '<circle cx="72" cy="47" r="4.5" fill="' + skin + '"/>';
        svg += '<circle cx="28.5" cy="47" r="2" fill="rgba(0,0,0,0.06)"/>';
        svg += '<circle cx="71.5" cy="47" r="2" fill="rgba(0,0,0,0.06)"/>';

        // 6. Face
        svg += '<ellipse cx="50" cy="46" rx="20" ry="24" fill="' + skin + '"/>';

        // 7. Cheeks blush
        svg += '<circle cx="36" cy="51" r="3.2" fill="rgba(244,63,94,0.18)"/>';
        svg += '<circle cx="64" cy="51" r="3.2" fill="rgba(244,63,94,0.18)"/>';

        // 8. Eyes & Eyelashes
        svg += '<circle cx="43" cy="46" r="2.4" fill="#1e293b"/>';
        svg += '<circle cx="57" cy="46" r="2.4" fill="#1e293b"/>';
        svg += '<circle cx="43.8" cy="45.2" r="0.8" fill="#ffffff"/>';
        svg += '<circle cx="57.8" cy="45.2" r="0.8" fill="#ffffff"/>';
        if (gender === "female") {
            svg += '<path d="M 40 44 L 38 42 M 60 44 L 62 42" stroke="#1e293b" stroke-width="1.2" stroke-linecap="round"/>';
        }

        // 9. Eyebrows
        if (gender === "female") {
            svg += '<path d="M 39 40 Q 43 38 47 40" stroke="' + hairColor + '" stroke-width="1.5" fill="none" stroke-linecap="round"/>';
            svg += '<path d="M 53 40 Q 57 38 61 40" stroke="' + hairColor + '" stroke-width="1.5" fill="none" stroke-linecap="round"/>';
        } else {
            svg += '<path d="M 39 41 Q 43 39 47 41" stroke="' + hairColor + '" stroke-width="2" fill="none" stroke-linecap="round"/>';
            svg += '<path d="M 53 41 Q 57 39 61 41" stroke="' + hairColor + '" stroke-width="2" fill="none" stroke-linecap="round"/>';
        }

        // 10. Nose & Smile
        svg += '<path d="M 49 48 L 50 51 L 52 51" stroke="rgba(0,0,0,0.18)" stroke-width="1.2" fill="none" stroke-linecap="round"/>';
        if (gender === "female") {
            svg += '<path d="M 44 55 Q 50 60 56 55" stroke="#be123c" stroke-width="2" fill="#fda4af" stroke-linecap="round"/>';
        } else {
            svg += '<path d="M 45 55 Q 50 59 55 55" stroke="#991b1b" stroke-width="1.8" fill="none" stroke-linecap="round"/>';
        }

        // 11. Hair Front / Bangs
        if (gender === "female") {
            svg += '<path d="M 27 38 C 29 23, 40 19, 50 19 C 60 19, 71 23, 73 38 C 68 31, 58 29, 50 29 C 42 29, 32 31, 27 38 Z" fill="' + hairColor + '"/>';
            svg += '<path d="M 27 38 C 26 48, 27 60, 32 70 C 30 60, 29 48, 31 39 Z" fill="' + hairColor + '"/>';
            svg += '<path d="M 73 38 C 74 48, 73 60, 68 70 C 70 60, 71 48, 69 39 Z" fill="' + hairColor + '"/>';
        } else {
            svg += '<path d="M 27 40 C 27 24, 38 20, 50 20 C 62 20, 73 24, 73 40 C 73 35, 69 30, 63 29 C 57 28, 43 27, 33 33 C 29 35, 27 38, 27 40 Z" fill="' + hairColor + '"/>';
            svg += '<path d="M 28 42 L 28 47 L 31 44 Z" fill="' + hairColor + '"/>';
            svg += '<path d="M 72 42 L 72 47 L 69 44 Z" fill="' + hairColor + '"/>';
        }

        // 12. Glasses / Sunglasses
        if (glasses === "glasses") {
            svg += '<rect x="35" y="40" width="13" height="11" rx="3" fill="none" stroke="#1f2937" stroke-width="1.6"/>';
            svg += '<rect x="52" y="40" width="13" height="11" rx="3" fill="none" stroke="#1f2937" stroke-width="1.6"/>';
            svg += '<line x1="48" y1="45" x2="52" y2="45" stroke="#1f2937" stroke-width="1.6"/>';
            svg += '<line x1="31" y1="43" x2="35" y2="44" stroke="#1f2937" stroke-width="1.3"/>';
            svg += '<line x1="65" y1="44" x2="69" y2="43" stroke="#1f2937" stroke-width="1.3"/>';
        } else if (glasses === "sunglasses") {
            svg += '<rect x="34" y="39" width="14" height="12" rx="3" fill="#111827" stroke="#000" stroke-width="1.5"/>';
            svg += '<rect x="52" y="39" width="14" height="12" rx="3" fill="#111827" stroke="#000" stroke-width="1.5"/>';
            svg += '<line x1="48" y1="44" x2="52" y2="44" stroke="#000" stroke-width="1.8"/>';
            svg += '<line x1="30" y1="42" x2="34" y2="43" stroke="#000" stroke-width="1.4"/>';
            svg += '<line x1="66" y1="43" x2="70" y2="42" stroke="#000" stroke-width="1.4"/>';
        }

        svg += '</svg>';
        return svg;
    }

    var avatarConfig = {
        gender: "male",
        skin: "#FFDFC4",
        shirt: "#2563EB",
        hairColor: "#18181B",
        bg: "#DBEAFE",
        glasses: "none"
    };

    try {
        var savedCfg = localStorage.getItem("user_avatar_config");
        if (savedCfg) {
            avatarConfig = Object.assign(avatarConfig, JSON.parse(savedCfg));
        }
    } catch (e) {}

    function updateAvatarElements(svgStr) {
        if (!svgStr) return;
        document.querySelectorAll(".user-avatar-container, .ws-avatar, #settingsProfileAvatar").forEach(function (el) {
            el.innerHTML = svgStr;
        });
    }

    // Load saved avatar on initial startup
    try {
        var storedAvatar = localStorage.getItem("user_avatar");
        if (storedAvatar && storedAvatar.indexOf("<svg") !== -1) {
            updateAvatarElements(storedAvatar);
        }
    } catch (e) {}

    function renderAvatarPreview() {
        var previewBox = document.getElementById("avatarLivePreview");
        if (!previewBox) return;
        var svg = generateAvatarSVG(avatarConfig);
        previewBox.innerHTML = svg;
    }

    // Avatar Builder UI logic
    var avatarBuilder = document.getElementById("avatar-builder");
    if (avatarBuilder) {
        renderAvatarPreview();

        // Gender buttons
        avatarBuilder.querySelectorAll("[data-avatar-gender]").forEach(function (btn) {
            btn.addEventListener("click", function () {
                avatarBuilder.querySelectorAll("[data-avatar-gender]").forEach(function (b) { b.classList.remove("active"); });
                btn.classList.add("active");
                avatarConfig.gender = btn.getAttribute("data-avatar-gender");
                renderAvatarPreview();
            });
        });

        // Skin Swatches
        avatarBuilder.querySelectorAll("[data-skin]").forEach(function (btn) {
            btn.addEventListener("click", function () {
                avatarBuilder.querySelectorAll("[data-skin]").forEach(function (b) { b.classList.remove("active"); });
                btn.classList.add("active");
                avatarConfig.skin = btn.getAttribute("data-skin");
                renderAvatarPreview();
            });
        });

        // Shirt Swatches
        avatarBuilder.querySelectorAll("[data-shirt]").forEach(function (btn) {
            btn.addEventListener("click", function () {
                avatarBuilder.querySelectorAll("[data-shirt]").forEach(function (b) { b.classList.remove("active"); });
                btn.classList.add("active");
                avatarConfig.shirt = btn.getAttribute("data-shirt");
                renderAvatarPreview();
            });
        });

        // Hair Swatches
        avatarBuilder.querySelectorAll("[data-hair]").forEach(function (btn) {
            btn.addEventListener("click", function () {
                avatarBuilder.querySelectorAll("[data-hair]").forEach(function (b) { b.classList.remove("active"); });
                btn.classList.add("active");
                avatarConfig.hairColor = btn.getAttribute("data-hair");
                renderAvatarPreview();
            });
        });

        // Glasses buttons
        avatarBuilder.querySelectorAll("[data-glasses]").forEach(function (btn) {
            btn.addEventListener("click", function () {
                avatarBuilder.querySelectorAll("[data-glasses]").forEach(function (b) { b.classList.remove("active"); });
                btn.classList.add("active");
                avatarConfig.glasses = btn.getAttribute("data-glasses");
                renderAvatarPreview();
            });
        });

        // Background Swatches
        avatarBuilder.querySelectorAll("[data-bg]").forEach(function (btn) {
            btn.addEventListener("click", function () {
                avatarBuilder.querySelectorAll("[data-bg]").forEach(function (b) { b.classList.remove("active"); });
                btn.classList.add("active");
                avatarConfig.bg = btn.getAttribute("data-bg");
                renderAvatarPreview();
            });
        });

        // Randomize button
        var randomizeBtn = document.getElementById("randomizeAvatarBtn");
        if (randomizeBtn) {
            var skins = ["#FFDFC4", "#F0C8A0", "#DDA876", "#B97C4B", "#7D4B2A"];
            var shirts = ["#2563EB", "#059669", "#7C3AED", "#DC2626", "#1F2937", "#D97706", "#E11D48", "#0891B2"];
            var hairs = ["#18181B", "#452613", "#78350F", "#D9A05B", "#B45309", "#94A3B8"];
            var bgs = ["#DBEAFE", "#D1FAE5", "#EDE9FE", "#FFEDD5", "#E2E8F0"];
            var glassesOpts = ["none", "glasses", "sunglasses"];
            var genders = ["male", "female"];

            randomizeBtn.addEventListener("click", function () {
                avatarConfig.gender = genders[Math.floor(Math.random() * genders.length)];
                avatarConfig.skin = skins[Math.floor(Math.random() * skins.length)];
                avatarConfig.shirt = shirts[Math.floor(Math.random() * shirts.length)];
                avatarConfig.hairColor = hairs[Math.floor(Math.random() * hairs.length)];
                avatarConfig.bg = bgs[Math.floor(Math.random() * bgs.length)];
                avatarConfig.glasses = glassesOpts[Math.floor(Math.random() * glassesOpts.length)];

                // Update UI active states
                avatarBuilder.querySelectorAll("[data-avatar-gender]").forEach(function (b) {
                    b.classList.toggle("active", b.getAttribute("data-avatar-gender") === avatarConfig.gender);
                });
                avatarBuilder.querySelectorAll("[data-skin]").forEach(function (b) {
                    b.classList.toggle("active", b.getAttribute("data-skin") === avatarConfig.skin);
                });
                avatarBuilder.querySelectorAll("[data-shirt]").forEach(function (b) {
                    b.classList.toggle("active", b.getAttribute("data-shirt") === avatarConfig.shirt);
                });
                avatarBuilder.querySelectorAll("[data-hair]").forEach(function (b) {
                    b.classList.toggle("active", b.getAttribute("data-hair") === avatarConfig.hairColor);
                });
                avatarBuilder.querySelectorAll("[data-glasses]").forEach(function (b) {
                    b.classList.toggle("active", b.getAttribute("data-glasses") === avatarConfig.glasses);
                });
                avatarBuilder.querySelectorAll("[data-bg]").forEach(function (b) {
                    b.classList.toggle("active", b.getAttribute("data-bg") === avatarConfig.bg);
                });

                renderAvatarPreview();
            });
        }

        // Save Avatar button
        var saveAvatarBtn = document.getElementById("saveAvatarBtn");
        if (saveAvatarBtn) {
            saveAvatarBtn.addEventListener("click", function () {
                var svg = generateAvatarSVG(avatarConfig);
                try {
                    localStorage.setItem("user_avatar_config", JSON.stringify(avatarConfig));
                    localStorage.setItem("user_avatar", svg);
                } catch (e) {}

                updateAvatarElements(svg);

                // Send to backend
                fetch("/user/preferences", {
                    method: "POST",
                    headers: { "Content-Type": "application/x-www-form-urlencoded" },
                    body: "avatar=" + encodeURIComponent(svg)
                }).catch(function () {});

                var badge = document.getElementById("avatarStatusBadge");
                if (badge) {
                    badge.style.display = "inline-block";
                    badge.textContent = "✓ " + (t("badge_saved") || "Saved");
                    setTimeout(function () { badge.style.display = "none"; }, 3000);
                }
            });
        }
    }

    /* ---------------- Calendar Dynamic Rendering ---------------- */
    var calendarDate = new Date();
    calendarDate.setDate(1);
    var selectedDateKey = null;

    var calendarEvents = [];
    var jsonEl = document.getElementById("calendarEventsJson");
    if (jsonEl && jsonEl.textContent) {
        try {
            calendarEvents = JSON.parse(jsonEl.textContent.trim());
        } catch (e) {
            console.warn("Failed to parse calendarEventsJson", e);
        }
    }
    if ((!calendarEvents || calendarEvents.length === 0) && window.calendarEventsData) {
        if (Array.isArray(window.calendarEventsData)) {
            calendarEvents = window.calendarEventsData;
        } else if (typeof window.calendarEventsData === "string") {
            try {
                calendarEvents = JSON.parse(window.calendarEventsData);
            } catch (e) {}
        }
    }
    if (typeof calendarEvents === "string") {
        try {
            calendarEvents = JSON.parse(calendarEvents);
        } catch (e) {
            calendarEvents = [];
        }
    }

    // Nama-nama Bulan Bahasa Indonesia (Januari, Februari, Maret, April, dst.)
    var monthNamesIndo = [
        "Januari", "Februari", "Maret", "April", "Mei", "Juni",
        "Juli", "Agustus", "September", "Oktober", "November", "Desember"
    ];

    var dayNamesIndoLong = [
        "Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"
    ];

    // Daftar Hari Libur Nasional Indonesia (Tanggal Merah Nasional)
    var indoHolidays = {
        "01-01": "Tahun Baru Masehi",
        "05-01": "Hari Buruh Internasional",
        "06-01": "Hari Lahir Pancasila",
        "08-17": "Hari Kemerdekaan RI",
        "12-25": "Hari Raya Natal",
        // Hari Libur Nasional Spesifik 2026:
        "2026-01-01": "Tahun Baru 2026",
        "2026-02-17": "Isra Mikraj",
        "2026-02-18": "Tahun Baru Imlek",
        "2026-03-20": "Hari Suci Nyepi",
        "2026-03-21": "Hari Raya Idul Fitri",
        "2026-03-22": "Hari Raya Idul Fitri",
        "2026-04-03": "Wafat Yesus Kristus",
        "2026-05-01": "Hari Buruh",
        "2026-05-14": "Kenaikan Yesus Kristus",
        "2026-05-27": "Hari Raya Idul Adha",
        "2026-05-31": "Hari Raya Waisak",
        "2026-06-01": "Hari Lahir Pancasila",
        "2026-06-16": "Tahun Baru Islam",
        "2026-08-17": "HUT Proklamasi RI",
        "2026-08-25": "Maulid Nabi Muhammad",
        "2026-12-25": "Hari Raya Natal"
    };

    function getIndoHoliday(dateKey) {
        if (!dateKey) return null;
        if (indoHolidays[dateKey]) return indoHolidays[dateKey];
        var mmdd = dateKey.substring(5);
        if (indoHolidays[mmdd]) return indoHolidays[mmdd];
        return null;
    }

    function getMonthNames() {
        return (i18n[currentLang] && i18n[currentLang].months) || i18n.en.months;
    }
    function getDayNames() {
        return (i18n[currentLang] && i18n[currentLang].days) || i18n.en.days;
    }

    function formatIndoFullDate(dateStr) {
        if (!dateStr) return "";
        var parts = dateStr.split("-");
        if (parts.length !== 3) return dateStr;
        var y = parseInt(parts[0], 10);
        var m = parseInt(parts[1], 10) - 1;
        var d = parseInt(parts[2], 10);
        var dt = new Date(y, m, d);
        var dayNames = getDayNames();
        var monthNames = getMonthNames();
        var dayName = dayNames[dt.getDay()] || "";
        return dayName + ", " + d + " " + (monthNames[m] || "") + " " + y;
    }

    // Jam Real-Time
    function updateIndoClock() {
        var clockEl = document.getElementById("liveIndoTime");
        if (!clockEl) return;
        var now = new Date();
        var loc = "en-US";
        if (currentLang === "id") loc = "id-ID";
        else if (currentLang === "ru") loc = "ru-RU";
        else if (currentLang === "zh") loc = "zh-CN";

        try {
            var options = {
                timeZone: "Asia/Jakarta",
                weekday: "long",
                year: "numeric",
                month: "long",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
                second: "2-digit",
                hour12: false
            };
            var formatter = new Intl.DateTimeFormat(loc, options);
            var tzSuffix = currentLang === "id" ? " WIB" : " (UTC+7)";
            clockEl.textContent = formatter.format(now).replace(/\./g, ":") + tzSuffix;
        } catch (e) {
            var dayNames = getDayNames();
            var monthNames = getMonthNames();
            var dName = dayNames[now.getDay()];
            var d = now.getDate();
            var m = monthNames[now.getMonth()];
            var y = now.getFullYear();
            var h = String(now.getHours()).padStart(2, "0");
            var mi = String(now.getMinutes()).padStart(2, "0");
            var s = String(now.getSeconds()).padStart(2, "0");
            clockEl.textContent = dName + ", " + d + " " + m + " " + y + " • " + h + ":" + mi + ":" + s + " (UTC+7)";
        }
    }
    setInterval(updateIndoClock, 1000);
    updateIndoClock();

    function renderDayDetail() {
        var detailCard = document.getElementById("calDayDetailCard");
        var titleEl = document.getElementById("dayDetailTitle");
        var subEl = document.getElementById("dayDetailSub");
        var listEl = document.getElementById("dayDetailList");
        if (!detailCard || !listEl) return;

        if (!selectedDateKey) {
            detailCard.style.display = "none";
            return;
        }

        detailCard.style.display = "block";
        var fullDateStr = formatIndoFullDate(selectedDateKey);
        if (titleEl) {
            var hol = getIndoHoliday(selectedDateKey);
            titleEl.textContent = fullDateStr + (hol ? " — (" + hol + ")" : "");
        }

        var matched = calendarEvents.filter(function (e) {
            return e.date === selectedDateKey;
        });

        if (subEl) {
            subEl.textContent = matched.length + " " + (t("tasks_scheduled") || "tasks scheduled on this date.");
        }

        if (matched.length === 0) {
            listEl.innerHTML = '<div class="day-detail-empty">' +
                '<p>📅 ' + (t("no_tasks_day") || "No tasks scheduled on") + ' ' + fullDateStr + '.</p>' +
                '</div>';
            return;
        }

        var html = "";
        matched.forEach(function (ev) {
            var statusLabel = ev.status_label || (ev.status === "sudah" ? t("status_completed") : ev.status === "sedang" ? t("status_ongoing") : t("status_pending"));
            var statusBadgeClass = ev.status === "sudah" ? "tag-completed" : ev.status === "sedang" ? "tag-progress" : "tag-pending";
            var timeStr = ev.time ? ev.time : "All day";

            html += '<article class="day-detail-item">' +
                '<div class="day-detail-info">' +
                    '<strong><a href="/todo/detail?id=' + ev.todo_id + '">' + (ev.title || "Untitled Task") + '</a></strong>' +
                    '<small>⏱ ' + timeStr + ' · ' + (ev.is_deadline ? t("th_deadline") : t("created_on")) + '</small>' +
                '</div>' +
                '<div class="day-detail-badges">' +
                    '<span class="tag ' + statusBadgeClass + '">' + statusLabel + '</span>' +
                    '<a href="/todo/detail?id=' + ev.todo_id + '" class="secondary-btn small-btn">' + (t("th_actions") || "Detail") + ' →</a>' +
                '</div>' +
            '</article>';
        });

        listEl.innerHTML = html;
    }

    function renderCalendar() {
        var grid = document.getElementById("calendarGrid");
        var titleEl = document.getElementById("calendarMonthTitle");
        var monthSelect = document.getElementById("calMonthSelect");
        var yearSelect = document.getElementById("calYearSelect");
        if (!grid) return;

        var y = calendarDate.getFullYear();
        var m = calendarDate.getMonth();
        var monthNames = getMonthNames();
        var dayNames = getDayNames();

        // Sinkronisasi teks judul dan dropdown selector
        if (titleEl) {
            titleEl.textContent = monthNames[m] + " " + y;
        }
        if (monthSelect) {
            monthSelect.innerHTML = monthNames.map(function (name, idx) {
                return '<option value="' + idx + '"' + (idx === m ? ' selected' : '') + '>' + name + '</option>';
            }).join("");
            monthSelect.value = m;
        }
        if (yearSelect) {
            yearSelect.value = y;
        }

        // Hitung hari awal dan total hari dalam bulan
        var start = new Date(y, m, 1).getDay(); // 0 = Minggu/Sunday, 1 = Senin/Monday, ...
        var days = new Date(y, m + 1, 0).getDate();
        var prevDays = new Date(y, m, 0).getDate();

        var html = dayNames.map(function (d, idx) {
            return '<div class="calendar-day-name' + (idx === 0 ? ' sunday' : '') + '">' + d + '</div>';
        }).join("");

        var totalCells = Math.ceil((start + days) / 7) * 7;
        var now = new Date();

        for (var i = 0; i < totalCells; i++) {
            var dayNum;
            var isMuted = false;
            var cellMonth = m + 1;
            var cellYear = y;
            var colIndex = i % 7;
            var isSunday = (colIndex === 0);

            if (i < start) {
                dayNum = prevDays - start + i + 1;
                isMuted = true;
                cellMonth = m;
                if (cellMonth === 0) { cellMonth = 12; cellYear = y - 1; }
            } else if (i >= start + days) {
                dayNum = i - start - days + 1;
                isMuted = true;
                cellMonth = m + 2;
                if (cellMonth === 13) { cellMonth = 1; cellYear = y + 1; }
            } else {
                dayNum = i - start + 1;
            }

            var dateKey = cellYear + "-" + String(cellMonth).padStart(2, "0") + "-" + String(dayNum).padStart(2, "0");
            var isToday = !isMuted && now.getFullYear() === y && now.getMonth() === m && now.getDate() === dayNum;
            var isSelected = (selectedDateKey === dateKey);

            var holidayName = !isMuted ? getIndoHoliday(dateKey) : null;

            var matched = calendarEvents.filter(function (e) {
                return e.date === dateKey;
            });

            var classes = "calendar-day";
            if (isMuted) classes += " muted-day";
            if (isSunday) classes += " sunday";
            if (holidayName) classes += " holiday";
            if (isToday) classes += " today";
            if (isSelected) classes += " selected-day";

            html += '<div class="' + classes + '" data-date="' + dateKey + '">';
            html += '<div class="calendar-day-header">';
            html += '<span class="calendar-day-num">' + dayNum + '</span>';
            html += '</div>';

            if (holidayName) {
                html += '<span class="holiday-tag" title="' + holidayName + '">' + holidayName + '</span>';
            }

            html += '<div class="calendar-events-wrap">';
            matched.forEach(function (ev) {
                var statusClass = ev.status_class || ev.status || "belum";
                var timeBadge = ev.time ? '<span class="ev-time">' + ev.time + '</span>' : '';
                var tooltip = (ev.title || "") + " (" + (ev.time || "WIB") + " - " + (ev.status_label || statusClass) + ")";
                html += '<a href="/todo/detail?id=' + ev.todo_id + '" class="calendar-event ' + statusClass + '" title="' + tooltip + '">' +
                    timeBadge +
                    '<span class="ev-title">' + (ev.title || "") + '</span>' +
                '</a>';
            });
            html += '</div>';

            html += '</div>';
        }

        grid.innerHTML = html;

        // Event listener klik pada sel tanggal
        grid.querySelectorAll(".calendar-day").forEach(function (dayEl) {
            dayEl.addEventListener("click", function (evt) {
                if (evt.target.closest("a")) return;
                var clickedDate = dayEl.getAttribute("data-date");
                if (selectedDateKey === clickedDate) {
                    selectedDateKey = null; // Toggle off jika diklik lagi
                } else {
                    selectedDateKey = clickedDate;
                }
                renderCalendar();
            });
        });

        renderDayDetail();
    }

    var calMonthSelect = document.getElementById("calMonthSelect");
    if (calMonthSelect) {
        calMonthSelect.addEventListener("change", function () {
            calendarDate.setMonth(parseInt(this.value, 10));
            selectedDateKey = null;
            renderCalendar();
        });
    }

    var calYearSelect = document.getElementById("calYearSelect");
    if (calYearSelect) {
        calYearSelect.addEventListener("change", function () {
            calendarDate.setFullYear(parseInt(this.value, 10));
            selectedDateKey = null;
            renderCalendar();
        });
    }

    var prevMonthBtn = document.getElementById("prevMonth");
    if (prevMonthBtn) {
        prevMonthBtn.addEventListener("click", function () {
            calendarDate.setMonth(calendarDate.getMonth() - 1);
            selectedDateKey = null;
            renderCalendar();
        });
    }

    var nextMonthBtn = document.getElementById("nextMonth");
    if (nextMonthBtn) {
        nextMonthBtn.addEventListener("click", function () {
            calendarDate.setMonth(calendarDate.getMonth() + 1);
            selectedDateKey = null;
            renderCalendar();
        });
    }

    var todayMonthBtn = document.getElementById("todayMonth");
    if (todayMonthBtn) {
        todayMonthBtn.addEventListener("click", function () {
            calendarDate = new Date();
            calendarDate.setDate(1);
            selectedDateKey = null;
            renderCalendar();
        });
    }

    var closeDayDetailBtn = document.getElementById("closeDayDetailBtn");
    if (closeDayDetailBtn) {
        closeDayDetailBtn.addEventListener("click", function () {
            selectedDateKey = null;
            renderCalendar();
        });
    }

    function showPage(page) {
        if (!pageViews[page]) {
            page = "dashboard";
        }
        for (var key in pageViews) {
            if (pageViews[key]) {
                pageViews[key].classList.toggle("active", key === page);
            }
        }
        document.querySelectorAll("[data-page]").forEach(function (el) {
            var target = el.getAttribute("data-page");
            if (el.classList.contains("nav-item") || el.classList.contains("nav-link")) {
                el.classList.toggle("active", target === page);
            }
        });
        if (breadcrumbTitle && titles[page]) {
            breadcrumbTitle.textContent = titles[page];
        }
        if (page === "calendar") {
            renderCalendar();
        }
    }

    document.addEventListener("click", function (e) {
        var nav = e.target.closest("[data-page]");
        if (nav && pageViews[nav.getAttribute("data-page")]) {
            var page = nav.getAttribute("data-page");
            if (document.getElementById("page-dashboard")) {
                e.preventDefault();
                showPage(page);
                if (history.pushState) {
                    var url = page === "dashboard" ? "/" : "/?page=" + page;
                    history.pushState(null, "", url);
                }
            }
        }
    });

    try {
        var urlParams = new URLSearchParams(window.location.search);
        var pageParam = urlParams.get("page");
        if (pageParam && pageViews[pageParam]) {
            showPage(pageParam);
        } else if (window.location.hash) {
            var h = window.location.hash.replace("#", "");
            if (pageViews[h]) {
                showPage(h);
            }
        }
    } catch (err) {}

    // Inisialisasi kalender awal jika halaman memuat tampilan kalender
    if (pageViews.calendar) {
        renderCalendar();
    }

    /* ---------------- Invite Modal ---------------- */
    var inviteModal = document.getElementById("inviteModal");
    var closeInviteModalBtn = document.getElementById("closeInviteModalBtn");
    var cancelInviteModalBtn = document.getElementById("cancelInviteModalBtn");

    function openInviteModal() {
        if (inviteModal) {
            inviteModal.hidden = false;
            var input = inviteModal.querySelector('input[name="invitee"]');
            if (input) {
                setTimeout(function () { input.focus(); }, 60);
            }
        }
    }

    function closeInviteModal() {
        if (inviteModal) {
            inviteModal.hidden = true;
        }
    }

    document.querySelectorAll(".open-invite-modal").forEach(function (btn) {
        btn.addEventListener("click", function (e) {
            e.preventDefault();
            openInviteModal();
        });
    });

    if (closeInviteModalBtn) {
        closeInviteModalBtn.addEventListener("click", closeInviteModal);
    }
    if (cancelInviteModalBtn) {
        cancelInviteModalBtn.addEventListener("click", closeInviteModal);
    }
    if (inviteModal) {
        inviteModal.addEventListener("click", function (e) {
            if (e.target === inviteModal) {
                closeInviteModal();
            }
        });
    }

    /* ---------------- Document Modal ---------------- */
    var docModal = document.getElementById("docModal");
    var closeDocModalBtn = document.getElementById("closeDocModalBtn");
    var cancelDocModalBtn = document.getElementById("cancelDocModalBtn");

    function openDocModal() {
        if (docModal) {
            docModal.hidden = false;
            var input = docModal.querySelector('input[name="document"]');
            if (input) {
                setTimeout(function () { input.focus(); }, 60);
            }
        }
    }

    function closeDocModal() {
        if (docModal) {
            docModal.hidden = true;
        }
    }

    document.querySelectorAll(".open-doc-modal").forEach(function (btn) {
        btn.addEventListener("click", function (e) {
            e.preventDefault();
            openDocModal();
        });
    });

    if (closeDocModalBtn) {
        closeDocModalBtn.addEventListener("click", closeDocModal);
    }
    if (cancelDocModalBtn) {
        cancelDocModalBtn.addEventListener("click", closeDocModal);
    }
    if (docModal) {
        docModal.addEventListener("click", function (e) {
            if (e.target === docModal) {
                closeDocModal();
            }
        });
    }

    // Initialize application language
    setLanguage(currentLang);
})();
