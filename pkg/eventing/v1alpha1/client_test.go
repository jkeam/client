// Copyright © 2019 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1/fake"
)

var testNamespace = "test-ns"

func setup() (fakeSvr fake.FakeEventingV1alpha1, client KnEventingClient) {
	fakeE := fake.FakeEventingV1alpha1{Fake: &client_testing.Fake{}}
	cli := NewKnEventingClient(&fakeE, testNamespace)
	return fakeE, cli
}

func TestDeleteTrigger(t *testing.T) {
	var name = "new-trigger"
	server, client := setup()

	server.AddReactor("delete", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorTrigger" {
				return true, nil, fmt.Errorf("error while deleting trigger %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteTrigger(name)
	assert.NilError(t, err)

	err = client.DeleteTrigger("errorTrigger")
	assert.ErrorContains(t, err, "errorTrigger")
}

func TestCreateTrigger(t *testing.T) {
	var name = "new-trigger"
	server, client := setup()

	objNew := newTrigger(name)

	server.AddReactor("create", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == objNew.Name {
				objNew.Generation = 2
				return true, objNew, nil
			}
			return true, nil, fmt.Errorf("error while creating trigger %s", name)
		})

	t.Run("create trigger without error", func(t *testing.T) {
		ins, err := client.CreateTrigger(objNew)
		assert.NilError(t, err)
		assert.Equal(t, ins.Name, name)
		assert.Equal(t, ins.Namespace, testNamespace)
	})

	t.Run("create trigger with an error returns an error object", func(t *testing.T) {
		_, err := client.CreateTrigger(newTrigger("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestGetTrigger(t *testing.T) {
	var name = "mytrigger"
	server, client := setup()

	server.AddReactor("get", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "errorTrigger" {
				return true, nil, fmt.Errorf("error while getting trigger %s", name)
			}
			return true, newTrigger(name), nil
		})

	trigger, err := client.GetTrigger(name)
	assert.NilError(t, err)
	assert.Equal(t, trigger.Name, name)
	assert.Equal(t, trigger.Spec.Broker, "default")

	_, err = client.GetTrigger("errorTrigger")
	assert.ErrorContains(t, err, "errorTrigger")
}

func newTrigger(name string) *v1alpha1.Trigger {
	obj := &v1alpha1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: "default",
			Filter: &v1alpha1.TriggerFilter{
				Attributes: &v1alpha1.TriggerFilterAttributes{
					"type": "foo",
				},
			},
		},
	}
	obj.Name = name
	obj.Namespace = testNamespace
	return obj
}