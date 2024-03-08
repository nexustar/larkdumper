package larkdumper

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-lark/lark"
)

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

func (d *Dumper) botDownloadFile(messageID, fileKey, fileType string, output io.Writer) error {
	header := make(http.Header)
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Add("Authorization", fmt.Sprintf("Bearer %s", d.bot.TenantAccessToken()))
	url := d.bot.ExpandURL(fmt.Sprintf("/open-apis/im/v1/messages/%s/resources/%s?type=%s", messageID, fileKey, fileType))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header = header

	client := &http.Client{
		Timeout: 0,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download file failed, status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(output, resp.Body)
	return err
}
