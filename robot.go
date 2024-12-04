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

	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/robot-framework-lib/utils"
	"github.com/sirupsen/logrus"
)

// iClient is an interface that defines methods for client-side interactions
type iClient interface {
	// CreatePRComment creates a comment for a pull request in a specified organization and repository
	CreatePRComment(org, repo, number, comment string) (success bool)

	AddPRLabels(org, repo, number string, labels []string) (success bool)
	RemovePRLabels(org, repo, number string, labels []string) (success bool)
	GetPullRequestCommits(org, repo, number string) (result []client.PRCommit, success bool)
	ListPullRequestComments(org, repo, number string) (result []client.PRComment, success bool)
	DeletePRComment(org, repo, commentID string) (success bool)
	CheckCLASignature(urlStr string) (signState string, success bool)
	CheckIfPRCreateEvent(evt *client.GenericEvent) (yes bool)
	CheckIfPRSourceCodeUpdateEvent(evt *client.GenericEvent) (yes bool)
	CheckPermission(org, repo, username string) (pass, success bool)
	GetPullRequestLabels(org, repo, number string) (result []string, success bool)
	MergePullRequest(org, repo, number, mergeMethod string) (success bool)
	CheckIfPRReopenEvent(evt *client.GenericEvent) (yes bool)
	CheckIfPRLabelsUpdateEvent(evt *client.GenericEvent) (yes bool)
}

type robot struct {
	cli iClient
	cnf *configuration
	log *logrus.Entry
}

func (bot *robot) GetConfigmap() config.Configmap {
	return bot.cnf
}

func newRobot(c *configuration, token []byte) *robot {
	logger := framework.NewLogger().WithField("component", component)
	return &robot{cli: client.NewClient(token, logger), cnf: c, log: logger}
}

func (bot *robot) NewConfig() config.Configmap {
	return &configuration{}
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegister) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
	p.RegisterPullRequestCommentHandler(bot.handlePullRequestCommentEvent)
}

func (bot *robot) GetLogger() *logrus.Entry {
	return bot.log
}

// getConfig first checks if the specified organization and repository is available in the provided repoConfig list.
// Returns an error if not found the available repoConfig.
func (bot *robot) getConfig(cnf config.Configmap, org, repo string) (*repoConfig, error) {
	c := cnf.(*configuration)
	if bc := c.get(org, repo); bc != nil {
		return bc, nil
	}

	return nil, errors.New("no config for this repo: " + org + "/" + repo)
}

func (bot *robot) handlePREvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	repoCnf, err := bot.getConfig(cnf, org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the error and returns
	if err != nil {
		logger.WithError(err).Warning()
		return
	}

	if bot.cli.CheckIfPRReopenEvent(evt) || bot.cli.CheckIfPRSourceCodeUpdateEvent(evt) {
		if err := bot.clearLabel(evt, org, repo, number); err != nil {
			logger.WithError(err).Warning()
			return
		}
	}
	if bot.cli.CheckIfPRLabelsUpdateEvent(evt) {
		if err := bot.handleMerge(repoCnf, org, repo, number); err != nil {
			logger.WithError(err).Warning()
			return
		}
	}
}

func (bot *robot) handlePullRequestCommentEvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	comment, commenter, author := utils.GetString(evt.Comment), utils.GetString(evt.Commenter), utils.GetString(evt.Author)
	repoCnf, err := bot.getConfig(cnf, org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the error and returns
	if err != nil {
		logger.WithError(err).Warning()
		return
	}

	if bot.cli.CheckIfPRLabelsUpdateEvent(evt) {
		if err := bot.handleRebase(comment, commenter, org, repo, number); err != nil {
			logger.WithError(err).Warning()
		}

		if err := bot.handledSquash(comment, commenter, org, repo, number); err != nil {
			logger.WithError(err).Warning()
		}

		if err := bot.handleLGTM(repoCnf, comment, commenter, author, org, repo, number); err != nil {
			logger.WithError(err).Warning()
		}

		if err := bot.handleApprove(repoCnf, comment, commenter, author, org, repo, number); err != nil {
			logger.WithError(err).Warning()
		}

		if err := bot.handleCheckPR(evt, repoCnf, org, repo, number); err != nil {
			logger.WithError(err).Warning()
		}
	}
}
