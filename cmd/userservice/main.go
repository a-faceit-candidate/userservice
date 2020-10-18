package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/a-faceit-candidate/userservice/internal/api"
	"github.com/a-faceit-candidate/userservice/internal/event"
	"github.com/a-faceit-candidate/userservice/internal/log"
	"github.com/a-faceit-candidate/userservice/internal/persistence"
	"github.com/a-faceit-candidate/userservice/internal/service"
	"github.com/colega/envconfig"
	"github.com/gin-gonic/gin"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

type config struct {
	// Host configures the interface where the app listens, can be empty
	Host string
	// Port configures the port where the app listens
	Port string `default:"8080"`
	// MysqlDSN is formed as "user:password@network(address)/database?options
	// We could split it into separate env vars and build the DSN in the service, but this is enough for the challenge
	MysqlDSN string

	// NsqdAddr is the TCP address of the nsqd to use
	NsqdAddr string `default:"nsqd:4150"`
}

func main() {
	var cfg config
	envconfig.MustProcess("APP", &cfg)

	producer, err := nsq.NewProducer(cfg.NsqdAddr, nsq.NewConfig())
	successOrPanicf("Can't instantiate NSQ producer: %s", err)

	db, err := sql.Open("mysql", cfg.MysqlDSN)
	successOrPanicf("Can't dial MySQL conn: %s", err)

	userRepo := persistence.NewObservedRepository(
		persistence.NewMysqlRepository(db),
		event.NewNSQPublisher(producer),
	)

	svc := service.New(userRepo)
	userResource := api.NewUsersResource(svc)

	g := gin.New()
	g.Use(log.AddLogContextBaggage)
	g.GET("/status", func(c *gin.Context) { c.Status(http.StatusOK) })
	userResource.AddRoutes(g.Group("/v1"))

	// TODO: prevent high cardinality metrics by removing :id params from labels
	ginprometheus.NewPrometheus("gin").Use(g)

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", addr)
	successOrPanicf("Can't listen: %s", err)
	logrus.Infof("Listening on %s", addr)

	srv := &http.Server{Handler: g}
	go func() {
		if err := srv.Serve(ln); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	defer srv.Close()

	waitForSignal()
}

func waitForSignal() {
	logrus.Infof("Listening for shutdown signal")
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	sig := <-ch
	logrus.Infof("Received signal %s", sig)
}

func successOrPanicf(msg string, err error) {
	if err != nil {
		panic(fmt.Errorf(msg, err))
	}
}
