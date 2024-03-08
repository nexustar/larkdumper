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
		type larkFile struct {
			Key       string
			Name      string
			Type      string
			MessageID string
		}
		var files []larkFile

		for _, msg := range msgs {
			if msg.Deleted {
				continue
			}

			switch msg.MsgType {
			case "file":
				var content lark.FileContent
				err = json.Unmarshal([]byte(msg.Body.Content), &content)
				if err != nil {
					return err
				}
				files = append(files, larkFile{
					Key:       content.FileKey,
					Name:      content.FileName,
					Type:      "file",
					MessageID: msg.MessageID,
				})
			case "image":
				var content lark.ImageContent
				err = json.Unmarshal([]byte(msg.Body.Content), &content)
				if err != nil {
					return err
				}
				files = append(files, larkFile{
					Key:       content.ImageKey,
					Name:      content.ImageKey,
					Type:      "image",
					MessageID: msg.MessageID,
				})
			case "post":
				var content lark.PostBody
				err = json.Unmarshal([]byte(msg.Body.Content), &content)
				if err != nil {
					return err
				}
				for _, line := range content.Content {
					for _, elem := range line {
						if elem.Tag == "img" {
							files = append(files, larkFile{
								Key:       *elem.ImageKey,
								Name:      *elem.ImageKey,
								Type:      "image",
								MessageID: msg.MessageID,
							})
						}
					}
				}
			default:
				continue
			}
		}

		for _, file := range files {
			f, err := os.Create(filepath.Join(dir, name, file.Name))
			if err != nil {
				return err
			}
			defer f.Close()

			err = d.botDownloadFile(file.MessageID, file.Key, file.Type, f)
			if err != nil {
				return err
			}
		}
	}

	return os.WriteFile(fmt.Sprintf("%s/%s.json", dir, name), txt, 0644)
}
