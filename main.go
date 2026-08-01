package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ras0q/goalie"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/config"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/database"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/injector"
	"github.com/traP-jp/anshin-techo-backend/internal/repository" // 追加
	"github.com/traP-jp/anshin-techo-backend/internal/service/bot"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("runtime error: %+v", err)
	}
}

func run() (err error) {
	g := goalie.New()
	defer g.Collect(&err)

	var c config.Config
	c.Parse()

	// データベースに接続してマイグレーション
	db, err := database.Setup(c.MySQLConfig())
	if err != nil {
		return err
	}
	defer g.Guard(db.Close)

	// Repository の初期化 (Schedulerで使うために必要)
	repo := repository.New(db)

	// Bot サービスの初期化 
	botService, err := bot.NewService(bot.Config{
		Origin:                os.Getenv("TRAQ_ORIGIN"),
		AccessToken:           os.Getenv("TRAQ_BOT_TOKEN"),
		TicketCreateChannelID: os.Getenv("TRAQ_CHANNEL_TICKET_CREATE"),
		TicketUpdateChannelID: os.Getenv("TRAQ_CHANNEL_TICKET_UPDATE"),
		NoteIncomingChannelID: os.Getenv("TRAQ_CHANNEL_NOTE_INCOMING"),
		NoteOutgoingChannelID: os.Getenv("TRAQ_CHANNEL_NOTE_OUTGOING"),
		NoteOtherChannelID:    os.Getenv("TRAQ_CHANNEL_NOTE_OTHER"),
		ReviewNotifyChannelID: os.Getenv("TRAQ_CHANNEL_REVIEW_NOTIFY"),
		ManagerID:             os.Getenv("TRAQ_USER_MANAGER_ID"),
	})
	if err != nil {
		return err
	}

	//  Scheduler (リマインダー) の初期化と起動
	scheduler := bot.NewScheduler(repo, botService)
	scheduler.Start()

	//  Bot のイベントハンドラを登録 (Handler側)
	botHandlerService := injector.InjectBotHandlerService(injector.Dependencies{
		DB:  db,
		Bot: botService,
	})
	botHandlerService.RegisterHandlers(botService)

	//  サーバーの初期化
	server, err := injector.InjectServer(injector.Dependencies{
		DB:  db,
		Bot: botService,
	})
	if err != nil {
		return err
	}

	// HTTP サーバーを goroutine で起動
	go func() {
		if err := http.ListenAndServe(c.AppAddr, server); err != nil {
			log.Fatalf("runtime error: %+v", err)
		}
	}()

	// Bot の WebSocket 接続を開始
	if err := botService.Start(); err != nil {
		return err
	}

	return nil
}