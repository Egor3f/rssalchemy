package pwextractor

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/patriciy/adblock/adblock"
	"github.com/playwright-community/playwright-go"
	"net/url"
	"sync"
)

//go:embed blocklists/easylist.txt
var easyList []byte

//go:embed blocklists/easyprivacy.txt
var easyPrivacy []byte

var matcher *adblock.RuleMatcher
var matcherMu sync.Mutex

func init() {
	allBuf := bytes.Buffer{}
	allBuf.Write(easyList)
	//allBuf.Write(easyPrivacy)
	r := bytes.NewReader(allBuf.Bytes())
	rules, err := adblock.ParseRules(r)
	if err != nil {
		panic(fmt.Sprintf("Parse rules: %v", err))
	}

	matcher = adblock.NewMatcher()
	for i, rule := range rules {
		if err := matcher.AddRule(rule, i); err != nil {
			panic(fmt.Sprintf("Add rule: %v", err))
		}
	}
}

func allowAdblock(url *url.URL) bool {
	matcherMu.Lock()
	defer matcherMu.Unlock()
	req := adblock.Request{
		URL:          url.String(),
		Domain:       url.Host,
		GenericBlock: playwright.Bool(false),
	}
	deny, ruleId, err := matcher.Match(&req)
	if err != nil {
		log.Errorf("Adblock error: %v (url %s)", err, url.String())
		return true
	}
	if deny {
		log.Infof("Adblock blocked %s (ruleId %d)", url.String(), ruleId)
	}
	return !deny
}
