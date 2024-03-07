package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-lark/lark"
	"github.com/nexustar/larkdumper"
	"github.com/spf13/cobra"
)

// cahtCmd represents the chat command
func newChatCmd() *cobra.Command {
	var all, withFile bool
	var resultPath string
	cmd := &cobra.Command{
		Use:   "chats <title query>",
		Short: "dump chats",
		Long:  `dump chats. Use environment variables LARK_APP_ID and LARK_APP_SECRET to set app id and secret.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := os.Getenv("LARK_APP_ID")
			appSecret := os.Getenv("LARK_APP_SECRET")
			if appID == "" || appSecret == "" {
				return errors.New("LARK_APP_ID and LARK_APP_SECRET are required")
			}
			dumper := larkdumper.NewDumper(appID, appSecret)
			err := dumper.Start()
			if err != nil {
				return err
			}
			defer dumper.Stop()

			var chats []lark.ChatListInfo
			if all {
				chats, err = dumper.GetAllChats()
			} else {
				if len(args) == 0 {
					cmd.Help()
					return errors.New("title query is required")
				}
				chats, err = dumper.SearchChats(args[0])
			}
			if err != nil {
				return err
			}

			for _, chat := range chats {
				fmt.Printf("---dumping chat %s---\n", chat.Name)
				err = dumper.Chat2JSON(chat.ChatID, resultPath, withFile)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "dump all accessible chats instead of searching by title")
	cmd.Flags().StringVarP(&resultPath, "path", "p", ".", "target path")
	cmd.Flags().BoolVar(&withFile, "with-file", false, "dump chat files")

	return cmd
}

func init() {
	rootCmd.AddCommand(newChatCmd())
}
