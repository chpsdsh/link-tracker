package integration

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	scrapperDockerfile    = "scrapper.Dockerfile"
	botDockerfile         = "bot.Dockerfile"
	scrapperPort          = "8081/tcp"
	botPort               = "8080/tcp"
	scrapperAlias         = "scrapper"
	botAlias              = "bot"
	telegramAPIKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
	githubAPIKey          = "GITHUB_API_KEY"
	stackoverflowAPIKey   = "STACKOVERFLOW_API_KEY"
	botServerAddress      = "BOT_SERVER_ADDRESS"
	botAPIFlag            = "WITH_TELEGRAM_API"
	APIToken              = "API_TOKEN"
	botServerAddr         = "http://bot:8080"
	scrapperServerAddr    = "http://scrapper:8081"
	withTelegramAPI       = "false"
	pathToDockerfile      = "../../"
	assetType             = "ASSET_TYPE"
	assetTypeBuilder      = "BUILDER"
	dbUser                = "user"
	dbPassword            = "password"
	dbName                = "mydb"
	exposedPort           = "5432/tcp"
	valkeyPort            = "6379/tcp"
	dbPort                = "5432"
	dbHost                = "postgres"
	postgresHost          = "POSTGRES_HOST"
	postgresPort          = "POSTGRES_PORT"
	postgresUser          = "POSTGRES_USER"
	postgresPassword      = "POSTGRES_PASSWORD"
	postgresDatabase      = "POSTGRES_DB"
	updatesSendTypeEnv    = "UPDATES_SEND_TYPE"
	updatesReceiveTypeEnv = "UPDATES_RECEIVER_TYPE"
	kafkaConsumerGroupEnv = "KAFKA_CONSUMER_GROUP"
	kafkaUserEnv          = "KAFKA_USER"
	kafkaPasswordEnv      = "KAFKA_PASSWORD"
	kafkaTopicEnv         = "KAFKA_TOPIC"
	kafkaBrokersEnv       = "KAFKA_BROKERS"
	kafkaDLQTopicEnv      = "KAFKA_DLQ_TOPIC"
	valkeyPasswordEnv     = "VALKEY_PASSWORD"
	valkeyAddressesEnv    = "VALKEY_ADDRESSES"
	valkeyTTLMinutesEnv   = "VALKEY_TTL_MINUTES"
	updatesSendType       = "http"
	kafkaUser             = "user1"
	kafkaPassword         = "user1-secret"
	kafkaTopic            = "notification-topic"
	kafkaDLQTopic         = "notification-dlq-topic"
	kafkaConsumerGroup    = "kafka-consumer-group"
	kafkaBrokers          = "kafka1:9092,kafka2:9092,kafka3:9092"
	valkeyPassword        = "valkey"
	valkeyAddresses       = "valkey-node-0:6379"
	valkeyTTLMinutes      = "5"
)

type Suite struct {
	suite.Suite
	botContainer      testcontainers.Container
	scrapperContainer testcontainers.Container
	dbContainer       testcontainers.Container
	valkeyContainer   testcontainers.Container
	network           *testcontainers.DockerNetwork
	scrapperURL       string
	botURL            string
}

func (s *Suite) SetupSuite() {
	ctx := context.Background()

	s.setupNetwork(ctx)

	s.setupPostgres(ctx)

	s.setupValkey(ctx)

	s.setupScrapper(ctx)

	s.setupBot(ctx)

	host, err := s.scrapperContainer.Host(ctx)
	s.Require().NoError(err)

	port, err := s.scrapperContainer.MappedPort(ctx, scrapperPort)
	s.Require().NoError(err)

	s.scrapperURL = "http://" + host + ":" + port.Port()

	host, err = s.botContainer.Host(ctx)
	s.Require().NoError(err)

	port, err = s.botContainer.MappedPort(ctx, botPort)
	s.Require().NoError(err)

	s.botURL = "http://" + host + ":" + port.Port()
}

func (s *Suite) TearDownSuite() {
	if s.scrapperContainer != nil {
		err := s.scrapperContainer.Terminate(context.Background())
		if err != nil {
			return
		}
	}
	if s.botContainer != nil {
		err := s.botContainer.Terminate(context.Background())
		if err != nil {
			return
		}
	}
	if s.dbContainer != nil {
		err := s.dbContainer.Terminate(context.Background())
		if err != nil {
			return
		}
	}

	if s.valkeyContainer != nil {
		err := s.valkeyContainer.Terminate(context.Background())
		if err != nil {
			return
		}
	}

	if s.network != nil {
		err := s.network.Remove(context.Background())
		if err != nil {
			return
		}
	}
}

func (s *Suite) setupBot(ctx context.Context) {
	botReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    pathToDockerfile,
			Dockerfile: botDockerfile,
		},
		ExposedPorts: []string{botPort},
		Env: map[string]string{
			botAPIFlag:            withTelegramAPI,
			telegramAPIKey:        APIToken,
			scrapperServerAddress: scrapperServerAddr,
			updatesReceiveTypeEnv: updatesSendType,
			kafkaUserEnv:          kafkaUser,
			kafkaPasswordEnv:      kafkaPassword,
			kafkaTopicEnv:         kafkaTopic,
			kafkaDLQTopicEnv:      kafkaDLQTopic,
			kafkaConsumerGroupEnv: kafkaConsumerGroup,
			kafkaBrokersEnv:       kafkaBrokers,
			postgresHost:          dbHost,
			postgresPort:          dbPort,
			postgresUser:          dbUser,
			postgresPassword:      dbPassword,
			postgresDatabase:      dbName,
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {botAlias},
		},
		WaitingFor: wait.ForListeningPort(botPort),
	}

	botC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: botReq,
			Started:          true,
		},
	)
	s.Require().NoError(err)
	s.botContainer = botC
}

func (s *Suite) setupScrapper(ctx context.Context) {
	scrapperReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    pathToDockerfile,
			Dockerfile: scrapperDockerfile,
		},
		ExposedPorts: []string{scrapperPort},
		Env: map[string]string{
			githubAPIKey:             APIToken,
			stackoverflowAPIKey:      APIToken,
			botServerAddress:         botServerAddr,
			assetType:                assetTypeBuilder,
			"SCRAPPER_TIME_INTERVAL": "10",
			"LINKS_BATCH_SIZE":       "20",
			"SCHEDULER_NUM_WORKERS":  "5",
			postgresHost:             dbHost,
			postgresPort:             dbPort,
			postgresUser:             dbUser,
			postgresPassword:         dbPassword,
			postgresDatabase:         dbName,
			updatesSendTypeEnv:       updatesSendType,
			kafkaUserEnv:             kafkaUser,
			kafkaPasswordEnv:         kafkaPassword,
			kafkaTopicEnv:            kafkaTopic,
			kafkaBrokersEnv:          kafkaBrokers,
			valkeyPasswordEnv:        valkeyPassword,
			valkeyAddressesEnv:       valkeyAddresses,
			valkeyTTLMinutesEnv:      valkeyTTLMinutes,
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {scrapperAlias},
		},
		WaitingFor: wait.ForListeningPort(scrapperPort),
	}

	scrapperContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: scrapperReq,
			Started:          true,
		},
	)
	s.Require().NoError(err)
	s.scrapperContainer = scrapperContainer

}

func (s *Suite) setupPostgres(ctx context.Context) {
	dbReq := testcontainers.ContainerRequest{
		Image:        "postgres:18",
		ExposedPorts: []string{exposedPort},
		Env: map[string]string{
			postgresUser:     dbUser,
			postgresPassword: dbPassword,
			postgresDatabase: dbName,
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {dbHost},
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	dbC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: dbReq,
		Started:          true,
	})
	s.Require().NoError(err)
	s.dbContainer = dbC
}

func (s *Suite) setupValkey(ctx context.Context) {
	valkeyReq := testcontainers.ContainerRequest{
		Image:        "valkey/valkey:9.0.3",
		ExposedPorts: []string{valkeyPort},
		Cmd: []string{
			"valkey-server",
			"--requirepass", valkeyPassword,
			"--protected-mode", "no",
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {"valkey-node-0"},
		},
		WaitingFor: wait.ForListeningPort(valkeyPort),
	}
	valkeyC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: valkeyReq,
		Started:          true,
	})
	s.Require().NoError(err)
	s.valkeyContainer = valkeyC

}

func (s *Suite) setupNetwork(ctx context.Context) {
	net, err := network.New(ctx)
	s.Require().NoError(err)

	s.network = net
}
