package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"plexscriptables.com/scriptables/utils"
)

const FIREWALL_LOG_ENTITY = "firewall"

type FirewallRule struct {
	Number int64
	Text   string
}

type FirewallState struct {
	Active bool
	Rules  []FirewallRule
}

var firewallRuleTypes = []string{"ALLOW IN", "ALLOW OUT", "DENY IN", "DENY OUT"}

func GetFirewallState() (FirewallState, error) {
	state := FirewallState{Rules: []FirewallRule{}}

	output, err := utils.RunCommandAsRoot("ufw", "status", "numbered")
	if err != nil {
		return state, errors.New("failed to read firewall rules: " + strings.TrimSpace(output))
	}

	state.Active = strings.Contains(output, "Status: active")

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "[") {
			state.Rules = append(state.Rules, FirewallRule{
				Number: parseRuleNumber(line),
				Text:   describeRule(line),
			})
		}
	}

	return state, nil
}

func describeRule(line string) string {
	for _, ruleType := range firewallRuleTypes {
		if !strings.Contains(line, ruleType) {
			continue
		}

		parts := strings.SplitN(line, ruleType, 2)
		to := strings.ReplaceAll(strings.TrimSpace(parts[0]), "]", "] TO : ")
		return to + "   FROM : " + strings.TrimSpace(parts[1]) + "  << " + ruleType + " >>"
	}

	return line
}

func parseRuleNumber(rule string) int64 {
	open := strings.Index(rule, "[")
	close := strings.Index(rule, "]")
	if open < 0 || close < open {
		return 0
	}

	number, err := strconv.ParseInt(strings.TrimSpace(rule[open+1:close]), 10, 64)
	if err != nil {
		return 0
	}

	return number
}

func DeleteFirewallRule(db *gorm.DB, ruleNumber int64, rule string, teamId int64) error {
	output, err := utils.RunCommandAsRoot("ufw", "--force", "delete", strconv.FormatInt(ruleNumber, 10))
	if err != nil {
		LogError(db, 0, FIREWALL_LOG_ENTITY, output, fmt.Sprintf("Failed deleting firewall rule number: %d, rule: %s", ruleNumber, rule), teamId)
		return errors.New(strings.TrimSpace(output))
	}

	LogInfo(db, 0, FIREWALL_LOG_ENTITY, output, fmt.Sprintf("Deleted firewall rule number: %d, rule: %s", ruleNumber, rule), teamId)
	return nil
}

func AddFirewallRule(db *gorm.DB, rule string, teamId int64) error {
	output, err := utils.RunCommandAsRoot("ufw", strings.Fields(rule)...)
	if err != nil {
		LogError(db, 0, FIREWALL_LOG_ENTITY, output, fmt.Sprintf("Failed adding firewall rule: %s", rule), teamId)
		return errors.New(strings.TrimSpace(output))
	}

	LogInfo(db, 0, FIREWALL_LOG_ENTITY, output, fmt.Sprintf("Added firewall rule: %s", rule), teamId)
	return nil
}
