package integration

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	scrapperImage         = "scrapper-image:latest"
	botImage              = "bot-image:latest"
	scrapperPort          = "8081/tcp"
	botPort               = "8080/tcp"
	scrapperAlias         = "scrapper"
	botAlias              = "bot"
	telegramAPIKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
	networkName           = "linktracker-test-network"
	githubAPIKey          = "GITHUB_API_KEY"
	stackoverflowAPIKey   = "STACKOVERFLOW_API_KEY"
	botServerAddress      = "BOT_SERVER_ADDRESS"
	botAPIFlag            = "WITH_TELEGRAM_API"
	APIToken              = "API_TOKEN"
	botServerAddr         = "http://bot:8080"
	scrapperServerAddr    = "http://scrapper:8081"
	withTelegramAPI       = "false"
)

type Suite struct {
	suite.Suite
	botContainer      testcontainers.Container
	scrapperContainer testcontainers.Container
	scrapperURL       string
	botURL            string
}

func (s *Suite) SetupSuite() {
	ctx := context.Background()

	scrapperReq := testcontainers.ContainerRequest{
		Image:        scrapperImage,
		ExposedPorts: []string{scrapperPort},
		Env: map[string]string{
			githubAPIKey:        APIToken,
			stackoverflowAPIKey: APIToken,
			botServerAddress:    botServerAddr,
		},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {scrapperAlias},
		},
		WaitingFor: wait.ForListeningPort(scrapperPort),
	}

	botReq := testcontainers.ContainerRequest{
		Image:        botImage,
		ExposedPorts: []string{botPort},
		Env: map[string]string{
			botAPIFlag:            withTelegramAPI,
			telegramAPIKey:        APIToken,
			scrapperServerAddress: scrapperServerAddr,
		},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {botAlias},
		},
		WaitingFor: wait.ForListeningPort(botPort),
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

	botC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: botReq,
			Started:          true,
		},
	)
	s.Require().NoError(err)
	s.botContainer = botC

	host, err := scrapperContainer.Host(ctx)
	s.Require().NoError(err)

	port, err := scrapperContainer.MappedPort(ctx, scrapperPort)
	s.Require().NoError(err)

	s.scrapperURL = "http://" + host + ":" + port.Port()

	host, err = botC.Host(ctx)
	s.Require().NoError(err)

	port, err = botC.MappedPort(ctx, botPort)
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
}
