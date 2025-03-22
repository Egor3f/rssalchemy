package pwextractor

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/scrapinghub/adblockgoparser"
	"net/url"
	"strings"
)

//go:embed blocklists/easylist.txt
var easyList string

//go:embed blocklists/easyprivacy.txt
var easyPrivacy string

var ruleSet *adblockgoparser.RuleSet

func init() {
	ruleSet = adblockgoparser.CreateRuleSet()
	for _, list := range []string{easyList, easyPrivacy} {
		for _, rec := range strings.Split(list, "\n") {
			rule, err := adblockgoparser.ParseRule(rec)
			if err != nil {
				if !errors.Is(err, adblockgoparser.ErrSkipComment) &&
					!errors.Is(err, adblockgoparser.ErrUnsupportedRule) &&
					!errors.Is(err, adblockgoparser.ErrSkipHTML) &&
					!errors.Is(err, adblockgoparser.ErrEmptyLine) {
					panic(fmt.Sprintf("Adblock rule parse: %v", err))
				}
				continue
			}
			ruleSet.AddRule(rule)
		}
	}
}

func allowAdblock(url *url.URL) bool {
	req := adblockgoparser.Request{
		URL: url,
	}
	allow := ruleSet.Allow(&req)
	if !allow {
		log.Infof("Adblock blocked %s", url.String())
	}
	return allow
}
