package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
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

func (s *Suite) registerChat(chatID int64) *http.Response {
	req, err := http.NewRequest(
		http.MethodPost,
		s.scrapperURL+"/tg-chat/"+strconv.FormatInt(chatID, 10),
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	println(resp.Status)
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
