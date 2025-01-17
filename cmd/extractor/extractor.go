package main

import (
	"github.com/egor3f/rssalchemy/internal/extractors/pwextractor"
	"github.com/egor3f/rssalchemy/internal/models"
	"github.com/labstack/gommon/log"
	"github.com/yassinebenaid/godump"
)

func main() {
	log.SetLevel(log.DEBUG)
	log.SetHeader(`${level}`)

	task := models.Task{
		URL:                 "https://vombat.su",
		SelectorPost:        "div.post-body",
		SelectorTitle:       "h1 a",
		SelectorLink:        "h1 a",
		SelectorDescription: "div.post-content-block p",
		SelectorAuthor:      "a:has(> span.post-author)",
		SelectorCreated:     "div:nth-of-type(1) > div:nth-of-type(1) > div:nth-of-type(1) > div:nth-of-type(2)",
		SelectorContent:     "div.post-content-block",
		SelectorEnclosure:   "article img.object-contain",
	}

	pwe, err := pwextractor.New()
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
		log.Panicf("extract: %v", err)
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
