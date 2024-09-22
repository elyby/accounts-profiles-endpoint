package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type MySQL struct {
	db                                *sql.DB
	findUsernameByUuidStmt            *sql.Stmt
	findUuidAndUsernameByUsernameStmt *sql.Stmt
}

func New(protocol string, host string, port uint, dbName string, user string, password string) (*MySQL, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%d)/%s", user, password, protocol, host, port, dbName))
	if err != nil {
		return nil, err
	}

	findUsernameByUuidStmt, err := db.Prepare(`
		SELECT username
		  FROM accounts
		 WHERE uuid = ?
		   AND status = 10
		 LIMIT 1
	 `)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare find username by uuid query: %w", err)
	}

	findUuidAndUsernameByUsernameStmt, err := db.Prepare(`
		SELECT uuid, username
		  FROM accounts
		 WHERE username = ?
		   AND status = 10
		 LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare find uuid by username query: %w", err)
	}

	return &MySQL{db, findUsernameByUuidStmt, findUuidAndUsernameByUsernameStmt}, nil
}

func NewWithConfig(config *viper.Viper) (*MySQL, error) {
	config.SetDefault("db.mysql.user", "root")
	config.SetDefault("db.mysql.password", "")
	config.SetDefault("db.mysql.host", "localhost")
	config.SetDefault("db.mysql.port", 3306)
	config.SetDefault("db.mysql.protocol", "tcp")

	return New(
		config.GetString("db.mysql.protocol"),
		config.GetString("db.mysql.host"),
		config.GetUint("db.mysql.port"),
		config.GetString("db.mysql.database"),
		config.GetString("db.mysql.user"),
		config.GetString("db.mysql.password"),
	)
}

func (m *MySQL) FindUsernameByUuid(ctx context.Context, uuid string) (string, error) {
	var username string
	err := m.findUsernameByUuidStmt.QueryRowContext(ctx, uuid).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return username, nil
}

func (m *MySQL) FindUuidByUsername(ctx context.Context, username string) (string, string, error) {
	var uuid, casedUsername string
	err := m.findUuidAndUsernameByUsernameStmt.QueryRowContext(ctx, username).Scan(&uuid, &casedUsername)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	} else if err != nil {
		return "", "", err
	}

	return uuid, casedUsername, nil
}

func (m *MySQL) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}
