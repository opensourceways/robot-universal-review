package main

import (
	"regexp"

	"github.com/sirupsen/logrus"
)

var (
	regAddRebase    = regexp.MustCompile(`(?mi)^/rebase\s*$`)
	regRemoveRebase = regexp.MustCompile(`(?mi)^/rebase cancel\s*$`)
)

const rebaseLabel = "merge/rebase"

func (bot *robot) handleRebase(comment, commenter, org, repo, number string) error {
	logrus.Infof("handleRebase, comment: %s, commenter: %s, org: %s, repo: %s, number: %s", comment, commenter, org, repo, number)
	if regAddRebase.MatchString(comment) {
		return bot.addRebase(commenter, org, repo, number)
	}

	if regRemoveRebase.MatchString(comment) {
		return bot.removeRebase(commenter, org, repo, number)
	}

	return nil
}

func (bot *robot) addRebase(commenter, org, repo, number string) error {
	logrus.Infof("addRebase, commenter: %s, org: %s, repo: %s, number: %s", commenter, org, repo, number)
	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		label := bot.getPRLabelSet(org, repo, number)
		if _, ok := label["merge/squash"]; ok {
			bot.cli.CreatePRComment(org, repo, number,
				"Please use **/squash cancel** to remove **merge/squash** label, and try **/rebase** again")
			return nil
		}
		bot.cli.AddPRLabels(org, repo, number, []string{rebaseLabel})
	}
	return nil

}

func (bot *robot) removeRebase(commenter, org, repo, number string) error {
	logrus.Infof("removeRebase, commenter: %s, org: %s, repo: %s, number: %s", commenter, org, repo, number)
	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		bot.cli.RemovePRLabels(org, repo, number, []string{rebaseLabel})
	}
	return nil
}
