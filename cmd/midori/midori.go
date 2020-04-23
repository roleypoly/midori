package main

import (
	"github.com/roleypoly/midori/cmd/midori/run"
	"github.com/roleypoly/midori/cmd/midori/run/watchdog"
	"github.com/roleypoly/midori/internal/version"
	"github.com/roleypoly/midori/webhooks"
	"go.uber.org/fx"
	"k8s.io/klog"
)

type klogShim struct{}

func (*klogShim) Printf(format string, args ...interface{}) {
	klog.Infof(format, args...)
}

func main() {
	klog.Infof(
		"Starting midori service.\n Build %s (%s) at %s",
		version.GitCommit,
		version.GitBranch,
		version.BuildDate,
	)

	app := fx.New(
		fx.Logger(&klogShim{}),
		fx.Provide(
			startBot,
			run.NewHealthz,
			webhooks.NewWebhookMux,
			watchdog.NewWatchdog,
		),
		fx.Invoke(startHealthz),
		fx.Invoke(startHTTP),
		fx.Invoke(startWatchdog),
	)

	app.Run()
	<-app.Done()
}
