package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/VJ-2303/jobhub/internal/data"
	"github.com/VJ-2303/jobhub/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type application struct {
	logger  *slog.Logger
	config  config
	models  data.Models
	limiter *Limiter
	mailer  *mailer.EmailService
	wg      sync.WaitGroup
}

type config struct {
	port        int
	environment string
	db          struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	smtp struct {
		port     int
		username string
		host     string
		password string
	}
}

func main() {
	godotenv.Load()
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cfg.environment, "environment", "DEV", "Environment (PROD | DEV | STAG )")

	flag.StringVar(&cfg.db.dsn, "dsn", "postgres://username:password@hostname:port/database_name", "Postgres connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql connections max idle timeout")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.gmail.com", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-user", "vanaraj1018@gmail.com", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-pass", os.Getenv("SMTPPASS"), "SMTP password")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	mailer, err := mailer.NewEmailService(cfg.smtp.port, cfg.smtp.host, cfg.smtp.username, cfg.smtp.password)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("Database connection established")

	app := &application{
		logger:  logger,
		config:  cfg,
		models:  data.NewModels(db),
		limiter: NewLimiter(1, 3),
		mailer:  mailer,
	}
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
	}
	logger.Info("Starting server", "ENVIRONMENT", cfg.environment, "PORT", cfg.port)

	err = srv.ListenAndServe()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
