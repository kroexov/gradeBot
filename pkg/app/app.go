package app

import (
	"context"
	"time"

	botService "gradebot/pkg/bot"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"

	"github.com/go-pg/pg/v10"
	"github.com/go-telegram/bot"
)

type Config struct {
	Database *pg.Options
	Bot      struct {
		Token string
	}
}

type App struct {
	embedlog.Logger
	appName string
	cfg     Config
	db      db.DB
	b       *bot.Bot
	dbc     *pg.DB

	bs *botService.BotService
}

func New(appName string, verbose bool, cfg Config, db db.DB, dbc *pg.DB) *App {
	a := &App{
		appName: appName,
		cfg:     cfg,
		db:      db,
		dbc:     dbc,
	}

	a.SetStdLoggers(verbose)

	a.bs = botService.NewBotService(a.Logger, a.db)

	opts := []bot.Option{bot.WithDefaultHandler(a.bs.DefaultHandler)}
	b, err := bot.New(cfg.Bot.Token, opts...)
	if err != nil {
		panic(err)
	}
	a.b = b

	return a
}

// Run is a function that runs application.
func (a *App) Run() error {
	//registerBotHandlers()
	a.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "song", bot.MatchTypeExact, botService.FindUndefinedSong)
	go a.b.Start(context.TODO())
	return nil
}

// Shutdown is a function that gracefully stops HTTP server.
func (a *App) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if _, err := a.b.Close(ctx); err != nil {
		a.Errorf("shutting down bot err=%q", err)
	}
}
