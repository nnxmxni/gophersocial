package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/nnxmxni/gophersocial/internals/auth"
	"github.com/nnxmxni/gophersocial/internals/db"
	"github.com/nnxmxni/gophersocial/internals/env"
	"github.com/nnxmxni/gophersocial/internals/ratelimiter"
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/internals/store/cache"
	"go.uber.org/zap"
	"time"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		dbConfig: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/gophersocial?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			OTPExpiration: time.Hour * 24 * 3, // 3 days
		},
		auth: authConfig{
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "fallback"),
				exp:    time.Hour * 24 * 3,
				host:   "gophersocial",
			},
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pwd:     env.GetString("REDIS_PASSWORD", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATELIMITER_ENABLED", true),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	database, err := db.New(
		cfg.dbConfig.addr,
		cfg.dbConfig.maxOpenConns,
		cfg.dbConfig.maxIdleConns,
		cfg.dbConfig.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer database.Close()

	logger.Info("Database connection pool established")

	var rdb *redis.Client

	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pwd, cfg.redisCfg.db)
		logger.Info("Redis connection established")
	}

	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	cacheStorage := cache.NewRedisStorage(rdb)
	storage := store.NewStorage(database)

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.host, cfg.auth.token.host)

	app := &application{
		config:        cfg,
		store:         storage,
		cacheStorage:  cacheStorage,
		logger:        logger,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
