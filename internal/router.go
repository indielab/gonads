package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"northstar/config"
	counterFeature "northstar/internal/features/counter"
	indexFeature "northstar/internal/features/index"
	monitorFeature "northstar/internal/features/monitor"
	reverseFeature "northstar/internal/features/reverse"
	sortableFeature "northstar/internal/features/sortable"
	"northstar/web/resources"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/starfederation/datastar-go/datastar"
)

func SetupRoutes(ctx context.Context, router chi.Router, sessionStore *sessions.CookieStore, ns *embeddednats.Server) (err error) {

	if config.Global.Environment == config.Dev {
		setupReload(ctx, router)
	}

	router.Handle("/static/*", resources.Handler())

	if err := errors.Join(
		indexFeature.SetupRoutes(router, sessionStore, ns),
		counterFeature.SetupRoutes(router, sessionStore),
		monitorFeature.SetupRoutes(router),
		sortableFeature.SetupRoutes(router),
		reverseFeature.SetupRoutes(router),
	); err != nil {
		return fmt.Errorf("error setting up routes: %w", err)
	}

	return nil
}

func setupReload(ctx context.Context, router chi.Router) {
	reloadChan := make(chan struct{}, 1)
	var hotReloadOnce sync.Once
	router.Get("/reload", func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		reload := func() { sse.ExecuteScript("window.location.reload()") }
		hotReloadOnce.Do(reload)
		select {
		case <-reloadChan:
			reload()
		case <-r.Context().Done():
		}
	})

	router.Get("/hotreload", func(w http.ResponseWriter, r *http.Request) {
		select {
		case reloadChan <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

}
