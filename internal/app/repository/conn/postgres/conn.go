package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	"github.com/Staysy-V/catalog-service/internal/app/config/section"
	"github.com/Staysy-V/catalog-service/migration"
)

type (
	Client struct {
		_bunDB
		rawBunDB *bun.DB

		cfg section.RepositoryPostgres
	}

	_bunDB = bun.IDB
)

func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawBunDB
}

func NewClient(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	var u url.URL
	u.Scheme = "postgres"
	u.Host = cfg.Address
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	u.Path = cfg.Name

	args := make(url.Values)
	args.Set("sslmode", "disable")
	u.RawQuery = args.Encode()

	dsn := u.String()

	log.Printf("postgres: read_timeout=%s write_timeout=%s",
		cfg.ReadTimeout, cfg.WriteTimeout)

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	sqlDB.SetMaxOpenConns(10)

	db := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Client{
		_bunDB:   db,
		rawBunDB: db,
		cfg:      cfg,
	}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()
	if err := migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("discover migrations: %w", err)
	}

	migrator := migrate.NewMigrator(
		c.rawBunDB,
		migrations,
		migrate.WithTableName(c.cfg.MigrationTable),
	)

	if err := migrator.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("init migrations table: %w", err)
	}

	applied, err := migrator.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("get applied migrations: %w", err)
	}

	oldVer = applied.LastGroupID()

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return oldVer, 0, fmt.Errorf("apply migrations: %w", err)
	}

	if group.IsZero() {
		newVer = oldVer
	} else {
		newVer = group.ID
	}

	return oldVer, newVer, nil
}
