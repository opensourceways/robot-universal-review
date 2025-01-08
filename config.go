// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"errors"
	"fmt"

	"github.com/opensourceways/server-common-lib/config"
)

// configuration holds a list of repoConfig configurations.
type configuration struct {
	ConfigItems          []repoConfig `json:"config_items,omitempty"`
	UserMarkFormat       string       `json:"user_mark_format,omitempty"`
	PlaceholderCommenter string       `json:"placeholder_commenter"`
	// Sig information url.
	SigInfoURL string `json:"sig_info_url" required:"true"`
	// Community name used as a request parameter to getRepoConfig sig information.
	CommunityName string `json:"community_name" required:"true"`
}

// Validate to check the configmap data's validation, returns an error if invalid
func (c *configuration) Validate() error {
	if c == nil {
		return errors.New("configuration is nil")
	}

	// Validate each repo configuration
	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

// get retrieves a repoConfig for a given organization and repository.
// Returns the repoConfig if found, otherwise returns nil.
func (c *configuration) get(org, repo string) *repoConfig {
	if c == nil || len(c.ConfigItems) == 0 {
		return nil
	}

	for i := range c.ConfigItems {
		ok, _ := c.ConfigItems[i].RepoFilter.CanApply(org, org+"/"+repo)
		if ok {
			return &c.ConfigItems[i]
		}
	}

	return nil
}

// repoConfig is a configuration struct for a organization and repository.
// It includes a RepoFilter and a boolean value indicating if an issue can be closed only when its linking PR exists.
type repoConfig struct {
	// RepoFilter is used to filter repositories.
	config.RepoFilter
	// LegalOperator means who can add or remove labels legally
	LegalOperator string `json:"legal_operator,omitempty"`

	// LgtmCountsRequired specifies the number of lgtm label which will be need for the pr.
	// When it is greater than 1, the lgtm label is composed of 'lgtm-login'.
	// The default value is 1 which means the lgtm label is itself.
	LgtmCountsRequired uint `json:"lgtm_counts_required,omitempty"`

	// LabelsForMerge specifies the labels except approved and lgtm relevant labels
	// that must be available to merge pr
	LabelsForMerge []string `json:"labels_for_merge,omitempty"`

	// LabelsNotAllowMerge means that if pull request has these labels, it can not been merged
	// even all conditions are met
	LabelsNotAllowMerge []string `json:"labels_not_allow_merge,omitempty"`

	// MergeMethod is the method to merge PR.
	// The default method of merge. Valid options are squash and merge.
	MergeMethod string `json:"merge_method,omitempty"`
}

type freezeFile struct {
	Owner  string `json:"owner" required:"true"`
	Repo   string `json:"repo" required:"true"`
	Branch string `json:"branch" required:"true"`
	Path   string `json:"path" required:"true"`
}

type branchKeeper struct {
	Owner  string `json:"owner" required:"true"`
	Repo   string `json:"repo" required:"true"`
	Branch string `json:"branch" required:"true"`
}

func (b branchKeeper) validate() error {
	if b.Owner == "" {
		return fmt.Errorf("missing owner of branch keeper")
	}

	if b.Repo == "" {
		return fmt.Errorf("missing repo of branch keeper")
	}

	if b.Branch == "" {
		return fmt.Errorf("missing branch of branch keeper")
	}

	return nil
}

func (f freezeFile) validate() error {
	if f.Owner == "" {
		return fmt.Errorf("missing owner of freeze file")
	}

	if f.Repo == "" {
		return fmt.Errorf("missing repo of freeze file")
	}

	if f.Branch == "" {
		return fmt.Errorf("missing branch of freeze file")
	}

	if f.Path == "" {
		return fmt.Errorf("missing path of freeze file")
	}

	return nil
}

// validate to check the repoConfig data's validation, returns an error if invalid
func (c *repoConfig) validate() error {
	// If the bot is not configured to monitor any repositories, return an error.
	if len(c.Repos) == 0 {
		return errors.New("the repositories configuration can not be empty")
	}

	return c.RepoFilter.Validate()
}

type litePRCommiter struct {
	// Email is the one of committer in a commit when a PR is lite
	Email string `json:"email" required:"true"`

	// Name is the one of committer in a commit when a PR is lite
	Name string `json:"name" required:"true"`
}
