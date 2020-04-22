package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lampjaw/discordgobot"
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
