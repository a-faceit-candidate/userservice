package suite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/a-faceit-candidate/restuser"
	"github.com/cabify/aceptadora"
	"github.com/colega/envconfig"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/suite"
)

const expectedMockedDependencyInventedHTTPStatusCode = 288

type Config struct {
	Aceptadora aceptadora.Config

	// ServicesAddress is the address where services started by suiteAceptadora can be found
	// It differs from env to env, and it's set up in env-specific configs
	ServicesAddress string

	RESTClient restuser.Config
	MysqlDSN   string
}

type acceptanceSuite struct {
	suite.Suite

	cfg Config

	suiteAceptadora *aceptadora.Aceptadora
	testAceptadora  *aceptadora.Aceptadora

	client *restuser.API
	db     *sql.DB

	userCreatedConsumer, userUpdatedConsumer, userDeletedConsumer *nsq.Consumer
	userCreatedMessages, userUpdatedMessages, userDeletedMessages chan string
}

func (s *acceptanceSuite) SetupSuite() {
	aceptadora.SetEnv(
		s.T(),
		aceptadora.EnvConfigAlways("../config/default.env"),
		aceptadora.EnvConfigAlways("acceptance.env"),
	)
	s.Require().NoError(envconfig.Process("ACCEPTANCE", &s.cfg))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s.suiteAceptadora = aceptadora.New(s.T(), s.cfg.Aceptadora)
	s.suiteAceptadora.PullImages(ctx)

	// we only start mysql once per suite as it takes too long to start
	// then we just truncate the tables after each run
	s.suiteAceptadora.Run(ctx, "mysql")
	s.Require().Eventually(mysqlHealthCheck(ctx, s.cfg.MysqlDSN), time.Minute, 100*time.Millisecond, "mysql didn't start")

	db, err := sql.Open("mysql", s.cfg.MysqlDSN)
	s.Require().NoError(err)

	s.db = db

	s.client = restuser.New(s.cfg.RESTClient)
}

func (s *acceptanceSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s.testAceptadora = aceptadora.New(s.T(), s.cfg.Aceptadora)
	s.testAceptadora.Run(ctx, "nsqd")
	s.Require().Eventually(
		httpHealthCheck(s.cfg.ServicesAddress, 4151, "/ping"),
		time.Minute, 50*time.Millisecond,
		"nsqd didn't start",
	)
	s.userCreatedConsumer, s.userCreatedMessages = s.nsqConsumerToChannel("user.created")
	s.userUpdatedConsumer, s.userUpdatedMessages = s.nsqConsumerToChannel("user.updated")
	s.userDeletedConsumer, s.userDeletedMessages = s.nsqConsumerToChannel("user.deleted")

	s.testAceptadora.Run(ctx, "userservice")
	s.Require().Eventually(
		httpHealthCheck(s.cfg.ServicesAddress, 8080, "/status"),
		time.Minute, 50*time.Millisecond,
		"userservice didn't start",
	)
}

func (s *acceptanceSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.userCreatedConsumer.Stop()
	s.userUpdatedConsumer.Stop()
	s.userDeletedConsumer.Stop()

	s.testAceptadora.StopAll(ctx)
	_, err := s.db.ExecContext(ctx, "TRUNCATE user")
	s.Require().NoError(err)
}

func (s *acceptanceSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.suiteAceptadora.StopAll(ctx)
}

func TestAcceptanceSuite(t *testing.T) {
	suite.Run(t, new(acceptanceSuite))
}

// nsqConsumerToChannel creates a nsq.Consumer that handles JSON string messages from the given topic
// and sends them through the provided channel
func (s acceptanceSuite) nsqConsumerToChannel(topic string) (*nsq.Consumer, chan string) {
	ch := make(chan string, 1000) // a channel with enough buffer to keep our messages
	consumer, err := nsq.NewConsumer(topic, "acceptance-test", nsq.NewConfig())
	s.Require().NoError(err)

	consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {
		var str string
		if err := json.Unmarshal(msg.Body, &str); err != nil {
			s.T().Logf("Failed to unmarshal msg from topic %s as string: %s", topic, err)
			return err
		}
		ch <- str
		return nil
	}))

	err = consumer.ConnectToNSQD(fmt.Sprintf("%s:4150", s.cfg.ServicesAddress))
	s.Require().NoError(err)

	return consumer, ch
}

// httpHealthCheck will return true if the provided endpoint on a given host/port returns a 200 status code to a GET request
func httpHealthCheck(host string, port int, endpoint string) func() bool {
	return func() bool {
		url := fmt.Sprintf("http://%s:%d%s", host, port, endpoint)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			panic(fmt.Errorf("can't build http request for healthcheck, maybe wrong config? %s", err))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
}

// mysqlHealthCheck will try to instantiate an sql.DB and ping it (as instantiating doesn't create any connection)
func mysqlHealthCheck(ctx context.Context, dsn string) func() bool {
	return func() bool {
		// if ping fails, mysql logger will complain
		// we don't want to listen to that, so we silence it during the healthcheck
		mysql.SetLogger(noopLogger{})
		defer mysql.SetLogger(mysql.Logger(log.New(os.Stderr, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile)))

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return false
		}
		defer db.Close()
		return db.PingContext(ctx) == nil
	}
}

type noopLogger struct{}

func (noopLogger) Print(...interface{}) {}
