package main

import (
	"encoding/json"
	"github.com/egor3f/rssalchemy/internal/config"
	"github.com/egor3f/rssalchemy/internal/dateparser"
	"github.com/egor3f/rssalchemy/internal/extractors/pwextractor"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/yassinebenaid/godump"
	"io"
	"os"
	"time"
)

func main() {
	log.SetLevel(log.DEBUG)
	log.SetHeader(`${level}`)

	taskFileName := "task.json"
	if len(os.Args) > 1 {
		taskFileName = os.Args[1]
	}

	taskFile, err := os.Open(taskFileName)
	if err != nil {
		log.Panicf("open file: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer taskFile.Close()
	fileContents, err := io.ReadAll(taskFile)
	if err != nil {
		log.Panicf("read file: %v", err)
	}
	var task models.Task
	if err := json.Unmarshal(fileContents, &task); err != nil {
		log.Panicf("unmarshal task: %v", err)
	}

	cfg, err := config.Read()
	if err != nil {
		log.Panicf("read config: %v", err)
	}

	pwe, err := pwextractor.New(pwextractor.Config{
		Proxy: cfg.Proxy,
		DateParser: &dateparser.DateParser{
			CurrentTimeFunc: func() time.Time {
				return time.Date(2025, 01, 10, 10, 00, 00, 00, time.UTC)
			},
		},
	})
	if err != nil {
		log.Panicf("create pw extractor: %v", err)
	}
	defer func() {
		if err := pwe.Stop(); err != nil {
			log.Errorf("stop pw extractor: %v", err)
		}
	}()

	result, err := pwe.Extract(task)
	if err != nil {
		log.Errorf("extract: %v", err)
		scrResult, err := pwe.Screenshot(task)
		if err != nil {
			log.Errorf("screenshot failed: %v", err)
			return
		}
		err = os.WriteFile("screenshot.png", scrResult.Image, 0600)
		if err != nil {
			log.Errorf("screenshot save failed: %v", err)
		}
		return
	}

	dumper := godump.Dumper{Theme: godump.Theme{
		String:        godump.RGB{117, 54, 217},
		Quotes:        godump.RGB{143, 41, 0},
		Bool:          godump.RGB{6, 168, 199},
		Number:        godump.RGB{245, 77, 13},
		Types:         godump.RGB{255, 105, 56},
		Address:       godump.RGB{50, 162, 255},
		PointerTag:    godump.RGB{145, 145, 145},
		Nil:           godump.RGB{36, 198, 229},
		Func:          godump.RGB{95, 165, 35},
		Fields:        godump.RGB{66, 79, 61},
		Chan:          godump.RGB{60, 101, 179},
		UnsafePointer: godump.RGB{166, 62, 75},
		Braces:        godump.RGB{70, 169, 169},
	}}
	_ = dumper.Println(result)
}
