package botconfig

import (
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type channelConfig struct {
	PublicNotifications  string
	PrivateNotifications string
}

type tfCloudConfig struct {
	Org   string
	Token string
}

type webhookTokens struct {
	GitHub  []string
	TfCloud []string
	Actions []string
}

var (
	DiscordCommandPrefix = os.Getenv("DISCORD_COMMAND_PREFIX")
	DiscordBotToken      = os.Getenv("DISCORD_BOT_TOKEN")
	AuthorizedUsers      = os.Getenv("AUTHORIZED_USERS")
	ServicePort          = os.Getenv("MIDORI_SVC_PORT")
	HealthzPort          = ":1" + os.Getenv("MIDORI_SVC_PORT")[1:]
	Channels             = channelConfig{
		PublicNotifications:  os.Getenv("PUBLIC_CHANNEL"),
		PrivateNotifications: os.Getenv("PRIVATE_CHANNEL"),
	}
	TerraformCloud = tfCloudConfig{
		Org:   os.Getenv("TFC_ORG"),
		Token: os.Getenv("TFC_TOKEN"),
	}
	WebhookTokens = webhookTokens{
		GitHub:  strings.Split(os.Getenv("GITHUB_WEBHOOK_TOKENS"), ","),
		TfCloud: strings.Split(os.Getenv("TFC_WEBHOOK_TOKENS"), ","),
		Actions: strings.Split(os.Getenv("ACTIONS_WEBHOOK_TOKENS"), ","),
	}
)
