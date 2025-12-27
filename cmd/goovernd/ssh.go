package main

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ionut-maxim/goovern/app"
	"github.com/ionut-maxim/goovern/db"
)

func makeTeaHandler(pool *pgxpool.Pool, dbClient *db.DB) func(ssh.Session) (tea.Model, []tea.ProgramOption) {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		return app.NewModel(s, pool, dbClient)
	}
}

func startSSHServer(pool *pgxpool.Pool, db *db.DB, port int, logger *slog.Logger, done chan<- os.Signal) *ssh.Server {
	rateLimiter := ratelimiter.NewRateLimiter(2.0, 5, 200)

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("", strconv.Itoa(port))),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(makeTeaHandler(pool, db)),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			ratelimiter.Middleware(rateLimiter),
			logging.Middleware(),
		),
	)
	if err != nil {
		logger.Error("Could not start server", "error", err)
	}

	logger.Info("Starting SSH server", "host", "", "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	return s
}
