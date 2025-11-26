package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"tailscale.com/tsnet"
)

func InitDb(cfg *Config) (*sql.DB, error) {
	log.Println("[INIT] initializing database connection")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		cfg.Database.User,
		cfg.Database.Pass,
		cfg.Database.Host,
		cfg.Database.Name,
	)

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.TailscaleAuthKey != "" {
		log.Printf("[INIT] need tsnet for host: %s\n", cfg.Database.Host)

		// We use os.TempDir because Cloud Run filesystems are ephemeral (in-memory)
		dir := filepath.Join(os.TempDir(), "tsnet-state")
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create tsnet dir: %w", err)
		}

		s := &tsnet.Server{
			Hostname: "go-proxy",
			AuthKey:  cfg.TailscaleAuthKey,
			Dir:      dir,
			// Reduce log noise
			Logf: func(format string, args ...any) {},
		}

		// IMPORTANT: We do not defer s.Close() here.
		// The server needs to live as long as the application lives.

		// Override the Dialer to use the VPN
		pgxConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// This hijacks the TCP connection process
			return s.Dial(ctx, network, addr)
		}
	} else {
		log.Println("[INIT] No Tailscale key found")
	}

	// allows us to use database/sql with our custom dialer
	connStr := stdlib.RegisterConnConfig(pgxConfig)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql connection: %w", err)
	}

	// VPN handshakes can take a while
	log.Printf("[INIT] pinging database with 10s T/O...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("[INIT] Database connected successfully")

	query := `
        SELECT EXISTS (
    		SELECT 1
    		FROM pg_tables
    		WHERE schemaname = 'default' AND tablename = 'api_metrics'
		);
    `
	if _, err := db.Exec(query); err != nil {
		log.Printf("[INIT] could not run basic table check")
	}

	return db, nil

}
