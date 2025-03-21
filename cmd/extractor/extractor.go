package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/egor3f/rssalchemy/internal/config"
	dummycookies "github.com/egor3f/rssalchemy/internal/cookiemgr/dummy"
	"github.com/egor3f/rssalchemy/internal/dateparser"
	"github.com/egor3f/rssalchemy/internal/extractors/pwextractor"
	"github.com/egor3f/rssalchemy/internal/limiter/dummy"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"io"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	log.SetLevel(log.DEBUG)
	log.SetHeader(`${time_rfc3339_nano} ${level}`)

	outFile := flag.String("o", "", "Output file name")
	skipOutput := flag.Bool("s", false, "Skip json output; show just logs")
	useProfiler := flag.Bool("p", false, "Use profiler")
	flag.Parse()

	if *useProfiler {
		profFile, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatalf("Profile create file: %v", err)
		}
		defer profFile.Close()
		if err := pprof.StartCPUProfile(profFile); err != nil {
			log.Fatalf("Start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	taskFileName := "task.json"
	if flag.NArg() > 0 {
		taskFileName = flag.Arg(0)
	}

	out := os.Stdout
	if len(*outFile) > 0 {
		var err error
		out, err = os.OpenFile(*outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Panicf("open output file: %v", err)
		}
		//goland:noinspection GoUnhandledErrorResult
		defer out.Close()
	}

	task, err := loadTask(taskFileName)
	if err != nil {
		log.Panicf("load task: %v", err)
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
		CookieManager: dummycookies.New(),
		Limiter:       &dummy.Limiter{},
	})
	if err != nil {
		log.Panicf("create pw extractor: %v", err)
	}
	defer func() {
		if err := pwe.Stop(); err != nil {
			log.Errorf("stop pw extractor: %v", err)
		}
	}()

	start := time.Now()
	result, err := pwe.Extract(task)
	log.Infof("Extract took %v ms", time.Since(start).Milliseconds())
	if err != nil {
		log.Errorf("extract: %v", err)
		scrResult, err := pwe.Screenshot(task)
		if err != nil {
			log.Errorf("screenshot failed: %v", err)
			panic(err)
		}
		err = os.WriteFile("screenshot.png", scrResult.Image, 0600)
		if err != nil {
			log.Errorf("screenshot save failed: %v", err)
		}
		panic(err)
	}

	if !*skipOutput {
		resultStr, err := json.MarshalIndent(result, "", "\t")
		if err != nil {
			log.Panicf("marshal result: %v", err)
		}
		n, err := out.Write(resultStr)
		if err != nil {
			log.Panicf("write output: %v", err)
		}
		log.Infof("Result written (%d bytes)", n)
	}
}

func loadTask(taskFileName string) (models.Task, error) {
	taskFile, err := os.Open(taskFileName)
	if err != nil {
		return models.Task{}, fmt.Errorf("open task file: %w", err)
	}
	defer taskFile.Close()

	fileContents, err := io.ReadAll(taskFile)
	if err != nil {
		return models.Task{}, fmt.Errorf("read file: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal(fileContents, &task); err != nil {
		return models.Task{}, fmt.Errorf("unmarshal task: %w", err)
	}

	return task, err
}
