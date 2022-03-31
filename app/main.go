package main

import (
	"context"
	"fmt"
	"os"

	"github.com/soerenstahlmann/go-auth/ent"
	"github.com/soerenstahlmann/go-auth/server"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		return fmt.Errorf("failed opening connection to database: %w", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}

	// create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("could not create zap logger: %w", err)
	}

	s, err := server.NewServer(
		server.WithClient(client),
		server.WithJWTSecret([]byte("<YOUR SECRET HERE>")), // TODO: Load from env var or CLI arg
		server.WithLogger(logger.Sugar()),
	)
	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}

	return s.Run()

}
