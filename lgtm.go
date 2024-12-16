package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// the gitee platform limits the maximum length of label to 20.
	labelLenLimit = 20
	lgtmLabel     = "lgtm"

	commentAddLGTMBySelf            = "***lgtm*** can not be added in your self-own pull request. :astonished:"
	commentClearLabelCaseByPRUpdate = `New code changes of pr are detected and remove these labels ***%s***. :flushed: `
	commentClearLabelCaseByReopenPR = `When PR is reopened, remove these labels ***%s***. :flushed: `
	commentNoPermissionForLgtmLabel = `Thanks for your review, ***%s***, your opinion is very important to us.:wave:
The maintainers will consider your advice carefully.`
	commentNoPermissionForLabel = `
***@%s*** has no permission to %s ***%s*** label in this pull request. :astonished:
Please contact to the collaborators in this repository.`
	commentAddLabel = `***%s*** was added to this pull request by: ***%s***. :wave: 
**NOTE:** If this pull request is not merged while all conditions are met, comment "/check-pr" to try again. :smile: `
	commentRemovedLabel = `***%s*** was removed in this pull request by: ***%s***. :flushed: `
)

var (
	regAddLgtm    = regexp.MustCompile(`(?mi)^/lgtm\s*$`)
	regRemoveLgtm = regexp.MustCompile(`(?mi)^/lgtm cancel\s*$`)
)

func (bot *robot) handleLGTM(configmap *repoConfig, comment, commenter, author, org, repo, number string) error {
	if regAddLgtm.MatchString(comment) {
		return bot.addLGTM(commenter, author, org, repo, number, configmap.LgtmCountsRequired)
	}

	if regRemoveLgtm.MatchString(comment) {
		return bot.removeLGTM(commenter, author, org, repo, number, configmap.LgtmCountsRequired)
	}

	return nil
}

func (bot *robot) addLGTM(commenter, author, org, repo, number string, lgtmCounts uint) error {
	logrus.Infof("addLGTM, commenter: %s, author: %s, org: %s, repo: %s, number: %s", commenter, author, org, repo, number)
	if author == commenter {
		if ok := bot.cli.CreatePRComment(org, repo, number, commentAddLGTMBySelf); !ok {
			return fmt.Errorf("failed to comment on pull request")
		}
		return nil
	}
	if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
		label := genLGTMLabel(commenter, lgtmCounts)

		if ok := bot.cli.AddPRLabels(org, repo, number, []string{label}); !ok {
			return fmt.Errorf("failed to add label on pull request")
		}
		if ok := bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentAddLabel, label, commenter)); !ok {
			return fmt.Errorf("failed to comment on pull request")
		}
	} else {
		bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentNoPermissionForLgtmLabel, commenter))
	}
	return nil

}

func (bot *robot) removeLGTM(commenter, author, org, repo, number string, lgtmCounts uint) error {
	logrus.Infof("removeLGTM, commenter: %s, author: %s, org: %s, repo: %s, number: %s", commenter, author, org, repo, number)
	if author == commenter {
		bot.cli.RemovePRLabels(org, repo, number, getLGTMLabelsOnPR(bot.getPRLabelSet(org, repo, number)))
	} else {
		if pass, ok := bot.cli.CheckPermission(org, repo, commenter); pass && ok {
			label := genLGTMLabel(commenter, lgtmCounts)
			bot.cli.RemovePRLabels(org, repo, number, []string{label})
			bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentRemovedLabel, label, commenter))
		} else {
			bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentNoPermissionForLabel, commenter, "remove", lgtmLabel))
		}

	}
	return nil
}

func genLGTMLabel(commenter string, lgtmCount uint) string {
	if lgtmCount <= 1 {
		return lgtmLabel
	}

	l := fmt.Sprintf("%s-%s", lgtmLabel, strings.ToLower(commenter))
	if len(l) > labelLenLimit {
		return l[:labelLenLimit]
	}

	return l
}

func getLGTMLabelsOnPR(labels sets.Set[string]) []string {
	var r []string

	for l := range labels {
		if strings.HasPrefix(l, lgtmLabel) {
			r = append(r, l)
		}
	}

	return r
}
