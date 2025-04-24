package pwextractor

import (
	_ "embed"
	"fmt"
	"github.com/AdguardTeam/urlfilter"
	"github.com/AdguardTeam/urlfilter/filterlist"
	"github.com/AdguardTeam/urlfilter/rules"
	"github.com/labstack/gommon/log"
	"net/url"
)

//go:embed blocklists/easylist.txt
var easyList string

//go:embed blocklists/easyprivacy.txt
var easyPrivacy string

var ruleLists = []string{
	easyList,
	easyPrivacy,
}

var engine *urlfilter.Engine

func init() {
	lists := make([]filterlist.RuleList, len(ruleLists))
	for i, rulesStr := range ruleLists {
		lists[i] = &filterlist.StringRuleList{
			RulesText:      rulesStr,
			ID:             i,
			IgnoreCosmetic: true,
		}
	}
	storage, err := filterlist.NewRuleStorage(lists)
	if err != nil {
		panic(fmt.Sprintf("initialize adblock: NewRuleStorage: %v", err))
	}
	engine = urlfilter.NewEngine(storage)
}

func allowAdblock(url *url.URL, sourceUrl *url.URL) bool {
	req := rules.NewRequest(url.String(), sourceUrl.String(), rules.TypeOther)
	res := engine.MatchRequest(req)
	rule := res.GetBasicResult()
	allow := rule == nil || rule.Whitelist
	if !allow {
		log.Infof("Adblock blocked %s from %s by rule %s", url, sourceUrl, rule.String())
	}
	return allow
}
