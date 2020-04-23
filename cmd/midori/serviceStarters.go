package main

import (
	"context"
	"net/http"

	"github.com/lampjaw/discordgobot"
	"github.com/roleypoly/midori/cmd/midori/run"
	"github.com/roleypoly/midori/cmd/midori/run/watchdog"
	botconfig "github.com/roleypoly/midori/internal/config"
	"github.com/roleypoly/midori/webhooks"
	"go.uber.org/fx"
	"k8s.io/klog"
)

func startHTTP(lc fx.Lifecycle, healthz *run.Healthz, mux *webhooks.WebhookMux) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			klog.Info("startHTTP: OnStart fired")
			go func() {

				err := http.ListenAndServe(botconfig.ServicePort, mux.Mux)
				if err != nil {
					klog.Fatal("startHTTP: ", err)
				}
			}()
			healthz.WebhooksReady = true
			return nil
		},
		OnStop: func(ctx context.Context) error {
			klog.Info("startHTTP: OnStop fired")

			healthz.WebhooksReady = false
			return nil
		},
	})
}

func startBot(lc fx.Lifecycle, healthz *run.Healthz) *discordgobot.Gobot {
	if botconfig.DiscordBotToken == "" {
		klog.Fatal("No bot token set, cannot launch.")
		return nil
	}

	config := &discordgobot.GobotConf{
		CommandPrefix: botconfig.DiscordCommandPrefix,
	}

	bot, err := discordgobot.NewBot(botconfig.DiscordBotToken, config, nil)
	if err != nil {
		klog.Fatal("Bot initialization failed.")
		return nil
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			klog.Info("startBot: OnStart fired")
			go func() {
				err := bot.Open()
				if err != nil {
					klog.Fatal("startBot: Bot start failed.")
				}
			}()

			healthz.DiscordReady = true

			return nil
		},
		OnStop: func(ctx context.Context) error {
			healthz.DiscordReady = false

			return nil
		},
	})

	return bot
}

func startHealthz(lc fx.Lifecycle, healthz *run.Healthz) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			klog.Info("startHealthz: OnStart fired")
			go healthz.Start(botconfig.HealthzPort)
			return nil
		},
	})
}

func startWatchdog(lc fx.Lifecycle, wd *watchdog.Watchdog, bot *discordgobot.Gobot) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			klog.Info("startWatchdog: OnStart fired")

			go wd.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			klog.Info("startWatchdog: OnStop fired")

			wd.Stop()
			return nil
		},
	})
}
