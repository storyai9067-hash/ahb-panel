package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       string    `json:"id"`
	UUID     string    `json:"uuid"`
	Traffic  int64     `json:"traffic"`
	Expiry   time.Time `json:"expiry"`
	IsActive bool      `json:"isActive"`
	Email    string    `json:"email"`
}

type Admin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Settings struct {
	Domain string `json:"domain"`
}

var users []User
var settings = Settings{Domain: "railway.app"}
var admin = Admin{Username: "admin", Password: "admin"}
var mu sync.Mutex

func main() {
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			username := r.FormValue("username")
			password := r.FormValue("password")
			if username == admin.Username && password == admin.Password {
				http.SetCookie(w, &http.Cookie{
					Name:  "session",
					Value: "authenticated",
					Path:  "/",
				})
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
			return
		}
		tmpl := `<!DOCTYPE html>
		<html>
		<head><title>ورود به ahb panel</title>
		<style>
			body { font-family: Arial; background: #0f172a; color: #fff; display: flex; justify-content: center; align-items: center; height: 100vh; }
			.login-box { background: #1e293b; padding: 40px; border-radius: 15px; max-width: 400px; width: 100%; }
			h1 { color: #38bdf8; text-align: center; }
			input { width: 100%; padding: 10px; margin: 10px 0; border-radius: 8px; border: 1px solid #334155; background: #0f172a; color: #fff; }
			button { width: 100%; padding: 12px; background: #38bdf8; color: #0f172a; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; }
			.error { color: #ff6b6b; text-align: center; }
		</style>
		</head>
		<body>
		<div class="login-box">
			<h1>🔐 ahb panel</h1>
			{{if .}}<div class="error">❌ نام کاربری یا رمز اشتباه است!</div>{{end}}
			<form method="POST">
				<input name="username" placeholder="نام کاربری" required>
				<input name="password" type="password" placeholder="رمز عبور" required>
				<button type="submit">ورود</button>
			</form>
		</div>
		</body>
		</html>`
		t := template.Must(template.New("login").Parse(tmpl))
		errorParam := r.URL.Query().Get("error")
		t.Execute(w, errorParam == "1")
	})

	func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value != "authenticated" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next(w, r)
		}
	}

	http.HandleFunc("/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<!DOCTYPE html>
		<html>
		<head><title>ahb panel</title>
		<style>
			body { font-family: Arial; background: #0f172a; color: #fff; padding: 20px; }
			.container { max-width: 800px; margin: 0 auto; background: #1e293b; padding: 20px; border-radius: 15px; }
			h1 { color: #38bdf8; }
			form { background: #0f172a; padding: 20px; border-radius: 10px; margin: 20px 0; }
			input { width: 100%; padding: 10px; margin: 8px 0; border-radius: 8px; border: 1px solid #334155; background: #1e293b; color: #fff; }
			button { background: #38bdf8; color: #0f172a; padding: 12px 20px; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; }
			ul { list-style: none; padding: 0; }
			li { background: #0f172a; margin: 5px 0; padding: 10px; border-radius: 8px; border-left: 4px solid #38bdf8; }
			.nav { display: flex; gap: 20px; margin-bottom: 20px; flex-wrap: wrap; }
			.nav a { color: #38bdf8; text-decoration: none; padding: 10px; background: #0f172a; border-radius: 8px; }
			.logout { color: #ff6b6b; }
		</style>
		</head>
		<body>
		<div class="container">
			<h1>🚀 ahb panel</h1>
			<div class="nav">
				<a href="/">📋 کاربران</a>
				<a href="/settings">⚙️ تنظیمات</a>
				<a href="/admin">👤 مدیریت حساب</a>
				<a href="/logout" class="logout">🚪 خروج</a>
			</div>
			<h2>➕ ساخت کاربر جدید</h2>
			<form action="/create-user" method="POST">
				<input name="email" placeholder="ایمیل کاربر" required>
				<input name="traffic" placeholder="حجم (گیگابایت)" type="number" required>
				<input name="expiry" placeholder="انقضا (مثلاً 2025-12-31)" required>
				<button type="submit">ساخت کاربر</button>
			</form>
			<h2>📋 لیست کاربران</h2>
			<ul>
				{{range .}}
				<li>
					<strong>UUID:</strong> {{.UUID}} |
					<strong>حجم:</strong> {{.Traffic}} بایت |
					<strong>انقضا:</strong> {{.Expiry}} |
					<strong>وضعیت:</strong> {{if .IsActive}}✅ فعال{{else}}❌ غیرفعال{{end}} |
					<a href="/sub/{{.UUID}}" style="color:#38bdf8;">🔗 لینک ساب</a>
				</li>
				{{else}}
				<li>هیچ کانفیگی ساخته نشده است.</li>
				{{end}}
			</ul>
		</div>
		</body>
		</html>`
		t := template.Must(template.New("index").Parse(tmpl))
		mu.Lock()
		t.Execute(w, users)
		mu.Unlock()
	}))

	http.HandleFunc("/admin", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			newUsername := r.FormValue("username")
			newPassword := r.FormValue("password")
			if newUsername != "" && newPassword != "" {
				mu.Lock()
				admin.Username = newUsername
				admin.Password = newPassword
				mu.Unlock()
			}
			http.Redirect(w, r, "/admin?success=1", http.StatusSeeOther)
			return
		}
		tmpl := `<!DOCTYPE html>
		<html>
		<head><title>مدیریت حساب</title>
		<style>
			body { font-family: Arial; background: #0f172a; color: #fff; padding: 20px; }
			.container { max-width: 600px; margin: 0 auto; background: #1e293b; padding: 20px; border-radius: 15px; }
			h1 { color: #38bdf8; }
			form { background: #0f172a; padding: 20px; border-radius: 10px; margin: 20px 0; }
			input { width: 100%; padding: 10px; margin: 8px 0; border-radius: 8px; border: 1px solid #334155; background: #1e293b; color: #fff; }
			button { background: #38bdf8; color: #0f172a; padding: 12px 20px; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; }
			.nav a { color: #38bdf8; text-decoration: none; padding: 10px; background: #0f172a; border-radius: 8px; margin-right: 10px; }
			.success { color: #51cf66; }
		</style>
		</head>
		<body>
		<div class="container">
			<h1>👤 مدیریت حساب</h1>
			<div class="nav"><a href="/">📋 کاربران</a><a href="/settings">⚙️ تنظیمات</a><a href="/logout">🚪 خروج</a></div>
			{{if .}}<div class="success">✅ اطلاعات با موفقیت تغییر کرد!</div>{{end}}
			<form method="POST">
				<input name="username" placeholder="نام کاربری جدید" value="{{.Username}}">
				<input name="password" type="password" placeholder="رمز عبور جدید">
				<button type="submit">به‌روزرسانی</button>
			</form>
			<div style="margin-top:20px; color:#94a3b8;">نام کاربری فعلی: <strong>{{.Username}}</strong></div>
		</div>
		</body>
		</html>`
		t := template.Must(template.New("admin").Parse(tmpl))
		mu.Lock()
		successParam := r.URL.Query().Get("success")
		data := struct {
			Username string
			Success  bool
		}{admin.Username, successParam == "1"}
		t.Execute(w, data)
		mu.Unlock()
	}))

	http.HandleFunc("/settings", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<!DOCTYPE html>
		<html>
		<head><title>تنظیمات</title>
		<style>
			body { font-family: Arial; background: #0f172a; color: #fff; padding: 20px; }
			.container { max-width: 800px; margin: 0 auto; background: #1e293b; padding: 20px; border-radius: 15px; }
			h1 { color: #38bdf8; }
			form { background: #0f172a; padding: 20px; border-radius: 10px; margin: 20px 0; }
			input { width: 100%; padding: 10px; margin: 8px 0; border-radius: 8px; border: 1px solid #334155; background: #1e293b; color: #fff; }
			button { background: #38bdf8; color: #0f172a; padding: 12px 20px; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; }
			.nav a { color: #38bdf8; text-decoration: none; padding: 10px; background: #0f172a; border-radius: 8px; margin-right: 10px; }
			.info { background: #0f172a; padding: 15px; border-radius: 8px; margin: 10px 0; }
		</style>
		</head>
		<body>
		<div class="container">
			<h1>⚙️ تنظیمات</h1>
			<div class="nav"><a href="/">📋 کاربران</a><a href="/admin">👤 مدیریت حساب</a><a href="/logout">🚪 خروج</a></div>
			<h2>🔗 تغییر دامنه لینک ساب</h2>
			<form action="/update-domain" method="POST">
				<input name="domain" placeholder="دامنه جدید (مثلاً ahb.railway.app)" value="{{.Domain}}" required>
				<button type="submit">به‌روزرسانی دامنه</button>
			</form>
			<div class="info"><strong>دامنه فعلی:</strong> {{.Domain}}</div>
		</div>
		</body>
		</html>`
		t := template.Must(template.New("settings").Parse(tmpl))
		mu.Lock()
		t.Execute(w, settings)
		mu.Unlock()
	}))

	http.HandleFunc("/update-domain", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		newDomain := r.FormValue("domain")
		if newDomain != "" {
			mu.Lock()
			settings.Domain = newDomain
			mu.Unlock()
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	}))

	http.HandleFunc("/create-user", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		email := r.FormValue("email")
		trafficGB := r.FormValue("traffic")
		expiryStr := r.FormValue("expiry")

		traffic, _ := strconv.ParseInt(trafficGB, 10, 64)
		trafficBytes := traffic * 1024 * 1024 * 1024
		expiry, _ := time.Parse("2006-01-02", expiryStr)

		mu.Lock()
		defer mu.Unlock()
		newUser := User{
			ID:       fmt.Sprintf("user-%d", len(users)+1),
			UUID:     uuid.New().String(),
			Traffic:  trafficBytes,
			Expiry:   expiry,
			IsActive: true,
			Email:    email,
		}
		users = append(users, newUser)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	http.HandleFunc("/sub/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		uuid := r.URL.Path[len("/sub/"):]
		mu.Lock()
		defer mu.Unlock()
		for _, user := range users {
			if user.UUID == uuid && user.IsActive {
				domain := settings.Domain
				config := fmt.Sprintf("vless://%s@%s:443?encryption=none&security=tls&sni=%s&fp=chrome&type=tcp#ahb-panel", user.UUID, domain, domain)
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprint(w, config)
				return
			}
		}
		http.Error(w, "User not found or inactive", http.StatusNotFound)
	}))

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	fmt.Println("🚀 ahb panel is running on :8080")
	http.ListenAndServe(":8080", nil)
}
