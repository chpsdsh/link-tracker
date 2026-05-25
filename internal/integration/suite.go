package integration

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	scrapperDockerfile = "scrapper.Dockerfile"
	botDockerfile      = "bot.Dockerfile"
	scrapperPort       = "8081/tcp"
	botPort            = "8080/tcp"
	scrapperAlias      = "scrapper"
	botAlias           = "bot"

	telegramAPIKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
	githubAPIKey          = "GITHUB_API_KEY"
	stackoverflowAPIKey   = "STACKOVERFLOW_API_KEY"
	botServerAddress      = "BOT_SERVER_ADDRESS"
	botAPIFlag            = "WITH_TELEGRAM_API"

	scrapperTimeIntervalEnv = "SCRAPPER_TIME_INTERVAL"
	scrapperTimeInterval    = "10"
	linksBatchSizeEnv       = "LINKS_BATCH_SIZE"
	linksBatchSize          = "20"
	schedulerNumWorkersEnv  = "SCHEDULER_NUM_WORKERS"
	schedulerNumWorkers     = "5"

	APIToken           = "API_TOKEN"
	botServerAddr      = "http://bot:8080"
	scrapperServerAddr = "http://scrapper:8081"
	withTelegramAPI    = "false"

	pathToDockerfile = "../../"
	assetType        = "ASSET_TYPE"
	assetTypeBuilder = "BUILDER"

	dbUser      = "user"
	dbPassword  = "password"
	dbName      = "mydb"
	exposedPort = "5432/tcp"
	valkeyPort  = "6379/tcp"
	dbPort      = "5432"
	dbHost      = "postgres"

	postgresHost     = "POSTGRES_HOST"
	postgresPort     = "POSTGRES_PORT"
	postgresUser     = "POSTGRES_USER"
	postgresPassword = "POSTGRES_PASSWORD"
	postgresDatabase = "POSTGRES_DB"

	updatesHandleTypeEnv = "UPDATES_HANDLE_TYPE"
	updatesHandleType    = "fallback"

	kafkaConsumerGroupEnv = "KAFKA_CONSUMER_GROUP"
	kafkaUserEnv          = "KAFKA_USER"
	kafkaPasswordEnv      = "KAFKA_PASSWORD"
	kafkaTopicEnv         = "KAFKA_TOPIC"
	kafkaBrokersEnv       = "KAFKA_BROKERS"
	kafkaDLQTopicEnv      = "KAFKA_DLQ_TOPIC"

	kafkaUser                         = "user1"
	kafkaPassword                     = "user1-secret"
	kafkaRawTopic                     = "raw-notification-topic"
	kafkaProcessedTopic               = "processed-notification-topic"
	kafkaDLQTopic                     = "notification-dlq"
	kafkaRawTopicEnv                  = "KAFKA_RAW_TOPIC"
	kafkaProcessedTopicEnv            = "KAFKA_PROCESSED_TOPIC"
	kafkaConsumerGroup                = "notification-consumers"
	kafkaRawNotificationConsumerGroup = "raw-notification-consumers"
	kafkaBrokers                      = "kafka1:9092"

	valkeyPasswordEnv   = "VALKEY_PASSWORD"
	valkeyAddressesEnv  = "VALKEY_ADDRESSES"
	valkeyTTLMinutesEnv = "VALKEY_TTL_MINUTES"

	valkeyPassword   = "valkey"
	valkeyAddresses  = "valkey-node-0:6379"
	valkeyTTLMinutes = "5"

	httpClientTimeoutEnv = "HTTP_CLIENT_TIMEOUT"
	retryMaxAttemptsEnv  = "RETRY_MAX_ATTEMPTS"
	retryDelayEnv        = "RETRY_DELAY"
	retryableStatusesEnv = "RETRYABLE_STATUSES"

	httpClientTimeout = "10s"
	retryMaxAttempts  = "3"
	retryDelay        = "500ms"
	retryableStatuses = "500,502,503,504"

	circuitBreakerIntervalEnv     = "CIRCUIT_BREAKER_INTERVAL"
	circuitBreakerTimeoutEnv      = "CIRCUIT_BREAKER_TIMEOUT"
	circuitBreakerMaxRequestsEnv  = "CIRCUIT_BREAKER_MAX_REQUESTS"
	circuitBreakerFailureRatioEnv = "CIRCUIT_BREAKER_FAILURE_RATIO"

	circuitBreakerInterval     = "10s"
	circuitBreakerTimeout      = "5s"
	circuitBreakerMaxRequests  = "3"
	circuitBreakerFailureRatio = "0.6"

	rateLimitRPSEnv   = "RATE_LIMIT_RPS"
	rateLimitBurstEnv = "RATE_LIMIT_BURST"

	rateLimitRPS   = "20"
	rateLimitBurst = "20"

	agentDockerfile             = "agent.dockerfile"
	agentAlias                  = "agent"
	kafkaImage                  = "confluentinc/cp-kafka:7.8.0"
	kafkaAlias                  = "kafka1"
	aiStopWordsEnv              = "AI_STOP_WORDS"
	aiExcludedAuthorsEnv        = "AI_EXCLUDED_AUTHORS"
	aiMinLengthEnv              = "AI_MIN_LENGTH"
	aiSummarizationThresholdEnv = "AI_SUMMARIZATION_THRESHOLD"
	aiStopWords                 = "spam,ads,promo"
	aiExcludedAuthors           = "bot-user"
	aiMinLength                 = "2"
	aiSummarizationThreshold    = "500"
	yandexAPIKeyEnv             = "YANDEX_API_KEY"
	yandexFolderIDEnv           = "YANDEX_FOLDER_ID"
	yandexModelEnv              = "YANDEX_MODEL"
	yandexBaseURLEnv            = "YANDEX_BASE_URL"
	yandexAPIKey                = "test-api-key"
	yandexFolderID              = "test-folder-id"
	yandexModel                 = "yandexgpt-5-lite"
	yandexBaseURL               = "http://localhost:9999/v1"
	aiHighPriorityKeyWordsEnv   = "AI_HIGH_PRIORITY_KEY_WORDS"
	groupWindowMSEnv            = "GROUP_WINDOW_MS"
	aiLowPriorityKeyWordsEnv    = "AI_LOW_PRIORITY_KEY_WORDS"
	aiHighPriorityKeyWordsValue = "critical,urgent,breaking"
	groupWindowMSValue          = "3000"
	aiLowPriorityKeyWordsValue  = "minor,typo,docs,chore"
)

type Suite struct {
	suite.Suite
	botContainer         testcontainers.Container
	scrapperContainer    testcontainers.Container
	dbContainer          testcontainers.Container
	valkeyContainer      testcontainers.Container
	kafkaContainer       testcontainers.Container
	kafkaInitContainer   testcontainers.Container
	agentContainer       testcontainers.Container
	network              *testcontainers.DockerNetwork
	scrapperURL          string
	botURL               string
	kafkaExternalBrokers []string
}

func (s *Suite) SetupSuite() {
	ctx := context.Background()

	s.setupNetwork(ctx)
	s.setupPostgres(ctx)
	s.setupValkey(ctx)
	s.setupKafka(ctx)
	s.setupKafkaTopics(ctx)
	s.setupScrapper(ctx)
	s.setupBot(ctx)
	s.setupAgent(ctx)

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

	ctx := context.Background()
	if s.agentContainer != nil {
		_ = s.agentContainer.Terminate(ctx)
	}
	if s.scrapperContainer != nil {
		_ = s.scrapperContainer.Terminate(ctx)
	}
	if s.botContainer != nil {
		_ = s.botContainer.Terminate(ctx)
	}
	if s.kafkaInitContainer != nil {
		_ = s.kafkaInitContainer.Terminate(ctx)
	}
	if s.kafkaContainer != nil {
		_ = s.kafkaContainer.Terminate(ctx)
	}
	if s.dbContainer != nil {
		_ = s.dbContainer.Terminate(ctx)
	}
	if s.valkeyContainer != nil {
		_ = s.valkeyContainer.Terminate(ctx)
	}
	if s.network != nil {
		_ = s.network.Remove(ctx)
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
			telegramAPIKey:        APIToken,
			scrapperServerAddress: scrapperServerAddr,
			botAPIFlag:            withTelegramAPI,

			updatesHandleTypeEnv: updatesHandleType,

			kafkaUserEnv:           kafkaUser,
			kafkaPasswordEnv:       kafkaPassword,
			kafkaRawTopicEnv:       kafkaRawTopic,
			kafkaProcessedTopicEnv: kafkaProcessedTopic,
			kafkaDLQTopicEnv:       kafkaDLQTopic,
			kafkaConsumerGroupEnv:  kafkaConsumerGroup,
			kafkaBrokersEnv:        kafkaBrokers,

			postgresHost:     dbHost,
			postgresPort:     dbPort,
			postgresUser:     dbUser,
			postgresPassword: dbPassword,
			postgresDatabase: dbName,

			httpClientTimeoutEnv: httpClientTimeout,
			retryMaxAttemptsEnv:  retryMaxAttempts,
			retryDelayEnv:        retryDelay,
			retryableStatusesEnv: retryableStatuses,

			circuitBreakerIntervalEnv:     circuitBreakerInterval,
			circuitBreakerTimeoutEnv:      circuitBreakerTimeout,
			circuitBreakerMaxRequestsEnv:  circuitBreakerMaxRequests,
			circuitBreakerFailureRatioEnv: circuitBreakerFailureRatio,

			rateLimitRPSEnv:   rateLimitRPS,
			rateLimitBurstEnv: rateLimitBurst,
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {botAlias},
		},
	}

	botC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: botReq,
			Started:          true,
		},
	)
	if err != nil {
		fmt.Printf("Failed to start bot container: %s\n", err)
	}
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
			githubAPIKey:                  APIToken,
			stackoverflowAPIKey:           APIToken,
			botServerAddress:              botServerAddr,
			scrapperTimeIntervalEnv:       scrapperTimeInterval,
			linksBatchSizeEnv:             linksBatchSize,
			schedulerNumWorkersEnv:        schedulerNumWorkers,
			assetType:                     assetTypeBuilder,
			postgresHost:                  dbHost,
			postgresPort:                  dbPort,
			postgresUser:                  dbUser,
			postgresPassword:              dbPassword,
			postgresDatabase:              dbName,
			updatesHandleTypeEnv:          updatesHandleType,
			kafkaUserEnv:                  kafkaUser,
			kafkaPasswordEnv:              kafkaPassword,
			kafkaTopicEnv:                 kafkaRawTopic,
			kafkaBrokersEnv:               kafkaBrokers,
			valkeyPasswordEnv:             valkeyPassword,
			valkeyAddressesEnv:            valkeyAddresses,
			valkeyTTLMinutesEnv:           valkeyTTLMinutes,
			httpClientTimeoutEnv:          httpClientTimeout,
			retryMaxAttemptsEnv:           retryMaxAttempts,
			retryDelayEnv:                 retryDelay,
			retryableStatusesEnv:          retryableStatuses,
			circuitBreakerIntervalEnv:     circuitBreakerInterval,
			circuitBreakerTimeoutEnv:      circuitBreakerTimeout,
			circuitBreakerMaxRequestsEnv:  circuitBreakerMaxRequests,
			circuitBreakerFailureRatioEnv: circuitBreakerFailureRatio,
			rateLimitRPSEnv:               rateLimitRPS,
			rateLimitBurstEnv:             rateLimitBurst,
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

func (s *Suite) setupKafka(ctx context.Context) {
	kafkaReq := testcontainers.ContainerRequest{
		Image:        kafkaImage,
		ExposedPorts: []string{"9092/tcp", "9094/tcp"},
		Env: map[string]string{
			"KAFKA_NODE_ID":       "1",
			"KAFKA_BROKER_ID":     "1",
			"KAFKA_PROCESS_ROLES": "broker,controller",

			"KAFKA_CONTROLLER_QUORUM_VOTERS": "1@kafka1:9093",

			"KAFKA_LISTENERS": "INTERNAL://kafka1:9092,EXTERNAL://0.0.0.0:9094,CONTROLLER://kafka1:9093",

			"KAFKA_ADVERTISED_LISTENERS": "INTERNAL://kafka1:9092,EXTERNAL://127.0.0.1:9094",

			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_CONTROLLER_LISTENER_NAMES":      "CONTROLLER",
			"KAFKA_INTER_BROKER_LISTENER_NAME":     "INTERNAL",

			"CLUSTER_ID": "EmptNWtoR4GGWx-BH6nGLQ",

			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_DEFAULT_REPLICATION_FACTOR":               "1",
			"KAFKA_MIN_INSYNC_REPLICAS":                      "1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                "true",
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = nat.PortMap{
				"9094/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "9094",
					},
				},
			}
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {kafkaAlias},
		},
		WaitingFor: wait.ForListeningPort("9094/tcp"),
	}

	kafkaC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: kafkaReq,
			Started:          true,
		},
	)
	s.Require().NoError(err)

	s.kafkaContainer = kafkaC

	s.kafkaExternalBrokers = []string{"127.0.0.1:9094"}
}

func (s *Suite) setupKafkaTopics(ctx context.Context) {
	initReq := testcontainers.ContainerRequest{
		Image: kafkaImage,
		Entrypoint: []string{
			"/bin/sh",
			"-c",
		},
		Cmd: []string{
			`
echo "Waiting for Kafka..."
cub kafka-ready -b kafka1:9092 1 60

echo "Creating topics..."
kafka-topics --bootstrap-server kafka1:9092 --create --if-not-exists --topic raw-notification-topic --replication-factor 1 --partitions 3
kafka-topics --bootstrap-server kafka1:9092 --create --if-not-exists --topic processed-notification-topic --replication-factor 1 --partitions 3
kafka-topics --bootstrap-server kafka1:9092 --create --if-not-exists --topic notification-dlq --replication-factor 1 --partitions 3

echo "Topics:"
kafka-topics --bootstrap-server kafka1:9092 --list
`,
		},
		Networks:   []string{s.network.Name},
		WaitingFor: wait.ForLog("Topics:"),
	}

	initC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: initReq,
			Started:          true,
		},
	)
	s.Require().NoError(err)

	s.kafkaInitContainer = initC
}

func (s *Suite) setupAgent(ctx context.Context) {

	agentReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    pathToDockerfile,
			Dockerfile: agentDockerfile,
		},
		Env: map[string]string{
			postgresHost:                dbHost,
			postgresPort:                dbPort,
			postgresUser:                dbUser,
			postgresPassword:            dbPassword,
			postgresDatabase:            dbName,
			kafkaUserEnv:                kafkaUser,
			kafkaPasswordEnv:            kafkaPassword,
			kafkaBrokersEnv:             kafkaBrokers,
			kafkaDLQTopicEnv:            kafkaDLQTopic,
			kafkaConsumerGroupEnv:       kafkaRawNotificationConsumerGroup,
			kafkaRawTopicEnv:            kafkaRawTopic,
			kafkaProcessedTopicEnv:      kafkaProcessedTopic,
			aiStopWordsEnv:              aiStopWords,
			aiExcludedAuthorsEnv:        aiExcludedAuthors,
			aiMinLengthEnv:              aiMinLength,
			aiSummarizationThresholdEnv: aiSummarizationThreshold,
			yandexAPIKeyEnv:             yandexAPIKey,
			yandexFolderIDEnv:           yandexFolderID,
			yandexModelEnv:              yandexModel,
			yandexBaseURLEnv:            yandexBaseURL,
			aiHighPriorityKeyWordsEnv:   aiHighPriorityKeyWordsValue,
			aiLowPriorityKeyWordsEnv:    aiLowPriorityKeyWordsValue,
			groupWindowMSEnv:            groupWindowMSValue,
		},
		Networks: []string{s.network.Name},
		NetworkAliases: map[string][]string{
			s.network.Name: {agentAlias},
		},
	}
	agentC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: agentReq,
			Started:          true,
		},
	)
	s.Require().NoError(err)
	s.agentContainer = agentC

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
