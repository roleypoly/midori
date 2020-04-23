package watchdog

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lampjaw/discordgobot"
	botconfig "github.com/roleypoly/midori/internal/config"
	"k8s.io/klog"
)

type watchdogState struct {
	lastSeen time.Time
	alerted  bool
}

type Watchdog struct {
	Targets []string
	Timeout time.Duration

	bot        *discordgobot.Gobot
	states     map[string]watchdogState
	stateMutex sync.RWMutex
	ticker     *time.Ticker
	quit       chan bool
}

func NewWatchdog(bot *discordgobot.Gobot) *Watchdog {
	timeout, err := time.ParseDuration(botconfig.WatchdogTimeout)
	if err != nil {
		klog.Warningf("WATCHDOG_TIMEOUT value of `%s` did not parse. Defaulting to 15 minutes.", botconfig.WatchdogTimeout)
		timeout = 15 * time.Minute
	}

	klog.Info("watchdog: watching with a timeout of ", timeout.String())

	return &Watchdog{
		Timeout: timeout,
		Targets: botconfig.WatchdogTargets,

		bot:    bot,
		states: map[string]watchdogState{},
		quit:   make(chan bool),
	}
}

func (w *Watchdog) Start() {
	w.stateMutex.Lock()
	initialStates := map[string]watchdogState{}

	for _, id := range w.Targets {
		initialStates[id] = watchdogState{
			lastSeen: time.Now(),
			alerted:  false,
		}
	}

	w.states = initialStates
	w.stateMutex.Unlock()

	go w.startTimers()
	return
}

func (w *Watchdog) Stop() {
	w.ticker.Stop()
	w.quit <- true
}

func (w *Watchdog) startTimers() {
	klog.V(1).Info("watchdog: starting ticker")

	interval := 30 * time.Second

	w.ticker = time.NewTicker(interval)
	for {
		select {
		case <-w.quit:
			klog.V(1).Info("watchdog: ticker quit")
			return
		case <-w.ticker.C:
			klog.V(1).Info("watchdog: ticker ticked")
			w.runChecks()
		}
	}
}

type lastStateContextType string

var lastStateContext = lastStateContextType("lastState")

func (w *Watchdog) runChecks() {
	for _, id := range w.Targets {
		p, err := w.bot.Client.Session.State.Presence(botconfig.PrimaryGuild, id)
		if err != nil {
			gm, err := w.bot.Client.GuildMember(botconfig.PrimaryGuild, id)
			if err != nil {
				klog.Errorf("watchdog: Error finding user %s, %v", id, err)
				return
			}

			p.User = gm.User
		}

		w.stateMutex.RLock()
		lastState := w.states[id]
		w.stateMutex.RUnlock()

		ctx := context.WithValue(context.Background(), lastStateContext, lastState)

		if p.Status != discordgo.StatusOffline {
			if lastState.alerted {
				klog.Infof("watchdog: user %s recovered", id)
				go w.RecoveryHandler(ctx, p.User)
			}

			nextState := watchdogState{
				lastSeen: time.Now(),
				alerted:  false,
			}

			w.stateMutex.Lock()
			w.states[id] = nextState
			w.stateMutex.Unlock()

			klog.V(1).Infof("watchdog: %s successfully checked", id)
		} else {
			if lastState.lastSeen.Add(w.Timeout).Before(time.Now()) {
				klog.Warningf("watchdog: user %s timed out!!!", id)
				go w.FailureHandler(ctx, p.User)
				nextState := watchdogState{
					lastSeen: lastState.lastSeen,
					alerted:  true,
				}

				w.stateMutex.Lock()
				w.states[id] = nextState
				w.stateMutex.Unlock()
			}

			klog.Infof("watchdog: user %s failed check", id)
		}
	}
}

func (w *Watchdog) RecoveryHandler(ctx context.Context, user *discordgo.User) {
	channel := botconfig.Channels.PrivateNotifications
	lastState := ctx.Value(lastStateContext).(watchdogState)
	klog.V(1).Infof("watchdog: sending recovery alert to %s", channel)

	embed := &discordgo.MessageEmbed{
		Color: 0x2bd64d,
		Title: "Midori Watchdog",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "State",
				Value: "Recovered",
			},
			{
				Name:  "Bot",
				Value: user.Mention(),
			},
			{
				Name:  "Downtime",
				Value: time.Since(lastState.lastSeen).String(),
			},
		},
	}

	err := w.bot.Client.DiscordClient.SendEmbedMessage(channel, embed)
	if err != nil {
		klog.Error()
	}
}

func (w *Watchdog) FailureHandler(ctx context.Context, user *discordgo.User) {
	channel := botconfig.Channels.PrivateNotifications
	lastState := ctx.Value(lastStateContext).(watchdogState)
	klog.V(1).Infof("watchdog: sending failure alert to %s", channel)

	embed := &discordgo.MessageEmbed{
		Color: 0xff0033,
		Title: "Midori Watchdog",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "State",
				Value: "Down",
			},
			{
				Name:  "Bot",
				Value: user.Mention(),
			},
			{
				Name:  "Downtime",
				Value: time.Since(lastState.lastSeen).String(),
			},
		},
	}

	err := w.bot.Client.DiscordClient.SendEmbedMessage(channel, embed)
	if err != nil {
		klog.Error()
	}
}
