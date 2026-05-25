package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	contentTypeJSON     = "application/json"
	tgChatHeader        = "Tg-Chat-Id"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
)

func (s *Suite) TestSendUpdatesValidRequest() {
	update := pkg.LinkUpdate{Description: "new commit", TgChatIDs: []int64{1}, ID: 1, URL: "https://github.com"}
	data, err := json.Marshal(update)
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodPost, s.botURL+"/updates", bytes.NewReader(data))
	s.Require().NoError(err)
	req.Header.Add(contentTypeKey, typeApplicationJSON)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *Suite) TestSendUpdatesInvalidRequest() {
	body := `{
		"invalid": 1
	}`

	req, err := http.NewRequest(http.MethodPost, s.botURL+"/updates", bytes.NewBufferString(body))
	s.Require().NoError(err)

	req.Header.Set(contentTypeKey, typeApplicationJSON)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.NotEqual(http.StatusOK, resp.StatusCode)
}

func (s *Suite) TestAddAndGetLink() {
	resp := s.registerChat(1)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.addLink(1, "https://github.com/golang/go")
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.getLinks(1)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	result := s.decodeListLinks(resp)
	s.Equal(int32(1), result.Size)
	s.Len(result.Links, 1)
	s.Equal("https://github.com/golang/go", result.Links[0].URL)
}

func (s *Suite) TestAddAndDeleteLink() {
	resp := s.registerChat(2)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.addLink(2, "https://github.com/golang/go")
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.deleteLink(2, "https://github.com/golang/go")
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.getLinks(2)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	result := s.decodeListLinks(resp)
	s.Equal(int32(0), result.Size)
	s.Empty(result.Links)
}

func (s *Suite) TestDeleteLinkFromNonExistingChat() {
	resp := s.registerChat(3)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.addLink(3, "https://github.com/golang/go")
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.deleteLink(999, "https://github.com/golang/go")
	defer resp.Body.Close()
	s.NotEqual(http.StatusOK, resp.StatusCode)

	resp = s.getLinks(3)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	result := s.decodeListLinks(resp)
	s.Equal(int32(1), result.Size)
	s.Len(result.Links, 1)
	s.Equal("https://github.com/golang/go", result.Links[0].URL)
}

func (s *Suite) TestAddLinkToNonExistingChat() {
	resp := s.registerChat(4)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.addLink(999, "https://github.com/golang/go")
	defer resp.Body.Close()
	s.NotEqual(http.StatusOK, resp.StatusCode)
}

func (s *Suite) TestWorkWithDeletedChat() {
	resp := s.registerChat(5)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.deleteChat(5)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	resp = s.addLink(5, "https://github.com/myrepo/go")
	defer resp.Body.Close()
	s.NotEqual(http.StatusOK, resp.StatusCode)
}

func (s *Suite) TestDeleteNonExistingChat() {
	resp := s.deleteChat(9999)
	defer resp.Body.Close()
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *Suite) TestAgentConsumesValidRawUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rawUpdate := pkg.LinkUpdate{
		Description: "This is a valid",
		TgChatIDs:   []int64{1},
		URL:         "https://github.com/myrepo/go",
	}

	body, err := json.Marshal(rawUpdate)
	s.Require().NoError(err)

	err = sendKafkaMessage(s.kafkaExternalBrokers, kafkaRawTopic, "https://github.com/myrepo/go", body)
	s.Require().NoError(err)

	msg, err := readKafkaMessage(ctx, s.kafkaExternalBrokers, kafkaProcessedTopic, "test-agent-valid-consumer")
	s.Require().NoError(err)
	var processed pkg.ProcessedLinkUpdate
	err = json.Unmarshal(msg.Value, &processed)
	s.Require().NoError(err)

	s.Equal(rawUpdate.ID, processed.ID)
	s.Equal(rawUpdate.Description+"\nСсылка: https://github.com/myrepo/go\n", processed.Description)
	s.Equal(rawUpdate.TgChatIDs, processed.TgChatIDs)
	s.Equal("MEDIUM", processed.Priority)
}

func (s *Suite) TestAgentConsumesInvalidRawUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := sendKafkaMessage(s.kafkaExternalBrokers, kafkaRawTopic, "https://github.com/myrepo/go", []byte("invalid msg"))
	s.Require().NoError(err)

	msg, err := readKafkaMessage(ctx, s.kafkaExternalBrokers, kafkaDLQTopic, "test-agent-invalid-consumer")
	s.Require().NoError(err)

	s.Contains(string(msg.Value), `"value":"invalid msg"`)
	checkCtx, checkCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer checkCancel()
	isRunning, err := isContainerRunning(checkCtx, s.agentContainer)
	s.Require().NoError(err)
	s.True(isRunning)
}

func sendKafkaMessage(brokers []string, topic string, url string, value []byte) error {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_6_0_0
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = producer.Close() }()
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(url),
		Value: sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{{
			Key:   []byte("event_id"),
			Value: []byte("event_id"),
		}},
	})
	return err
}

func readKafkaMessage(
	ctx context.Context,
	brokers []string,
	topic string,
	groupID string,

) (*sarama.ConsumerMessage, error) {

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = consumerGroup.Close() }()
	handler := &testConsumerGroupHandler{
		messageCh: make(chan *sarama.ConsumerMessage, 1),
	}
	errCh := make(chan error, 1)
	go func() {
		for {
			if consumeErr := consumerGroup.Consume(ctx, []string{topic}, handler); consumeErr != nil {
				errCh <- consumeErr
				return
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	select {
	case msg := <-handler.messageCh:
		return msg, nil
	case consumeErr := <-errCh:
		return nil, consumeErr
	case <-ctx.Done():
		return nil, ctx.Err()
	}

}

func isContainerRunning(ctx context.Context, container testcontainers.Container) (bool, error) {
	state, err := container.State(ctx)
	if err != nil {
		return false, err
	}

	return state.Running, nil
}

func (s *Suite) registerChat(chatID int64) *http.Response {
	req, err := http.NewRequest(
		http.MethodPost,
		s.scrapperURL+"/tg-chat/"+strconv.FormatInt(chatID, 10),
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	return resp
}

func (s *Suite) deleteChat(chatID int64) *http.Response {
	req, err := http.NewRequest(http.MethodDelete,
		s.scrapperURL+"/tg-chat/"+strconv.FormatInt(chatID, 10),
		nil)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *Suite) addLink(chatID int64, link string) *http.Response {
	body, err := json.Marshal(pkg.AddLinkRequest{Link: link, Tags: []string{"work"}})
	s.Require().NoError(err)

	req, err := http.NewRequest(
		http.MethodPost, s.scrapperURL+"/links", bytes.NewReader(body),
	)
	s.Require().NoError(err)

	req.Header.Set(contentTypeKey, contentTypeJSON)
	req.Header.Set(tgChatHeader, strconv.FormatInt(chatID, 10))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *Suite) deleteLink(chatID int64, link string) *http.Response {
	body, err := json.Marshal(bot.RemoveLinkRequest{Link: link})
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodDelete, s.scrapperURL+"/links", bytes.NewReader(body))
	s.Require().NoError(err)

	req.Header.Set(contentTypeKey, contentTypeJSON)
	req.Header.Set(tgChatHeader, strconv.FormatInt(chatID, 10))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *Suite) getLinks(chatID int64) *http.Response {
	req, err := http.NewRequest(
		http.MethodGet,
		s.scrapperURL+"/links",
		nil,
	)
	s.Require().NoError(err)

	req.Header.Set(tgChatHeader, strconv.FormatInt(chatID, 10))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *Suite) decodeListLinks(resp *http.Response) bot.ListLinksResponse {
	var result bot.ListLinksResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)
	return result
}

func TestRegisterGroup(t *testing.T) {
	t.Parallel()

	suite.Run(t, &Suite{})
}
