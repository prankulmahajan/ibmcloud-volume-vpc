/**
 * Copyright 2020 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package vpcvolume_test ...
package vpcvolume_test

import (
	"net/http"
	"testing"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetSnapshot(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name string

		// backend url
		url string

		// Response
		status  int
		content string

		// Expected return
		expectErr string
		verify    func(*testing.T, *models.Snapshot, error)
	}{
		{
			name:   "Verify that the correct endpoint is invoked",
			status: http.StatusNoContent,
			url:    vpcvolume.Version + "/snapshots/snapshot1",
		}, {
			name:      "Verify that a 404 is returned to the caller",
			status:    http.StatusNotFound,
			url:       vpcvolume.Version + "/snapshots/snapshot1",
			content:   "{\"errors\":[{\"message\":\"testerr\"}]}",
			expectErr: "Trace Code:, testerr Please check ",
		}, {
			name:    "Verify that the snapshot is parsed correctly",
			status:  http.StatusOK,
			url:     vpcvolume.Version + "/snapshots/snapshot1",
			content: "{\"id\":\"snapshot1\",\"name\":\"snapshot1\",\"lifecycle_state\":\"pending\"}",
			verify: func(t *testing.T, snapshot *models.Snapshot, err error) {
				assert.NotNil(t, snapshot)
				assert.Nil(t, err)
			},
		}, {
			name:    "False positive: What if the snapshot ID is not matched",
			status:  http.StatusOK,
			content: "{\"id\":\"wrong-snap\",\"name\":\"wrong-snap\",\"lifecycle_state\":\"pending\"}",
			verify: func(t *testing.T, snapshot *models.Snapshot, err error) {
				if assert.NotNil(t, snapshot) {
					assert.NotEqual(t, "snap1", snapshot.ID)
				}
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			emptyString := ""
			test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshots/snapshot1", http.MethodGet, &emptyString, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			snapshotService := vpcvolume.NewSnapshotManager(client)
			snapshot, err := snapshotService.GetSnapshot("snapshot1", logger)
			logger.Info("Snapshot details", zap.Reflect("snapshot", snapshot))

			if testcase.verify != nil {
				testcase.verify(t, snapshot, err)
			}
		})
	}
}

func TestGetSnapshotByName(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name string

		// Response
		status    int
		snapshots string
		content   string

		// Expected return
		expectErr string
		verify    func(*testing.T, *models.Snapshot, error)
	}{
		{
			name:   "Verify that the correct endpoint is invoked",
			status: http.StatusNoContent,
		}, {
			name:      "Verify that a 404 is returned to the caller",
			status:    http.StatusNotFound,
			content:   "{\"errors\":[{\"message\":\"testerr\"}]}",
			expectErr: "Trace Code:, testerr Please check ",
		}, {
			name:    "Verify that the snapshot name is parsed correctly",
			status:  http.StatusOK,
			content: "{\"snapshots\":[{\"id\":\"snap1\",\"name\":\"snap1\",\"lifecycle_state\":\"pending\"}]}",
			verify: func(t *testing.T, snapshot *models.Snapshot, err error) {
				if assert.NotNil(t, snapshot) {
					assert.Equal(t, "snap1", snapshot.ID)
				}
			},
		}, {
			name:      "Verify that the snapshot is empty if the snapshots are empty",
			status:    http.StatusOK,
			expectErr: "Trace Code:, testerr Please check ",
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			emptyString := ""
			test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshots", http.MethodGet, &emptyString, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			snapshotService := vpcvolume.NewSnapshotManager(client)
			snapshot, err := snapshotService.GetSnapshotByName("snap1", logger)
			logger.Info("Snapshot details", zap.Reflect("snapshot", snapshot))

			if testcase.verify != nil {
				testcase.verify(t, snapshot, err)
			}
		})
	}
}
