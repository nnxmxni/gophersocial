package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nnxmxni/gophersocial/internals/auth"
	"github.com/nnxmxni/gophersocial/internals/ratelimiter"
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/internals/store/cache"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr          string
	OTPExpiration time.Duration
	dbConfig      dbConfig
	mail          mailConfig
	auth          authConfig
	redisCfg      redisConfig
	rateLimiter   ratelimiter.Config
}

type redisConfig struct {
	addr    string
	pwd     string
	db      int
	enabled bool
}

type authConfig struct {
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	host   string
}

type mailConfig struct {
	OTPExpiration time.Duration
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.RateLimiterMiddleware)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/post", func(r chi.Router) {

			r.Use(app.EnsureAuthMiddleware)

			r.Post("/create", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {

				r.Use(app.postsContextMiddleware)

				r.Get("/show", app.getPostHandler)
				r.Patch("/update", app.EnsurePostOwnership("moderator", app.updatePostHandler))
				r.Delete("/delete", app.EnsurePostOwnership("admin", app.deletePostHandler))
			})
		})

		r.Route("/users", func(r chi.Router) {

			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {

				r.Use(app.EnsureAuthMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.EnsureAuthMiddleware)

				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Group(func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
			r.Post("/login", app.loginUserHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Infof("Received signal %s, shutting down the server", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server is listening on", "addr", app.config.addr)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	app.logger.Infow("server has stopped gracefully")
	err = <-shutdown
	if err != nil {
		return err
	}

	return nil
}
