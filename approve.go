package main

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

const approvedLabel = "approved"

var (
	regAddApprove    = regexp.MustCompile(`(?mi)^/approve\s*$`)
	regRemoveApprove = regexp.MustCompile(`(?mi)^/approve cancel\s*$`)
)

func (bot *robot) handleApprove(configmap *repoConfig, comment, commenter, author, org, repo, number string) error {
	if regAddApprove.MatchString(comment) {
		return bot.AddApprove(commenter, author, org, repo, number, configmap.LgtmCountsRequired)
	}

	if regRemoveApprove.MatchString(comment) {
		return bot.removeApprove(commenter, author, org, repo, number, configmap.LgtmCountsRequired)
	}

	return nil
}

func (bot *robot) AddApprove(commenter, author, org, repo, number string, lgtmCounts uint) error {
	logrus.Infof("AddApprove, commenter: %s, author: %s, org: %s, repo: %s, number: %s", commenter, author, org, repo, number)
	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		if ok := bot.cli.AddPRLabels(org, repo, number, []string{approvedLabel}); !ok {
			return fmt.Errorf("failed to add label on pull request")
		}
		if ok := bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentAddLabel, approvedLabel, commenter)); !ok {
			return fmt.Errorf("failed to comment on pull request")
		}
	} else if !pass {
		bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentNoPermissionForLgtmLabel, commenter))
	} else {
		return fmt.Errorf("failed to add label on pull request")
	}
	return nil
}

func (bot *robot) removeApprove(commenter, author, org, repo, number string, lgtmCounts uint) error {
	logrus.Infof("removeApprove, commenter: %s, author: %s, org: %s, repo: %s, number: %s", commenter, author, org, repo, number)

	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		bot.cli.RemovePRLabels(org, repo, number, []string{approvedLabel})
		bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentRemovedLabel, approvedLabel, commenter))
	} else if !pass {
		bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentNoPermissionForLabel, commenter, "remove", approvedLabel))
	} else {
		return fmt.Errorf("failed to remove label on pull request")
	}

	return nil
}
