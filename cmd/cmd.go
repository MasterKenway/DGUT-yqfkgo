package cmd

import (
	"DGUT-yqfkgo/internal/config"
	"DGUT-yqfkgo/internal/constant"
	"DGUT-yqfkgo/internal/log"
	"DGUT-yqfkgo/internal/push"
	"DGUT-yqfkgo/internal/service"
)

import (
	"github.com/urfave/cli/v2"
)

func App() *cli.App {
	return &cli.App{
		Name: "DGUT Epidemic Report",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "username",
				Value:    "",
				Aliases:  []string{"u"},
				EnvVars:  []string{constant.ENV_USERNAME},
				Usage:    "DGUT Ehall Username",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Value:    "",
				Aliases:  []string{"p"},
				EnvVars:  []string{constant.ENV_PASSWORD},
				Usage:    "DGUT Ehall Password",
				Required: true,
			},
			//&cli.StringFlag{
			//	Name:    "sckey",
			//	Aliases: []string{"k"},
			//	EnvVars: []string{constant.ENV_SCKEY},
			//	Usage:   "Message Push Service",
			//	Required: false,
			//},
			&cli.StringFlag{
				Name:     "tgToken",
				Value:    "",
				Aliases:  []string{"t"},
				EnvVars:  []string{constant.ENV_SCKEY},
				Usage:    "Telegram Bot Token",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "chatId",
				Value:    "",
				Aliases:  []string{"c"},
				EnvVars:  []string{constant.ENV_SCKEY},
				Usage:    "Telegram Chat Id (Attain From @userinfobot)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "runat, r",
				Aliases:  []string{"r"},
				EnvVars:  []string{constant.ENV_RUNAT},
				Usage:    "Schedule Time",
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			conf := &config.Config{
				Username:   ctx.String(constant.KEY_CMD_USERNAME),
				Password:   ctx.String(constant.KEY_CMD_PASSWORD),
				TgBotToken: ctx.String(constant.KEY_CMD_TGTOKEN),
				ChatId:     ctx.String(constant.KEY_CMD_CHATID),
				Sckey:      ctx.String(constant.KEY_CMD_SCKEY),
				RunAt:      ctx.String(constant.KEY_CMD_RUNAT),
			}

			log.Info().Msgf("Username: %s", conf.Username)
			log.Info().Msgf("Password: %s", conf.Password)

			if len(conf.TgBotToken) != 0 {
				if len(conf.ChatId) == 0 {
					log.Fatal().Msgf("Chat Id is Needed")
				}
				push.NewTgPusherWrapper(conf)
			}
			service.Start(conf)

			return nil
		},
	}
}
