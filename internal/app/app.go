package app

import (
	"time"

	"github.com/hrvadl/go-weekly/internal/crawler"
	"github.com/hrvadl/go-weekly/internal/tg"
	"github.com/hrvadl/go-weekly/internal/tg/formatter"
	"github.com/hrvadl/go-weekly/internal/translator"
	"github.com/hrvadl/go-weekly/pkg/logger"
)

type Config struct {
	TranslateBatchRequests int
	TranslateRetries       int
	TranslateTimeout       time.Duration
	TranslateInterval      time.Duration

	ArticlesRetries int
	ArticlesTimeout time.Duration

	TgToken  string
	TgChatID string
}

func New(cfg Config) *GoWeekly {
	return &GoWeekly{cfg}
}

type GoWeekly struct {
	cfg Config
}

func (o GoWeekly) TranslateAndSend() {
	start := time.Now()
	crawler := crawler.New(o.cfg.ArticlesTimeout, o.cfg.ArticlesRetries)
	bot := tg.NewBot(o.cfg.TgToken, o.cfg.TgChatID)
	formatter := formatter.NewMarkdown()
	translator := translator.NewLingvaClient(&translator.Config{
		Timeout:         o.cfg.TranslateTimeout,
		Retries:         o.cfg.TranslateRetries,
		RetriesInterval: o.cfg.TranslateInterval / 2,
		BatchRequests:   o.cfg.TranslateBatchRequests,
		BatchInterval:   o.cfg.TranslateInterval,
	})

	articles, err := crawler.ParseArticles()
	if err != nil {
		logger.Fatalf("Cannot parse articles: %v\n", err)
	}

	logger.Infof(
		"Successfully parsed articles in %v: %v\n",
		time.Since(start).String(),
		articles,
	)

	if err := translator.TranslateArticles(articles); err != nil {
		logger.Fatalf("Failed to translate articles: %v\n", err)
	}

	logger.Infof(
		"Successfully translated articles in %v: %v\n",
		time.Since(start).String(),
		articles,
	)

	bot.SendMessagesThroughoutWeek(formatter.FormatArticles(articles))
	logger.Info("Finished sending all the weekly articles")
}
