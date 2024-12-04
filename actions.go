package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/utils"
	"k8s.io/apimachinery/pkg/util/sets"
)

var regCheckPr = regexp.MustCompile(`(?mi)^/check-pr\s*$`)

func (bot *robot) clearLabel(evt *client.GenericEvent, org, repo, number string) error {

	labels := bot.getPRLabelSet(org, repo, number)
	v := getLGTMLabelsOnPR(labels)

	if labels.Has(approvedLabel) {
		v = append(v, approvedLabel)
	}

	if len(v) > 0 {

		if ok := bot.cli.RemovePRLabels(org, repo, number, v); !ok {
			return nil
		}

		var noteComment string
		if bot.cli.CheckIfPRSourceCodeUpdateEvent(evt) {
			noteComment = commentClearLabelCaseByReopenPR
		}

		if bot.cli.CheckIfPRSourceCodeUpdateEvent(evt) {
			noteComment = commentClearLabelCaseByPRUpdate
		}

		bot.cli.CreatePRComment(
			org, repo, number,
			fmt.Sprintf(noteComment, strings.Join(v, ", ")),
		)
	}

	return nil
}
func (bot *robot) checkCommenterPermission(org, repo, author, commenter string, fn func()) (pass bool) {
	if author == commenter {
		return true
	}
	pass, success := bot.cli.CheckPermission(org, repo, commenter)
	bot.log.Infof("request success: %t, the %s has permission to the repo[%s/%s]: %t", success, commenter, org, repo, pass)

	if success && !pass {
		fn()
	}
	return pass && success
}

func (bot *robot) getPRLabelSet(org, repo, number string) sets.Set[string] {
	res := sets.New[string]()

	labels, ok := bot.cli.GetPullRequestLabels(org, repo, number)
	if !ok {
		return res
	}

	for _, v := range labels {
		res.Insert(v)
	}

	if res.Has("") {
		res.Delete("")
	}

	return res
}

func (bot *robot) genMergeMethod(org, repo, number string) string {
	mergeMethod := "merge"

	prLabels := bot.getPRLabelSet(org, repo, number)

	for p := range prLabels {
		if strings.HasPrefix(p, "merge/") {
			if strings.Split(p, "/")[1] == "squash" {
				return "squash"
			}

			return strings.Split(p, "/")[1]
		}
	}

	return mergeMethod
}

func (bot *robot) handleCheckPR(evt *client.GenericEvent, configmap *repoConfig, org, repo, number string) error {
	comment := utils.GetString(evt.Comment)
	if !regCheckPr.MatchString(comment) {
		return nil
	}

	return bot.handleMerge(configmap, org, repo, number)
}
