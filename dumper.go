package larkdumper

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-lark/lark"
)

type Dumper struct {
	bot *lark.Bot
}

type ChatFile struct {
	Meta lark.ChatInfo    `json:"meta"`
	Msgs []lark.IMMessage `json:"msgs"`
}

func NewDumper(id, secret string) *Dumper {
	return &Dumper{
		bot: lark.NewChatBot(id, secret).WithUserIDType("open_id"),
	}
}

func (d *Dumper) Start() error {
	return d.bot.StartHeartbeat()
}

func (d *Dumper) Stop() {
	d.bot.StopHeartbeat()
}

func (d *Dumper) GetAllChats() ([]lark.ChatListInfo, error) {
	var chats []lark.ChatListInfo

	pageToken := ""
	for {
		resp, err := d.bot.ListChat("", pageToken, 20)
		if err != nil {
			return chats, err
		}
		chats = append(chats, resp.Data.Items...)

		if resp.Data.HasMore {
			pageToken = resp.Data.PageToken
		} else {
			break
		}
	}
	return chats, nil
}

func (d *Dumper) SearchChats(query string) ([]lark.ChatListInfo, error) {
	var chats []lark.ChatListInfo

	pageToken := ""
	for {
		resp, err := d.bot.SearchChat(query, pageToken, 20)
		if err != nil {
			return chats, err
		}
		chats = append(chats, resp.Data.Items...)

		if resp.Data.HasMore {
			pageToken = resp.Data.PageToken
		} else {
			break
		}
	}
	return chats, nil
}

func (d *Dumper) ExportChatMsgs(chatID string) ([]lark.IMMessage, error) {
	var msgs []lark.IMMessage

	pageToken := ""
	for {
		resp, err := d.botListMessage(chatID, pageToken, 50)
		if err != nil {
			return msgs, err
		}
		msgs = append(msgs, resp.Data.Items...)

		if resp.Data.HasMore {
			pageToken = resp.Data.PageToken
		} else {
			break
		}
	}

	return msgs, nil
}

func (d *Dumper) Chat2JSON(chatID, dir string) error {
	meta, err := d.bot.GetChat(chatID)
	if err != nil {
		return err
	}

	msgs, err := d.ExportChatMsgs(chatID)
	if err != nil {
		return err
	}

	file := ChatFile{
		Meta: meta.Data,
		Msgs: msgs,
	}
	txt, err := json.MarshalIndent(file, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("%s/%s.json", dir, fmt.Sprintf("%s-%s", chatID[len(chatID)-6:], meta.Data.Name)), txt, 0644)
}

type listMessageResponse struct {
	lark.BaseResponse

	Data struct {
		Items     []lark.IMMessage `json:"items"`
		PageToken string           `json:"page_token"`
		HasMore   bool             `json:"has_more"`
	} `json:"data"`
}

func (d *Dumper) botListMessage(chatID, pageToken string, pageSize int) (*listMessageResponse, error) {
	var respData listMessageResponse
	err := d.bot.GetAPIRequest("ListMessage", fmt.Sprintf("/open-apis/im/v1/messages?container_id_type=chat&container_id=%s&page_token=%s&page_size=%d", chatID, pageToken, pageSize), true, nil, &respData)
	return &respData, err
}
