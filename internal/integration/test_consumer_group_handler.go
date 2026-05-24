package integration

import "github.com/IBM/sarama"

type testConsumerGroupHandler struct {
	messageCh chan *sarama.ConsumerMessage
}

func (h *testConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *testConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *testConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for message := range claim.Messages() {
		session.MarkMessage(message, "")

		select {
		case h.messageCh <- message:
		default:
		}
	}

	return nil
}
