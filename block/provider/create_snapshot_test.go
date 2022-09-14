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

// Package provider ...
package provider

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateSnapshot(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		snapshotService *serviceFakes.SnapshotManager
	)
	timeNow := time.Now()

	testCases := []struct {
		testCaseName               string
		sourceVolumeID             string
		baseSnapshot               *models.Snapshot
		providerSnapshotParameters *provider.SnapshotParameters
		providerSnapshot           *provider.Snapshot
		setup                      func()

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, snapshotResponse *provider.Snapshot, err error)
	}{
		{
			testCaseName:   "Source VolumeID for snapshot not provided",
			sourceVolumeID: "",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: "test snapshot",
			},
			providerSnapshot: nil,
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: snapshotReadyState,
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:InvalidRequest, Description:'SourceVolumeID is required to complete the operation.",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:   "Source Volume provided for snapshot not present",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: "test snapshot",
			},
			providerSnapshot: nil,
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: snapshotReadyState,
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:RetrivalFailed, Description:'A volume with the specified volume ID '16f293bf-test-4bff-816f-e199c0c65db5' could not be found.",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		},
		{
			testCaseName:   "Snapshot creation failed",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: "test snapshot",
			},
			providerSnapshot:   nil,
			baseSnapshot:       nil,
			expectedErr:        "{Code:ErrorUnclassified, Type:ProvisioningFailed, Description:'Snapshot creation failed",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:   "Snapshot name is empty",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test-snapshot-name",
				LifecycleState: "pending",
				SourceVolume:   &models.SourceVolume{ID: "16f293bf-test-4bff-816f-e199c0c65db6"},
				CreatedAt:      &timeNow,
				Size:           100,
			},
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: "",
			},
			providerSnapshot: &provider.Snapshot{
				VolumeID:             "16f293bf-test-4bff-816f-e199c0c65db6",
				SnapshotID:           "16f293bf-test-4bff-816f-e199c0c65db5",
				SnapshotSize:         100,
				SnapshotCreationTime: timeNow,
				ReadyToUse:           false,
				VPC:                  provider.VPC{Href: "href"},
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.NotNil(t, snapshotResponse)
				assert.Nil(t, err)
			},
		}, {
			testCaseName: "Snapshot in pending state",
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test-snapshot-name",
				LifecycleState: "pending",
				SourceVolume:   &models.SourceVolume{ID: "16f293bf-test-4bff-816f-e199c0c65db6"},
				CreatedAt:      &timeNow,
				Size:           100,
			},
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: "test snapshot name",
			},
			providerSnapshot: &provider.Snapshot{
				VolumeID:             "16f293bf-test-4bff-816f-e199c0c65db6",
				SnapshotID:           "16f293bf-test-4bff-816f-e199c0c65db5",
				SnapshotSize:         100,
				SnapshotCreationTime: timeNow,
				ReadyToUse:           false,
				VPC:                  provider.VPC{Href: "href"},
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.NotNil(t, snapshotResponse)
				assert.Nil(t, err)
			},
		},
		{
			testCaseName: "Snapshot in stable state",
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test-snapshot-name",
				LifecycleState: snapshotReadyState,
				SourceVolume:   &models.SourceVolume{ID: "16f293bf-test-4bff-816f-e199c0c65db6"},
				CreatedAt:      &timeNow,
				Size:           100,
			},
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: "test snapshot name",
			},
			providerSnapshot: &provider.Snapshot{
				VolumeID:             "16f293bf-test-4bff-816f-e199c0c65db6",
				SnapshotID:           "16f293bf-test-4bff-816f-e199c0c65db5",
				SnapshotSize:         100,
				SnapshotCreationTime: timeNow,
				ReadyToUse:           true,
				VPC:                  provider.VPC{Href: "href"},
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.NotNil(t, snapshotResponse)
				assert.Nil(t, err)
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.testCaseName, func(t *testing.T) {
			vpcs, uc, sc, err := GetTestOpenSession(t, logger)
			assert.NotNil(t, vpcs)
			assert.NotNil(t, uc)
			assert.NotNil(t, sc)
			assert.Nil(t, err)

			snapshotService = &serviceFakes.SnapshotManager{}
			assert.NotNil(t, snapshotService)
			uc.SnapshotServiceReturns(snapshotService)

			if testcase.expectedErr != "" {
				snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, errors.New(testcase.expectedReasonCode))
			} else {
				snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, nil)
			}
			snapshot, err := vpcs.CreateSnapshot(testcase.sourceVolumeID, *testcase.providerSnapshotParameters)
			logger.Info("snapshot details", zap.Reflect("snapshot", snapshot))

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, snapshot, err)
			}
		})
	}
}
