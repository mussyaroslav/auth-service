package logger

import (
	"auth-service/config"
	"auth-service/pkg/logger/handlers/slogpretty"
	"auth-service/pkg/logger/timeformatter"
	"io"
	"log/slog"
	"os"
	"strings"
)

// константы уровеня логирования для окружения развёртывания
const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Initial(cfg *config.Config) (*slog.Logger, *os.File) {
	var logger *slog.Logger
	var logFile *os.File
	var err error
	var outW io.Writer

	outW = os.Stdout
	// если в конфиге указан еще и файл, то через MultiWriter пишем в оба направления
	if cfg.LogFile.Use {
		logFile, err = os.OpenFile(cfg.LogFile.Name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic("open log file: " + err.Error())
		}

		outW = io.MultiWriter(os.Stdout, logFile)
	}

	// функция replaceAttr для маскировки критических данных в логах
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		// KEYS
		if a.Key == slog.TimeKey {
			return slog.String(slog.TimeKey, timeformatter.GlobalTimeFormatter(a.Value.Time()))
		}
		if a.Key == slog.MessageKey {
			return slog.String("message", a.Value.String())
		}
		// VALUES
		// in lowercase!!!
		keyToMask := []string{"password", "pass", "secret", "db connect"}
		exist := false
		keyToFind := strings.ToLower(a.Key)
		for _, key := range keyToMask {
			if key == keyToFind {
				exist = true
			}
		}
		if exist {
			a.Value = slog.StringValue("<<MASKED>>")
		}
		return a
	}

	switch cfg.Env {
	case envLocal:
		// pretty для вывода в local
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}
		logger = slog.New(opts.NewPrettyHandler(os.Stdout))

	case envDev:
		logger = slog.New(slog.NewJSONHandler(outW,
			&slog.HandlerOptions{
				//AddSource:   true,
				Level:       slog.LevelDebug,
				ReplaceAttr: replaceAttr,
			}))

	case envProd:
		logger = slog.New(slog.NewJSONHandler(outW,
			&slog.HandlerOptions{
				Level:       slog.LevelInfo,
				ReplaceAttr: replaceAttr,
			}))

	default:
		panic("unknown env: " + cfg.Env)
	}

	return logger, logFile
}
