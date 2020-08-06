package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/ent"
	"github.com/huangc28/go-darkpanda-backend/ent/migrate"
	_ "github.com/lib/pq"
)

func init() {
	config.InitConfig()
}

//  - run the latest migrations
//  - record SQL that is going to be executed for historical traceback in the future
func main() {

	ac := config.GetAppConf()

	log.Printf("ac %v", ac)

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ac.DBConf.Host,
		ac.DBConf.Port,
		ac.DBConf.User,
		ac.DBConf.Password,
		ac.DBConf.Dbname,
	)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// run the auto migration tool.
	// we need to record the SQL that is going to be executed for historical traceback in the future.
	ctx := context.Background()

	if err = client.Schema.WriteTo(ctx, os.Stdout); err != nil {
		log.Fatalf("failed printing schema changes", err.Error())
	}

	err = client.Debug().Schema.Create(
		ctx,
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	)

	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
