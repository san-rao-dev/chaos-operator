/*
Copyright 2019 LitmusChaos Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
   http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	litmusFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"

	chaosTypes "github.com/litmuschaos/chaos-operator/pkg/types"

	"reflect"
)

func TestGetChaosRunnerENV(t *testing.T) {
	fakeEngineName := "Fake Engine"
	fakeNameSpace := "Fake NameSpace"
	fakeServiceAcc := "Fake Service Account"
	fakeTargets := "fakeTargets"
	fakeAppLabel := "Fake Label"
	fakeAppKind := "Fake Kind"
	fakeAExList := []string{"fake string"}
	fakeAuxilaryAppInfo := "ns1:name=percona,ns2:run=nginx"
	fakeClientUUID := "12345678-9012-3456-7890-123456789012"

	tests := map[string]struct {
		instance       *v1alpha1.ChaosEngine
		aExList        []string
		expectedResult []corev1.EnvVar
	}{
		"Test Positive-1": {
			instance: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeEngineName,
					Namespace: fakeNameSpace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					Appinfo: v1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    fakeNameSpace,
						AppKind:  fakeAppKind,
					},
					AuxiliaryAppInfo: fakeAuxilaryAppInfo,
				},
			},
			aExList: fakeAExList,
			expectedResult: []corev1.EnvVar{
				{
					Name:  "CHAOSENGINE",
					Value: fakeEngineName,
				},
				{
					Name:  "TARGETS",
					Value: fakeTargets,
				},
				{
					Name:  "EXPERIMENT_LIST",
					Value: fmt.Sprint(strings.Join(fakeAExList, ",")),
				},
				{
					Name:  "CHAOS_SVC_ACC",
					Value: fakeServiceAcc,
				},
				{
					Name:  "AUXILIARY_APPINFO",
					Value: fakeAuxilaryAppInfo,
				},
				{
					Name:  "CLIENT_UUID",
					Value: fakeClientUUID,
				},
				{
					Name:  "CHAOS_NAMESPACE",
					Value: fakeNameSpace,
				},
			},
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			engine := &chaosTypes.EngineInfo{Instance: mock.instance, Targets: fakeTargets, AppExperiments: fakeAExList}
			actualResult := getChaosRunnerENV(engine, fakeClientUUID)
			println(len(actualResult))
			if len(actualResult) != 7 {
				t.Fatalf("Test %q failed: expected array length to be 7", name)
			}
			for index, result := range actualResult {
				if result.Value != mock.expectedResult[index].Value {
					t.Fatalf("Test %q failed: actual result %q, received result %q", name, result, mock.expectedResult[index])
				}
			}
		})
	}
}

func TestUpdateEngineForComplete(t *testing.T) {
	tests := map[string]struct {
		engine chaosTypes.EngineInfo
		isErr  bool
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-complete-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-complete-p2",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusInitialized,
					},
				},
			},
			isErr: false,
		},
		"Test Positive-3": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-complete-p3",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusStopped,
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			err := r.Client.Create(context.TODO(), mock.engine.Instance)
			if err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}

			_, err = r.updateEngineForComplete(&mock.engine, true)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestUpdateEngineForRestart(t *testing.T) {
	tests := map[string]struct {
		engine chaosTypes.EngineInfo
		isErr  bool
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-restart-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-restart-p2",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusInitialized,
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			err := r.Client.Create(context.TODO(), mock.engine.Instance)
			if err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}

			_, err = r.updateEngineForRestart(&mock.engine)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestNewGoRunnerPodForCR(t *testing.T) {
	tests := map[string]struct {
		engine chaosTypes.EngineInfo
		isErr  bool
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
								Command: []string{
									"cmd1",
									"cmd2",
								},
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image:           "fake-runner-image",
								ImagePullPolicy: "Always",
								Args: []string{
									"args1",
									"args2",
								},
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},

			isErr: false,
		},
		"Test Positive-3": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image:           "fake-runner-image",
								ImagePullPolicy: "IfNotPresent",
								Command: []string{
									"cmd1",
									"cmd2",
								},
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},

			isErr: false,
		},
		"Test Positive-4": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image:           "fake-runner-image",
								ImagePullPolicy: "Never",
								Args: []string{
									"args1",
									"args2",
								},
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},

			isErr: false,
		},
		"Test Negative-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{},
				},
				AppExperiments: []string{"exp-1"},
			},
			isErr: true,
		},
		"Test Negative-2 ": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
					},
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: true,
		},
		"Test Negative-3 ": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
					},
				},

				AppExperiments: []string{},
			},
			isErr: true,
		},
		"Test Negative-4 ": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "",
							},
						},
					},
				},

				AppExperiments: []string{},
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			_, err := r.newGoRunnerPodForCR(&mock.engine)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestInitEngine(t *testing.T) {
	tests := map[string]struct {
		engine chaosTypes.EngineInfo
		isErr  bool
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-init-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-init-p2",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: "active",
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusStopped,
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			_, err := r.initEngine(&mock.engine)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestUpdateEngineState(t *testing.T) {
	tests := map[string]struct {
		isErr  bool
		engine chaosTypes.EngineInfo
		state  v1alpha1.EngineState
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-update-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			isErr: false,
			state: v1alpha1.EngineStateActive,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-update-p2",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			isErr: false,
			state: v1alpha1.EngineStateStop,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			err := r.Client.Create(context.TODO(), mock.engine.Instance)
			if err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}
			err = r.updateEngineState(&mock.engine, mock.state)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestCheckRunnerPodCompletedStatus(t *testing.T) {
	tests := map[string]struct {
		isErr  bool
		engine chaosTypes.EngineInfo
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-runner-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusInitialized,
					},
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-runner-p2",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			if err := r.Client.Create(context.TODO(), mock.engine.Instance); err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}
			if err := r.Client.Create(context.TODO(), &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mock.engine.Instance.Name + "-runner",
					Namespace: mock.engine.Instance.Namespace,
				},
			}); err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}
			val, err := r.checkRunnerContainerCompletedStatus(&mock.engine)
			if err != nil {
				fmt.Printf("Unable to check status: %v", err)
			}
			if mock.isErr && val == false {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && val == true {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestEngineRunnerPod(t *testing.T) {
	tests := map[string]struct {
		isErr  bool
		runner *podEngineRunner
	}{
		"Test Positive-1": {
			runner: &podEngineRunner{
				pod: &corev1.Pod{},
				engineRunner: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    make(map[string]string),
						Name:      "dummypod",
						Namespace: "dummyns",
					},
				},
				reconcileEngine: &reconcileEngine{
					r:         CreateFakeClient(t),
					reqLogger: chaosTypes.Log.WithValues(),
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			runner: &podEngineRunner{
				pod: &corev1.Pod{},
				engineRunner: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    make(map[string]string),
						Name:      "dummypresentpod",
						Namespace: "default",
					},
				},
				reconcileEngine: &reconcileEngine{
					r:         CreateFakeClient(t),
					reqLogger: chaosTypes.Log.WithValues(),
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			if name == "Test Positive-2" {
				require.NoError(t, mock.runner.r.Client.Create(context.TODO(), mock.runner.engineRunner))
			}
			err := engineRunnerPod(mock.runner)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestStartReqLogger(t *testing.T) {
	tests := map[string]struct {
		isErr   bool
		request reconcile.Request
	}{
		"Test Positive-1": {
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "default",
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			req := startReqLogger(mock.request)
			if mock.isErr && !reflect.ValueOf(req).IsZero()  {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && reflect.ValueOf(req).IsZero() {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestGetChaosEngineInstance(t *testing.T) {
	tests := map[string]struct {
		isErr   bool
		engine  chaosTypes.EngineInfo
		request reconcile.Request
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-p1",
					Namespace: "default",
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-n1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-n1",
					Namespace: "default",
				},
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			if name == "Test Positive-1" {
				err := r.Client.Create(context.TODO(), mock.engine.Instance)
				if err != nil {
					fmt.Printf("Unable to create engine: %v", err)
				}
			}
			err := r.getChaosEngineInstance(&mock.engine, mock.request)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestCheckEngineRunnerPod(t *testing.T) {
	tests := map[string]struct {
		engine chaosTypes.EngineInfo
		isErr  bool
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},

			isErr: false,
		},
		"Test Positive-3": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},

			isErr: false,
		},
		"Test Positive-4": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
					},
				},

				AppExperiments: []string{"exp-1"},
			},

			isErr: false,
		},
		"Test Negative-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{},
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: true,
		},
		"Test Negative-2 ": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
					},
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: true,
		},
		"Test Negative-3 ": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
					},
				},

				AppExperiments: []string{},
			},
			isErr: true,
		},
		"Test Negative-4 ": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "",
							},
						},
					},
				},

				AppExperiments: []string{},
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			reqLogger := chaosTypes.Log.WithValues()
			err := r.checkEngineRunnerPod(&mock.engine, reqLogger)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestReconcileForDelete(t *testing.T) {
	tests := map[string]struct {
		isErr   bool
		engine  chaosTypes.EngineInfo
		request reconcile.Request
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-p1",
					Namespace: "default",
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-n1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-n1",
					Namespace: "default",
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			if name == "Test Positive-1" {
				err := r.Client.Create(context.TODO(), mock.engine.Instance)
				if err != nil {
					fmt.Printf("Unable to create engine: %v", err)
				}
			}
			_, err := r.reconcileForDelete(&mock.engine, mock.request)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestForceRemoveAllChaosPods(t *testing.T) {
	tests := map[string]struct {
		isErr   bool
		engine  chaosTypes.EngineInfo
		request reconcile.Request
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-p1",
					Namespace: "default",
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-n1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-p2",
					Namespace: "default",
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			err := r.Client.Create(context.TODO(), mock.engine.Instance)
			if err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}
			err = r.forceRemoveChaosResources(&mock.engine, mock.request)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestGracefullyRemoveDefaultChaosResources(t *testing.T) {
	tests := map[string]struct {
		isErr   bool
		engine  chaosTypes.EngineInfo
		request reconcile.Request
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-p1",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-p1",
					Namespace: "default",
				},
			},
			isErr: false,
		},
		"Test Positive-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "engine-instance-p2",
						Namespace: "default",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "app=nginx",
							AppKind:  "deployment",
						},
						EngineState: v1alpha1.EngineStateActive,
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
					Status: v1alpha1.ChaosEngineStatus{
						EngineStatus: v1alpha1.EngineStatusCompleted,
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "engine-instance-p2",
					Namespace: "default",
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			err := r.Client.Create(context.TODO(), mock.engine.Instance)
			if err != nil {
				fmt.Printf("Unable to create engine: %v", err)
			}
			_, err = r.gracefullyRemoveDefaultChaosResources(&mock.engine, mock.request)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestReconcileForCreationAndRunning(t *testing.T) {
	tests := map[string]struct {
		engine chaosTypes.EngineInfo
		isErr  bool
	}{
		"Test Positive-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "reconcile-1",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "fakeAppLabel",
							Appns:    "fakeAppNs",
							AppKind:  "fakeAppKind",
						},
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
						Experiments: []v1alpha1.ExperimentList{
							{
								Name: "exp-1",
							},
						},
					},
				},

				AppInfo: v1alpha1.ApplicationParams{
					Applabel: "fakeAppLabel",
					Appns:    "fakeAppNs",
					AppKind:  "fakeAppKind",
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: false,
		},
		"Test Negative-1": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "reconcile-2",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "fakeAppLabel",
							Appns:    "fakeAppNs",
							AppKind:  "fakeAppKind",
						},
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
					},
				},

				AppInfo: v1alpha1.ApplicationParams{
					Applabel: "fakeAppLabel",
					Appns:    "fakeAppNs",
					AppKind:  "fakeAppKind",
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: true,
		},
		"Test Negative-2": {
			engine: chaosTypes.EngineInfo{
				Instance: &v1alpha1.ChaosEngine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "reconcile-3",
						Namespace: "test",
					},
					Spec: v1alpha1.ChaosEngineSpec{
						ChaosServiceAccount: "fake-serviceAccount",
						Appinfo: v1alpha1.ApplicationParams{
							Applabel: "fakeAppLabel",
							Appns:    "fakeAppNs",
							AppKind:  "fakeAppKind",
						},
						Components: v1alpha1.ComponentParams{
							Runner: v1alpha1.RunnerInfo{
								Image: "fake-runner-image",
							},
						},
					},
				},
				AppInfo: v1alpha1.ApplicationParams{
					Applabel: "fakeAppLabel",
					Appns:    "fakeAppNs",
					AppKind:  "fakeAppKind",
				},

				AppExperiments: []string{"exp-1"},
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := CreateFakeClient(t)
			reqLogger := chaosTypes.Log.WithValues()
			_, err := r.reconcileForCreationAndRunning(&mock.engine, reqLogger)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func CreateFakeClient(t *testing.T) *ChaosEngineReconciler {

	fakeClient := litmusFakeClientset.NewFakeClient()
	if fakeClient == nil {
		fmt.Println("litmusClient is not created")
	}

	s := scheme.Scheme

	engineR := &v1alpha1.ChaosEngine{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "dummyengine",
		},
	}

	chaosResultList := &v1alpha1.ChaosResultList{
		Items: []v1alpha1.ChaosResult{},
	}

	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, engineR, chaosResultList)

	recorder := record.NewFakeRecorder(1024)

	r := &ChaosEngineReconciler{
		Client:   fakeClient,
		Scheme:   s,
		Recorder: recorder,
	}

	return r
}
