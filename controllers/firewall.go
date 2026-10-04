package controllers

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"plexcorp.tech/scriptable/models"
)

var firewallPortPattern = regexp.MustCompile(`^(any|\d{1,5}(:\d{1,5})?)$`)

func (c *Controller) Firewall(gctx *gin.Context) {
	c.Render("firewall/index", gonja.Context{
		"title":     "Firewall rules",
		"highlight": "firewall",
	}, gctx)
}

func (c *Controller) renderFirewallRules(gctx *gin.Context, successMsg string, actionErr error) {
	vars := gonja.Context{}
	if successMsg != "" {
		vars["successMsg"] = successMsg
	}

	state, fetchErr := models.GetFirewallState()

	if actionErr != nil {
		vars["errorMsg"] = actionErr.Error()
	} else if fetchErr != nil {
		vars["errorMsg"] = fetchErr.Error()
	}

	vars["rules"] = state.Rules
	vars["numRules"] = len(state.Rules)
	vars["firewallInactive"] = fetchErr == nil && !state.Active
	vars["_csrf_token"] = c.SetAndGetCSRFToken(gctx)

	c.RenderWithoutLayout("firewall/_rules", vars, gctx)
}

func (c *Controller) FirewallRules(gctx *gin.Context) {
	c.renderFirewallRules(gctx, "", nil)
}

func (c *Controller) DeleteFirewallRule(gctx *gin.Context) {
	ruleNumber, _ := strconv.ParseInt(gctx.PostForm("rule_number"), 10, 64)
	sessUser := c.GetSessionUser(gctx)

	if ruleNumber == 0 {
		c.renderFirewallRules(gctx, "", errors.New("Bad rule number."))
		return
	}

	if err := models.DeleteFirewallRule(c.GetDB(gctx), ruleNumber, gctx.PostForm("rule"), sessUser.TeamId); err != nil {
		c.renderFirewallRules(gctx, "", err)
		return
	}

	c.renderFirewallRules(gctx, "Successfully deleted the rule.", nil)
}

func isValidFirewallAddress(ip string) bool {
	if strings.EqualFold(ip, "anywhere") || strings.EqualFold(ip, "any") {
		return true
	}

	if _, _, err := net.ParseCIDR(ip); err == nil {
		return true
	}

	return net.ParseIP(ip) != nil
}

func validateFirewallRule(allowBlock, direction, ip, port, protocol string) error {
	if allowBlock != "allow" && allowBlock != "deny" {
		return errors.New("Please choose whether to allow or deny traffic.")
	}

	if direction != "incoming" && direction != "outgoing" {
		return errors.New("Please choose a direction.")
	}

	if protocol != "tcp" && protocol != "udp" {
		return errors.New("Please choose a protocol.")
	}

	if !isValidFirewallAddress(ip) {
		return errors.New("Please enter a valid IP address, CIDR range or Anywhere.")
	}

	if !firewallPortPattern.MatchString(strings.ToLower(port)) {
		return errors.New("Please enter a valid port, port range (8000:8100) or any.")
	}

	return nil
}

func buildUfwRule(allowBlock, direction, ip, port, protocol string) string {
	var rule string
	if direction == "outgoing" {
		rule = fmt.Sprintf("%s out to %s port %s proto %s", allowBlock, ip, port, protocol)
	} else {
		rule = fmt.Sprintf("%s from %s to any port %s proto %s", allowBlock, ip, port, protocol)
	}

	rule = strings.ToLower(rule)

	if strings.Contains(rule, "anywhere") {
		rule = strings.ReplaceAll(rule, "to anywhere port", "")
		rule = strings.ReplaceAll(rule, "from anywhere to any port", "")
		rule = strings.ReplaceAll(rule, "  ", " ")
		rule = strings.ReplaceAll(rule, " proto tcp", "/tcp")
		rule = strings.ReplaceAll(rule, " proto udp", "/udp")
	}

	return strings.TrimSpace(strings.ReplaceAll(rule, "port any proto", "proto"))
}

func (c *Controller) AddFirewallRule(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)

	allowBlock := gctx.PostForm("allow_block")
	direction := gctx.PostForm("direction")
	ip := strings.TrimSpace(gctx.PostForm("ip"))
	port := strings.TrimSpace(gctx.PostForm("port"))
	protocol := gctx.PostForm("protocol")

	if err := validateFirewallRule(allowBlock, direction, ip, port, protocol); err != nil {
		c.renderFirewallRules(gctx, "", err)
		return
	}

	rule := buildUfwRule(allowBlock, direction, ip, port, protocol)

	if err := models.AddFirewallRule(c.GetDB(gctx), rule, sessUser.TeamId); err != nil {
		c.renderFirewallRules(gctx, "", fmt.Errorf("Failed to add rule %q: %s", rule, err))
		return
	}

	c.renderFirewallRules(gctx, "Successfully added the rule.", nil)
}
