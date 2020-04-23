package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/lampjaw/discordgobot"
	botconfig "github.com/roleypoly/midori/internal/config"
	"github.com/roleypoly/midori/internal/version"
	"k8s.io/klog"
)

func main() {
	klog.Infof(
		"Starting midori service.\n Build %s (%s) at %s",
		version.GitCommit,
		version.GitBranch,
		version.BuildDate,
	)

	defer awaitExit()

	if botconfig.DiscordBotToken == "" {
		klog.Fatal("No bot token set, cannot launch.")
		return
	}

	config := &discordgobot.GobotConf{
		CommandPrefix: botconfig.DiscordCommandPrefix,
	}

	bot, err := discordgobot.NewBot(botconfig.DiscordBotToken, config, nil)
	if err != nil {
		klog.Fatal("Bot initialization failed.")
		return
	}

	err = bot.Open()
	if err != nil {
		klog.Fatal("Bot start failed.")
		return
	}
}

func awaitExit() {
	syscallExit := make(chan os.Signal, 1)
	signal.Notify(
		syscallExit,
		syscall.SIGINT,
		syscall.SIGTERM,
		os.Interrupt,
		os.Kill,
	)
	<-syscallExit
}
