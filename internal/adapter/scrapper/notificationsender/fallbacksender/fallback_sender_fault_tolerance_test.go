package fallbacksender

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

func testUpdate() pkg.LinkUpdate {

	return pkg.LinkUpdate{
		ID:          1,
		URL:         "https://github.com/golang/go",
		Description: "new update",
		TgChatIDs:   []int64{123},
	}

}

func TestFallbackSender_SendLinkUpdate_HTTPIsSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kafkaSender := mocks.NewMockUpdateSender(ctrl)
	httpSender := mocks.NewMockUpdateSender(ctrl)

	update := testUpdate()
	eventID := "event-1"

	httpSender.EXPECT().
		SendLinkUpdate(update, "").
		Return(nil)

	kafkaSender.EXPECT().
		SendLinkUpdate(gomock.Any(), gomock.Any()).
		Times(0)

	sender := NewFallbackSender(kafkaSender, httpSender)

	err := sender.SendLinkUpdate(update, eventID)

	require.NoError(t, err)
}

func TestFallbackSender_SendLinkUpdate_HTTPFailedKafkaSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kafkaSender := mocks.NewMockUpdateSender(ctrl)
	httpSender := mocks.NewMockUpdateSender(ctrl)

	update := testUpdate()
	eventID := "event-1"
	httpErr := errors.New("http failed")

	httpSender.EXPECT().
		SendLinkUpdate(update, "").
		Return(httpErr)

	kafkaSender.EXPECT().
		SendLinkUpdate(update, eventID).
		Return(nil)

	sender := NewFallbackSender(kafkaSender, httpSender)

	err := sender.SendLinkUpdate(update, eventID)

	require.NoError(t, err)
}

func TestFallbackSender_SendLinkUpdate_HTTPFailedKafkaFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kafkaSender := mocks.NewMockUpdateSender(ctrl)
	httpSender := mocks.NewMockUpdateSender(ctrl)

	update := testUpdate()
	eventID := "event-1"

	httpErr := errors.New("http sender failed")
	kafkaErr := errors.New("kafka fallback failed")

	httpSender.EXPECT().
		SendLinkUpdate(update, "").
		Return(httpErr)

	kafkaSender.EXPECT().
		SendLinkUpdate(update, eventID).
		Return(kafkaErr)

	sender := NewFallbackSender(kafkaSender, httpSender)

	err := sender.SendLinkUpdate(update, eventID)

	require.Error(t, err)
	require.ErrorContains(t, err, "http sender failed")
	require.ErrorContains(t, err, "kafka fallback failed")
	require.ErrorIs(t, err, httpErr)
	require.ErrorIs(t, err, kafkaErr)
}
