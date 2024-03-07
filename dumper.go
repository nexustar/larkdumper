package larkdumper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (d *Dumper) Chat2JSON(chatID, dir string, withFile bool) error {
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

	name := fmt.Sprintf("%s-%s", chatID[len(chatID)-6:], strings.ReplaceAll(meta.Data.Name, "/", "_"))

	if withFile {
		err = os.MkdirAll(filepath.Join(dir, name), 0755)
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			var fileName, fileID string
			switch msg.MsgType {
			case "file":
				var content lark.FileContent
				err = json.Unmarshal([]byte(msg.Body.Content), &content)
				if err != nil {
					return err
				}
				fileID = content.FileKey
				fileName = content.FileName
			case "image":
				var content lark.ImageContent
				err = json.Unmarshal([]byte(msg.Body.Content), &content)
				if err != nil {
					return err
				}
				fileID = content.ImageKey
				fileName = content.ImageKey
			default:
				continue
			}

			file, err := os.Create(filepath.Join(dir, name, fileName))
			if err != nil {
				return err
			}
			defer file.Close()

			err = d.botDownloadFile(msg.MessageID, fileID, msg.MsgType, file)
			if err != nil {
				return err
			}
		}
	}

	return os.WriteFile(fmt.Sprintf("%s/%s.json", dir, name), txt, 0644)
}
