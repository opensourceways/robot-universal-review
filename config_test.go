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
	"os"
	"path/filepath"
	"testing"

	"github.com/opensourceways/server-common-lib/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {

	type args struct {
		cnf  *configuration
		path string
	}

	testCases := []struct {
		desc string
		in   args
		out  [2]error
	}{
		{
			"config is nil",
			args{
				nil,
				"",
			},
			[2]error{nil, errors.New("configuration is nil")},
		},
		{
			"config is empty",
			args{
				&configuration{},
				"",
			},
			[2]error{nil, nil},
		},
		{
			"no valid org or repo in the config",
			args{
				&configuration{},
				"config1.yaml",
			},
			[2]error{nil, errors.New("the repositories configuration can not be empty")},
		},
		{
			"the same org and repo conflicts in the config",
			args{
				&configuration{},
				"config2.yaml",
			},
			[2]error{nil, errors.New("some org or org/repo exists in both repos and excluded_repos")},
		},
		{
			"a correct config",
			args{
				&configuration{},
				configYaml,
			},
			[2]error{nil, nil},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].desc, func(t *testing.T) {
			if testCases[i].in.path != "" {
				err := utils.LoadFromYaml(findTestdata(t, testCases[i].in.path), testCases[i].in.cnf)
				assert.Equal(t, testCases[i].out[0], err)
			}

			err1 := testCases[i].in.cnf.Validate()
			assert.Equal(t, testCases[i].out[1], err1)
		})
	}

}

func TestGetRepoConfig(t *testing.T) {
	cnf := &configuration{}
	got := cnf.get("owner1", "")
	assert.Equal(t, (*repoConfig)(nil), got)

	err := utils.LoadFromYaml(findTestdata(t, configYaml), cnf)
	assert.Equal(t, nil, err)

	testCases := []struct {
		desc string
		in   [2]string
		out  *repoConfig
	}{
		{
			"org and repo are all empty",
			[2]string{"", ""},
			(*repoConfig)(nil),
		},
		{
			"org is empty, repo is not empty",
			[2]string{"", "repo1"},
			(*repoConfig)(nil),
		},
		{
			"org is not empty, repo is empty",
			[2]string{"owner3", ""},
			&cnf.ConfigItems[1],
		},
		{
			"org is not empty, repo is not empty",
			[2]string{"owner2", "repo1"},
			&cnf.ConfigItems[0],
		},
	}

	for i := range testCases {
		t.Run(testCases[i].desc, func(t *testing.T) {

			got = cnf.get(testCases[i].in[0], testCases[i].in[1])
			if testCases[i].out == nil {
				assert.Equal(t, testCases[i].out, got)
			} else {
				assert.Equal(t, true, got != nil)
				assert.Equal(t, *testCases[i].out, *got)
			}

		})
	}
}

func findTestdata(t *testing.T, path string) string {
	path = "testdata" + string(os.PathSeparator) + path
	i := 0
retry:
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Error(path + " not found")
		return ""
	}
	if _, err = os.Stat(absPath); !os.IsNotExist(err) {
		return absPath
	} else {
		i++
		path = ".." + string(os.PathSeparator) + path
		if i <= 3 {
			goto retry
		}
	}

	t.Log(path + " not found")
	return ""
}
