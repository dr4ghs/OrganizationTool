package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dr4ghs/orgtool/webapp/internal"
	"github.com/dr4ghs/orgtool/webapp/web/components/activities"
	"github.com/dr4ghs/orgtool/webapp/web/components/notifications"
	"github.com/dr4ghs/orgtool/webapp/web/middleware"
	"github.com/dr4ghs/orgtool/webapp/web/pages"
)

var mux *http.ServeMux

//go:embed all:frontend/dist
var static embed.FS

var fsHandler http.Handler

func init() {
	mux = http.NewServeMux()

	registerRouter(mux)
}

func registerRouter(mux *http.ServeMux) {
	staticFS, err := fs.Sub(static, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("GET /", landingPageHandler)
	mux.HandleFunc("GET /register", registerPageHandler)
	mux.HandleFunc("GET /login", loginPageHandler)

	// HOME
	mux.HandleFunc("GET /home", homePageHandler)
	mux.HandleFunc("GET /home/activities", listActivitiesHandler)
	mux.HandleFunc("POST /home/activities", newActivityHandler)

	// NOTIFICATIONS
	mux.HandleFunc("GET /notification", func(w http.ResponseWriter, r *http.Request) {
		typ, err := strconv.Atoi(r.URL.Query().Get("type"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if typ < 0 || typ > 2 {
			http.Error(w, "Wrong notification type", http.StatusBadRequest)
			return
		}

		message := r.Header.Get("X-Notification")
		if message == "" {
			http.Error(w, "Notification body cannot be empty", http.StatusBadRequest)
			return
		}

		middleware.Chain(w, r, notifications.Notification(internal.NotificationType(typ), message))
	})

	// mux.HandleFunc("POST /api/auth/signup", signUpPageHandler)
	mux.HandleFunc("GET /api/auth/login", checkLoginHandler)
	mux.HandleFunc("POST /api/auth/login", loginHandler)

	mux.HandleFunc("GET /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
		}

		http.SetCookie(w, cookie)
	})

	mux.HandleFunc("POST /api/activities", createUpdateActivityHandler)
	mux.HandleFunc("DELETE /api/activities/{id}", deleteActivityHandler)
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	middleware.Chain(w, r, pages.Landing())
}

func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		http.NotFound(w, r)
		return
	}

	middleware.Chain(w, r, pages.Register())
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
		return
	}

	middleware.Chain(w, r, pages.Login())
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}

	middleware.Protected(w, r, pages.Home())
}

// func signUpPageHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/api/auth/signup" {
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	url := fmt.Sprintf("%s/collections/users/records", os.Getenv("BACKEND_URL"))
// 	resp, err := http.Post(url, "application/json", io.NopCloser(bytes.NewReader(body)))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadGateway)
// 		return
// 	}
//
// 	if resp.StatusCode == http.StatusOK {
// 		middleware.Chain(w, r, pages.Register())
// 	}
// }

func checkLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/auth/login" {
		http.NotFound(w, r)
		return
	}

	middleware.Protected(w, r, pages.Home(), middleware.NoCache)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/auth/login" {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("%s/api/collections/users/auth-with-password", os.Getenv("BACKEND_URL"))
	resp, err := http.Post(url, "application/json", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		http.Error(w, err.Error(), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var data map[string]any
		json.NewDecoder(resp.Body).Decode(&data)

		cookie := http.Cookie{
			Name:     "access_token",
			Value:    data["token"].(string),
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}
		http.SetCookie(w, &cookie)

		record := data["record"].(map[string]any)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"id":     record["id"],
			"email":  record["email"],
			"name":   record["name"],
			"points": record["points"],
		})
	} else {
		http.Error(w, string(body), resp.StatusCode)
	}
}

type ListActivitiesResponse struct {
	Page       int                 `json:"page"`
	PerPage    int                 `json:"perPage"`
	TotalPages int                 `json:"totalPages"`
	TotalItems int                 `json:"totalItems"`
	Items      []internal.Activity `json:"items"`
}

func listActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home/activities" {
		http.NotFound(w, r)
		return
	}

	cookie, _ := r.Cookie("access_token")
	if cookie == nil {
		middleware.Chain(w, r, pages.Login())

		return
	}

	url := fmt.Sprintf("%s/api/collections/activities/records", os.Getenv("BACKEND_URL"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", cookie.Value)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var list ListActivitiesResponse
	if err := json.Unmarshal(body, &list); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	middleware.Protected(w, r, activities.ActivityList(list.Items), middleware.NoCache)
}

func newActivityHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home/activities" {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	a := internal.Activity{}
	json.Unmarshal(body, &a)

	middleware.Chain(w, r, activities.Activity(a, true), middleware.NoCache)
}

func createUpdateActivityHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/activities" {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a := internal.Activity{}
	if err := json.Unmarshal(body, &a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("access_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if err := a.Validate(); err != nil {
		internal.AddNotification(w, internal.WarningNotification, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.Save(cookie.Value); err != nil {
		internal.AddNotification(w, internal.ErrorNotification, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	internal.AddNotification(w, internal.InfoNotification, "Activity saved")
	middleware.Chain(w, r, activities.Activity(a, false))
}

// func updateActivityHandler(w http.ResponseWriter, r *http.Request) {
// 	id := r.PathValue("id")
//
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	a := internal.Activity{}
// 	if err := json.Unmarshal(body, &a); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	cookie, err := r.Cookie("access_token")
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, http.ErrNoCookie):
// 			w.WriteHeader(http.StatusUnauthorized)
// 		default:
// 			w.WriteHeader(http.StatusInternalServerError)
// 		}
//
// 		return
// 	}
//
// 	r.Header.Add("Authorization", cookie.Value)
// 	url := fmt.Sprintf("%s/api/collections/activities/records/%s", os.Getenv("BACKEND_URL"), id)
// 	req, _ := http.NewRequest(http.MethodPatch, url, nil)
// 	req.Header.Add("Authorization", cookie.Value)
// 	client := &http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		http.Error(w, err.Error(), res.StatusCode)
// 	}
// 	defer res.Body.Close()
// }

func deleteActivityHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cookie, err := r.Cookie("access_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	r.Header.Add("Authorization", cookie.Value)
	url := fmt.Sprintf("%s/api/collections/activities/records/%s", os.Getenv("BACKEND_URL"), id)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Add("Authorization", cookie.Value)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		internal.AddNotification(w, internal.ErrorNotification, err.Error())
		http.Error(w, err.Error(), res.StatusCode)
		return
	}
	defer res.Body.Close()

	internal.AddNotification(w, internal.InfoNotification, "Activity deleted")
	w.WriteHeader(http.StatusOK)
}
