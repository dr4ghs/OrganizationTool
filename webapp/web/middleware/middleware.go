package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"

	"github.com/dr4ghs/orgtool/webapp/web/pages"
)

type CustomContext struct {
	context.Context
	StartTime time.Time
}

type (
	CustomHandler    func(ctx *CustomContext, w http.ResponseWriter, r *http.Request)
	CustomMiddleware func(ctx *CustomContext, w http.ResponseWriter, r *http.Request) error
)

func Chain(
	w http.ResponseWriter,
	r *http.Request,
	component templ.Component,
	middleware ...CustomMiddleware,
) {
	ctx := &CustomContext{
		Context:   context.Background(),
		StartTime: time.Now(),
	}

	for _, mw := range middleware {
		if err := mw(ctx, w, r); err != nil {
			return
		}
	}

	component.Render(ctx, w)
	Log(ctx, w, r)
}

func Protected(
	w http.ResponseWriter,
	r *http.Request,
	component templ.Component,
	middleware ...CustomMiddleware,
) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		Chain(w, r, pages.Login())

		return
	}

	r.Header.Add("Authorization", cookie.Value)

	Chain(w, r, component, middleware...)
}

func NoCache(ctx *CustomContext, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	return nil
}

func Log(ctx *CustomContext, w http.ResponseWriter, r *http.Request) error {
	elapsedTime := time.Since(ctx.StartTime)
	formattedTime := time.Now().Format("2025-02-02 16:00:00.000")

	log.Printf("[%s] [%s] [%s] [%s]\n", formattedTime, r.Method, r.URL.Path, elapsedTime)

	return nil
}
