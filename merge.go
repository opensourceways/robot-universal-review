package main

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	msgPRConflicts        = "PR conflicts to the target branch."
	msgMissingLabels      = "PR does not have these lables: %s"
	msgInvalidLabels      = "PR should remove these labels: %s"
	msgNotEnoughLGTMLabel = "PR needs %d lgtm labels and now gets %d"
)

func (bot *robot) handleMerge(configmap *repoConfig, org, repo, number string) error {
	labels := bot.getPRLabelSet(org, repo, number)
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
