package main

import (
	"regexp"
)

var (
	regAddSquash    = regexp.MustCompile(`(?mi)^/squash\s*$`)
	regRemoveSquash = regexp.MustCompile(`(?mi)^/squash cancel\s*$`)
)

const squashLabel = "merge/squash"

func (bot *robot) handledSquash(comment, commenter, org, repo, number string) error {
	if regAddSquash.MatchString(comment) {
		return bot.addSquash(commenter, org, repo, number)
	}

	if regRemoveSquash.MatchString(comment) {
		return bot.removedSquash(commenter, org, repo, number)
	}

	return nil
}

func (bot *robot) addSquash(commenter, org, repo, number string) error {
	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		label := bot.getPRLabelSet(org, repo, number)
		if _, ok := label[rebaseLabel]; ok {
			bot.cli.CreatePRComment(org, repo, number,
				"Please use **/rebase cancel** to remove **merge/rebase** label, and try **/squash** again")
			return nil
		}
		bot.cli.AddPRLabels(org, repo, number, []string{squashLabel})
	}
	return nil

}

func (bot *robot) removedSquash(commenter, org, repo, number string) error {
	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		bot.cli.RemovePRLabels(org, repo, number, []string{squashLabel})
	}
	return nil
}
