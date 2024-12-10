package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	msgPRConflicts        = "PR conflicts to the target branch."
	msgMissingLabels      = "PR does not have these lables: %s"
	msgInvalidLabels      = "PR should remove these labels: %s"
	msgNotEnoughLGTMLabel = "PR needs %d lgtm labels and now gets %d"
	ActionAddLabel        = "add label"
)

type labelLog struct {
	label string
	who   string
	t     time.Time
}

func (bot *robot) handleMerge(configmap *repoConfig, org, repo, number string) error {
	labels := bot.getPRLabelSet(org, repo, number)
	ops, ok := bot.cli.ListPullRequestOperationLogs(org, repo, number)
	if !ok {
		return fmt.Errorf("failed to list pull request operation logs")
	}
	if err := checkLabelsLegal(configmap, ops, labels); err != nil {
		return err
	}
	if err := isLabelMatched(configmap, labels); err != nil {
		return err
	}

	methodOfMerge := bot.genMergeMethod(org, repo, number)
	if ok := bot.cli.MergePullRequest(org, repo, number, methodOfMerge); !ok {
		return fmt.Errorf("failed to merge pull request")
	}
	return nil
}

func isLabelMatched(configmap *repoConfig, labels sets.Set[string]) error {

	for _, l := range configmap.LabelsNotAllowMerge {
		if labels.Has(l) {
			return fmt.Errorf(msgInvalidLabels, l)
		}
	}

	needs := sets.New[string](approvedLabel)
	needs.Insert(configmap.LabelsForMerge...)

	if ln := configmap.LgtmCountsRequired; ln == 1 {
		needs.Insert(lgtmLabel)
	} else {
		v := getLGTMLabelsOnPR(labels)
		if n := uint(len(v)); n < ln {
			return fmt.Errorf(msgNotEnoughLGTMLabel, ln, n)
		}
	}

	if v := needs.Difference(labels); v.Len() > 0 {
		vl := v.UnsortedList()
		var vlp []string
		for _, i := range vl {
			vlp = append(vlp, fmt.Sprintf("***%s***", i))
		}
		return fmt.Errorf(msgMissingLabels, strings.Join(vlp, ", "))
	}

	return nil
}

func checkLabelsLegal(configmap *repoConfig, ops []client.PullRequestOperationLog, labels sets.Set[string]) error {
	needs := sets.New[string](approvedLabel)
	needs.Insert(configmap.LabelsForMerge...)
	if ln := configmap.LgtmCountsRequired; ln == 1 {
		needs.Insert(lgtmLabel)
	} else {
		v := getLGTMLabelsOnPR(labels)
		if n := uint(len(v)); n < ln {
			return fmt.Errorf(msgNotEnoughLGTMLabel, ln, n)
		}
	}
	legalOperator := configmap.LegalOperator
	reason := make([]string, 0, len(labels))
	for label := range labels {
		if ok := needs.Has(label); ok || strings.HasPrefix(label, lgtmLabel) {
			if s := isLabelLegal(ops, label, legalOperator); s != "" {
				reason = append(reason, s)
			}
		}
	}
	if n := len(reason); n > 0 {
		s := "label is "
		if n > 1 {
			s = "labels are "
		}
		return fmt.Errorf("**The following %s not ready**.\n\n%s", s, strings.Join(reason, "\n\n"))
	}
	return nil
}

func isLabelLegal(ops []client.PullRequestOperationLog, label string, legalOperator string) string {
	labelLog, ok := getLatestLog(ops, label)
	if !ok {
		logrus.Errorf("=== label: %s", label)
		return fmt.Sprintf("The corresponding operation log is missing. you should delete "+
			"the label ** %s ** and add it again by correct way", label)
	}
	if labelLog.who != legalOperator {
		return fmt.Sprintf("%s You can't add %s by yourself, please contact the maintainers", labelLog.who, labelLog.label)
	}
	return ""
}

func getLatestLog(ops []client.PullRequestOperationLog, label string) (labelLog, bool) {
	var t time.Time
	index := -1

	for i := range ops {
		op := &ops[i]
		logrus.Infof("===>op: %+v", op)
		if !strings.HasPrefix(op.Content, ActionAddLabel) || !strings.Contains(op.Content, label) {
			continue
		}

		if index < 0 || op.CreatedAt.After(t) {
			t = op.CreatedAt
			index = i
		}
	}

	if index >= 0 {
		if user := ops[index].UserName; user != "" {
			logrus.Infof("<===user add: %s %s", ops[index].Content, user)
			return labelLog{
				label: label,
				t:     t,
				who:   user,
			}, true
		}
	}
	return labelLog{}, false
}
