package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
	isCreatePRComment              bool
	isDeletePRComment              bool
	isAddPRLabels                  bool
	isRemovePRLabels               bool
	isCheckPermission              bool
	hasPermission                  bool
	isGetPullRequestCommits        bool
	isCheckIfPRReopen              bool
	isListPullRequestComments      bool
	isGetPullRequestLabels         bool
	isCheckIfPRCreate              bool
	isCheckIfPRSourceCode          bool
	isCheckIfPRLabels              bool
	isCheckIfPRSourceCodeUpdate    bool
	isMergePullRequest             bool
	isCheckCLASignature            bool
	isCheckIfPRReopenEvent         bool
	isCheckIfPRLabelsUpdateEvent   bool
	isListPullRequestOperationLogs bool
	method                         string
	pullRequestOperationLogs       []client.PullRequestOperationLog
	commits                        []client.PRCommit
	prComments                     []client.PRComment
	labels                         []string
	CLAState                       string
}

func (m *mockClient) AddPRLabels(org, repo, number string, labels []string) (success bool) {
	m.method = "AddPRLabels"
	return m.isAddPRLabels
}

func (m *mockClient) RemovePRLabels(org, repo, number string, labels []string) (success bool) {
	m.method = "RemovePRLabels"
	return m.isRemovePRLabels
}

func (m *mockClient) ListPullRequestComments(org, repo, number string) (result []client.PRComment, success bool) {
	m.method = "ListPullRequestComments"
	return m.prComments, m.isListPullRequestComments
}

func (m *mockClient) CheckIfPRCreateEvent(evt *client.GenericEvent) (yes bool) {
	m.method = "CheckIfPRCreateEvent"
	return m.isCheckIfPRCreate
}

func (m *mockClient) CheckIfPRSourceCodeUpdateEvent(evt *client.GenericEvent) (yes bool) {
	m.method = "CheckIfPRSourceCodeUpdateEvent"
	return m.isCheckIfPRSourceCode
}

func (m *mockClient) CheckPermission(org, repo, username string) (pass, success bool) {
	m.method = "CheckPermission"
	return m.hasPermission, m.isCheckPermission
}

func (m *mockClient) GetPullRequestLabels(org, repo, number string) (result []string, success bool) {
	m.method = "GetPullRequestLabels"
	return m.labels, m.isGetPullRequestLabels
}

func (m *mockClient) MergePullRequest(org, repo, number, mergeMethod string) (success bool) {
	m.method = "MergePullRequest"
	return m.isMergePullRequest
}

func (m *mockClient) CheckIfPRReopenEvent(evt *client.GenericEvent) (yes bool) {
	m.method = "CheckIfPRReopenEvent"
	return m.isCheckIfPRReopen
}

func (m *mockClient) CheckIfPRLabelsUpdateEvent(evt *client.GenericEvent) (yes bool) {
	m.method = "CheckIfPRLabelsUpdateEvent"
	return m.isCheckIfPRLabels
}

func (m *mockClient) ListPullRequestOperationLogs(org, repo, number string) (result []client.PullRequestOperationLog, success bool) {
	m.method = "ListPullRequestOperationLogs"
	return m.pullRequestOperationLogs, m.isListPullRequestOperationLogs
}

func (m *mockClient) CreatePRComment(org, repo, number, comment string) bool {
	m.method = "CreatePRComment"
	return m.isCreatePRComment
}

const (
	org       = "org1"
	repo      = "repo1"
	number    = "1"
	commenter = "commenter1"
	author    = "author1"
)

func TestRebase(t *testing.T) {
	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{}}
	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)

	// rebase success
	case1 := "AddPRLabels"
	cli.method = ""
	comment := "/rebase"
	cli.isCheckPermission = true
	cli.hasPermission = true
	cli.labels = []string{}
	err := bot.handleRebase(comment, commenter, org, repo, number)
	execMethod1 := cli.method
	assert.Equal(t, case1, execMethod1)
	assert.Equal(t, err, nil)
	// rebase cancel success
	comment = "/rebase cancel"
	case2 := "RemovePRLabels"
	cli.labels = []string{rebaseLabel}
	err = bot.handleRebase(comment, commenter, org, repo, number)
	execMethod2 := cli.method
	assert.Equal(t, case2, execMethod2)
	assert.Equal(t, err, nil)

	// rebase fail
	case3 := "CreatePRComment"
	comment = "/rebase"
	cli.labels = []string{squashLabel}
	cli.isGetPullRequestLabels = true
	err = bot.handleRebase(comment, commenter, org, repo, number)
	execMethod3 := cli.method
	assert.Equal(t, case3, execMethod3)
	assert.Equal(t, err, nil)
	// rebase cancel fail
	case3 = "RemovePRLabels"
	comment = "/rebase cancel"
	cli.labels = []string{squashLabel}
	err = bot.handleRebase(comment, commenter, org, repo, number)
	execMethod4 := cli.method
	assert.Equal(t, case3, execMethod4)
	assert.Equal(t, err, nil)
}

func TestSquash(t *testing.T) {
	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{}}
	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)

	// squash success
	case5 := "AddPRLabels"
	comment := "/squash"
	cli.isCheckPermission = true
	cli.hasPermission = true
	cli.labels = []string{}
	err := bot.handledSquash(comment, commenter, org, repo, number)
	execMethod5 := cli.method
	assert.Equal(t, case5, execMethod5)
	assert.Equal(t, err, nil)
	// squash cancel success
	comment = "/squash cancel"
	case6 := "RemovePRLabels"
	cli.labels = []string{squashLabel}
	err = bot.handledSquash(comment, commenter, org, repo, number)
	execMethod6 := cli.method
	assert.Equal(t, case6, execMethod6)
	assert.Equal(t, err, nil)

	// squash fail
	case7 := "CreatePRComment"
	comment = "/squash"
	cli.labels = []string{rebaseLabel}
	cli.isGetPullRequestLabels = true
	err = bot.handledSquash(comment, commenter, org, repo, number)
	execMethod7 := cli.method
	assert.Equal(t, case7, execMethod7)
	assert.Equal(t, err, nil)
	// squash cancel fail
	comment = "/squash cancel"
	case7 = "RemovePRLabels"
	err = bot.handledSquash(comment, commenter, org, repo, number)
	execMethod8 := cli.method
	assert.Equal(t, case7, execMethod8)
	assert.Equal(t, err, nil)
}

func TestLgtm(t *testing.T) {
	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{}}
	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)
	repoCnf := &repoConfig{
		LgtmCountsRequired: 1,
	}

	// lgtm success
	case9 := "CreatePRComment"
	comment := "/lgtm"
	cli.isCheckPermission = true
	cli.hasPermission = true
	cli.labels = []string{}
	cli.isAddPRLabels = true
	cli.isCreatePRComment = true
	err := bot.handleLGTM(repoCnf, comment, commenter, author, org, repo, number)
	execMethod9 := cli.method
	assert.Equal(t, case9, execMethod9)
	assert.Equal(t, err, nil)
	// lgtm cancel success
	comment = "/lgtm cancel"
	case10 := "CreatePRComment"
	cli.labels = []string{lgtmLabel}
	err = bot.handleLGTM(repoCnf, comment, commenter, author, org, repo, number)
	execMethod10 := cli.method
	assert.Equal(t, case10, execMethod10)
	assert.Equal(t, err, nil)

	// lgtm fail
	case11 := "CreatePRComment"
	comment = "/lgtm"
	cli.isCheckPermission = true
	cli.hasPermission = false
	cli.labels = []string{}
	err = bot.handleLGTM(repoCnf, comment, commenter, author, org, repo, number)
	execMethod11 := cli.method
	assert.Equal(t, case11, execMethod11)
	assert.Equal(t, err, nil)
	// lgtm fail
	comment = "/lgtm"
	cli.isCheckPermission = true
	cli.hasPermission = false
	cli.labels = []string{rebaseLabel}
	err = bot.handleLGTM(repoCnf, comment, commenter, commenter, org, repo, number)
	execMethod12 := cli.method
	assert.Equal(t, case11, execMethod12)
	assert.Equal(t, err, nil)
	// lgtm cancel fail
	comment = "/lgtm cancel"
	cli.isCheckPermission = true
	cli.hasPermission = false
	cli.labels = []string{lgtmLabel}
	err = bot.handleLGTM(repoCnf, comment, commenter, author, org, repo, number)
	execMethod13 := cli.method
	assert.Equal(t, case11, execMethod13)
	assert.Equal(t, err, nil)
}

func TestApprove(t *testing.T) {
	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{}}
	cli, ok := bot.cli.(*mockClient)
	repoCnf := &repoConfig{
		LgtmCountsRequired: 1,
	}
	assert.Equal(t, true, ok)

	// approve success
	case14 := "CreatePRComment"
	comment := "/approve"
	cli.isCheckPermission = true
	cli.hasPermission = true
	cli.isAddPRLabels = true
	cli.isCreatePRComment = true
	cli.labels = []string{}
	err := bot.handleApprove(repoCnf, comment, commenter, author, org, repo, number)
	execMethod14 := cli.method
	assert.Equal(t, case14, execMethod14)
	assert.Equal(t, err, nil)
	// approve cancel success
	comment = "/approve cancel"
	case15 := "CreatePRComment"
	cli.labels = []string{approvedLabel}
	err = bot.handleApprove(repoCnf, comment, commenter, author, org, repo, number)
	execMethod15 := cli.method
	assert.Equal(t, case15, execMethod15)
	assert.Equal(t, err, nil)

	// approve fail
	case16 := "CreatePRComment"
	comment = "/approve"
	cli.isCheckPermission = true
	cli.hasPermission = false
	cli.labels = []string{}
	err = bot.handleApprove(repoCnf, comment, commenter, author, org, repo, number)
	execMethod16 := cli.method
	assert.Equal(t, case16, execMethod16)
	assert.Equal(t, err, nil)
	// approve cancel fail
	comment = "/approve cancel"
	cli.isCheckPermission = true
	cli.hasPermission = false
	cli.labels = []string{approvedLabel}
	err = bot.handleApprove(repoCnf, comment, commenter, author, org, repo, number)
	execMethod17 := cli.method
	assert.Equal(t, case16, execMethod17)
	assert.Equal(t, err, nil)
}

func TestCheckPr(t *testing.T) {
	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{}}
	cli, ok := bot.cli.(*mockClient)
	repoCnf := &repoConfig{
		LegalOperator: commenter,
	}
	assert.Equal(t, true, ok)

	// check-pr success
	case18 := "MergePullRequest"
	comment := "/check-pr"
	cli.isCheckPermission = true
	cli.hasPermission = true
	cli.isCreatePRComment = true
	cli.isListPullRequestOperationLogs = true
	cli.pullRequestOperationLogs = []client.PullRequestOperationLog{{UserName: commenter, Action: ActionAddLabel, Content: ActionAddLabel + " " + approvedLabel, CreatedAt: time.Now()}}
	cli.labels = []string{approvedLabel}
	cli.isGetPullRequestLabels = true
	cli.isMergePullRequest = true
	err := bot.handleCheckPR(repoCnf, comment, commenter, org, repo, number)
	execMethod18 := cli.method
	assert.Equal(t, case18, execMethod18)
	assert.Equal(t, err, nil)
	// check-pr fail
	case19 := "CreatePRComment"
	comment = "/check-pr"
	cli.pullRequestOperationLogs = []client.PullRequestOperationLog{}
	err = bot.handleCheckPR(repoCnf, comment, commenter, org, repo, number)
	execMethod19 := cli.method
	err1 := fmt.Errorf("**The following label is  not ready**.\n\n" +
		"The corresponding operation log is missing." +
		" you should delete the label **approved** and add it again by correct way")
	assert.Equal(t, case19, execMethod19)
	assert.Equal(t, err1, err)
}
