package products

import (
	"encoding/json"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type AttributeOption struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func expandMatrix(matrix []AttributeOption) []map[string]string {
	combinations := []map[string]string{{}}
	for _, option := range matrix {
		if len(option.Values) == 0 {
			continue
		}
		next := make([]map[string]string, 0, len(combinations)*len(option.Values))
		for _, combo := range combinations {
			for _, value := range option.Values {
				copied := make(map[string]string, len(combo)+1)
				for k, v := range combo {
					copied[k] = v
				}
				copied[option.Name] = value
				next = append(next, copied)
			}
		}
		combinations = next
	}
	return combinations
}

func marshalAttributes(combination map[string]string) (json.RawMessage, error) {
	return json.Marshal(combination)
}

func selectBestRule(
	rules []*PriceRule,
	variant *ItemVariant,
	template *Item,
	qty amount.Amount,
	date time.Time,
	templateOfProduct func(itemID uint64) (uint64, bool, error),
) (*PriceRule, error) {
	var best *PriceRule
	bestScore := -1
	for _, rule := range rules {
		applies, err := ruleApplies(rule, variant, template, qty, date, templateOfProduct)
		if err != nil {
			return nil, err
		}
		if !applies {
			continue
		}
		score := ruleScore(rule)
		if score > bestScore {
			best = rule
			bestScore = score
		}
	}
	return best, nil
}

func ruleApplies(
	rule *PriceRule,
	variant *ItemVariant,
	template *Item,
	qty amount.Amount,
	date time.Time,
	templateOfProduct func(itemID uint64) (uint64, bool, error),
) (bool, error) {
	if !ruleInDate(rule, date) {
		return false, nil
	}
	if amount.FromFloat64(rule.MinQty).GreaterThan(qty) {
		return false, nil
	}
	switch rule.AppliesTo {
	case AppliesToAll:
		return true, nil
	case AppliesToCategory:
		return template.CategoryID != nil && rule.CategoryID != nil && *template.CategoryID == *rule.CategoryID, nil
	case AppliesToVariant:
		return rule.ItemID != nil && *rule.ItemID == variant.ID, nil
	case AppliesToProduct:
		if rule.ItemID == nil {
			return false, nil
		}
		ruleItemID, found, err := templateOfProduct(*rule.ItemID)
		if err != nil || !found {
			return false, err
		}
		return ruleItemID == template.ID, nil
	default:
		return false, nil
	}
}

func ruleInDate(rule *PriceRule, date time.Time) bool {
	return inDateWindow(date, rule.DateStart, rule.DateEnd)
}

func inDateWindow(date time.Time, from, to *time.Time) bool {
	if from != nil && date.Before(from.Truncate(24*time.Hour)) {
		return false
	}
	if to != nil && !date.Before(to.Truncate(24*time.Hour).Add(24*time.Hour)) {
		return false
	}
	return true
}

func ruleScore(rule *PriceRule) int {
	switch rule.AppliesTo {
	case AppliesToVariant:
		return 4
	case AppliesToProduct:
		return 3
	case AppliesToCategory:
		return 2
	default:
		return 1
	}
}
