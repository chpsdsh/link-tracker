package integration

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type FakeTgAPI struct {
	Updates tgbotapi.UpdatesChannel
	Sent    []tgbotapi.Chattable
}

func NewIntegrationTgAPI() *FakeTgAPI {
	return &FakeTgAPI{
		Updates: make(chan tgbotapi.Update),
		Sent:    make([]tgbotapi.Chattable, 0),
	}
}

func (in *FakeTgAPI) GetUpdatesChan(_ tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return in.Updates
}

func (in *FakeTgAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	in.Sent = append(in.Sent, c)

	msg := tgbotapi.Message{}

	if mc, ok := c.(tgbotapi.MessageConfig); ok {
		msg.Chat = &tgbotapi.Chat{ID: mc.ChatID}
		msg.Text = mc.Text
	}

	return msg, nil
}

func (in *FakeTgAPI) Request(_ tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{
		Ok: true,
	}, nil
}
