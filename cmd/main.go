package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	vmix "github.com/FlowingSPDG/vmix-go/http"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

const prefix = ".script"

const (
	messageInvalidArgs          = "正しい引数を指定して下さい"
	messageFailedtoConnectVMIX  = "vMixへの接続に失敗しました : ```%s```"
	messageWillExctuteScript    = "スクリプト ``%s`` を実行します"
	messageSuccessExcuteScript  = "スクリプト ``%s`` を正常に実行しました"
	messafeFailedToExcuteSCript = "スクリプト ``%s`` の実行に失敗しました : ```%s```"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Println("Please specify DISCORD_TOKEN env")
		return
	}
	s := state.New("Bot " + token)

	vMixHost := os.Getenv("VMIX_HOST")
	if vMixHost == "" {
		vMixHost = "localhost"
	}
	vMixPort := 8088

	me, err := s.Me()
	if err != nil {
		log.Println("Failed to retrieve myself:", err)
		return
	}

	s.AddIntents(gateway.IntentGuilds | gateway.IntentGuildMessages)
	s.AddHandler(func(m *gateway.MessageCreateEvent) {
		if m.Author.ID == me.ID {
			return
		}
		log.Printf("%s: %s", m.Author.Username, m.Content)
		f := strings.Fields(m.Content)
		if len(f) == 0 {
			return // 空白が送られてきた
		}
		if f[0] != prefix {
			return // コマンドに該当しない
		}
		if len(f) != 2 {
			s.SendMessageReply(m.ChannelID, messageInvalidArgs, m.ID) // 引数の数が正しくない
			return
		}
		scriptName := f[1]
		s.SendMessageReply(m.ChannelID, fmt.Sprintf(messageWillExctuteScript, scriptName), m.ID) // 実行しますメッセージ

		vMixcl, err := vmix.NewClient(vMixHost, vMixPort)
		if err != nil {
			s.SendMessageReply(m.ChannelID, fmt.Sprintf(messageFailedtoConnectVMIX, err.Error()), m.ID) // vMix接続失敗メッセージ
			return
		}

		if err := vMixcl.ScriptStart(scriptName); err != nil {
			s.SendMessageReply(m.ChannelID, fmt.Sprintf(messafeFailedToExcuteSCript, scriptName, err.Error()), m.ID) // 実行失敗メッセージ
			return
		}
		s.SendMessageReply(m.ChannelID, fmt.Sprintf(messageSuccessExcuteScript, scriptName), m.ID) // 実行成功メッセージ
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := s.Open(ctx); err != nil {
		log.Println("cannot open:", err)
		return
	}

	<-ctx.Done() // block until Ctrl+C

	if err := s.Close(); err != nil {
		log.Println("cannot close:", err)
	}
}
