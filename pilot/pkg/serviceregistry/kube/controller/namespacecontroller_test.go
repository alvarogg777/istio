// Copyright Istio Authors
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

package controller

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listerv1 "k8s.io/client-go/listers/core/v1"

	"istio.io/istio/pkg/kube"
	"istio.io/istio/pkg/kube/inject"
	"istio.io/istio/pkg/test/util/retry"
	"istio.io/istio/security/pkg/k8s"
)

func TestNamespaceController(t *testing.T) {
	client := kube.NewFakeClient()
	testdata := map[string]string{"key": "value"}
	nc := NewNamespaceController(func() map[string]string {
		return testdata
	}, client)
	nc.configmapLister = client.KubeInformer().Core().V1().ConfigMaps().Lister()
	stop := make(chan struct{})
	t.Cleanup(func() {
		close(stop)
	})
	client.RunAndWait(stop)
	nc.Run(stop)

	createNamespace(t, client, "foo", nil)
	expectConfigMap(t, nc.configmapLister, "foo", testdata)

	newData := map[string]string{"key": "value", "foo": "bar"}
	if err := k8s.InsertDataToConfigMap(client.CoreV1(), nc.configmapLister,
		metav1.ObjectMeta{Name: CACertNamespaceConfigMap, Namespace: "foo"}, newData); err != nil {
		t.Fatal(err)
	}
	expectConfigMap(t, nc.configmapLister, "foo", newData)

	deleteConfigMap(t, client, "foo")
	expectConfigMap(t, nc.configmapLister, "foo", testdata)

	for _, namespace := range inject.IgnoredNamespaces {
		createNamespace(t, client, namespace, testdata)
		expectConfigMapNotExist(t, nc.configmapLister, namespace)
	}
}

func deleteConfigMap(t *testing.T, client kubernetes.Interface, ns string) {
	t.Helper()
	_, err := client.CoreV1().ConfigMaps(ns).Get(context.TODO(), CACertNamespaceConfigMap, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CoreV1().ConfigMaps(ns).Delete(context.TODO(), CACertNamespaceConfigMap, metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
}

func createNamespace(t *testing.T, client kubernetes.Interface, ns string, labels map[string]string) {
	t.Helper()
	if _, err := client.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
}

func updateNamespace(t *testing.T, client kubernetes.Interface, ns string, labels map[string]string) {
	t.Helper()
	if _, err := client.CoreV1().Namespaces().Update(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels},
	}, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
}

func expectConfigMap(t *testing.T, client listerv1.ConfigMapLister, ns string, data map[string]string) {
	t.Helper()
	retry.UntilSuccessOrFail(t, func() error {
		cm, err := client.ConfigMaps(ns).Get(CACertNamespaceConfigMap)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(cm.Data, data) {
			return fmt.Errorf("data mismatch, expected %+v got %+v", data, cm.Data)
		}
		return nil
	}, retry.Timeout(time.Second*2))
}

func expectConfigMapNotExist(t *testing.T, client listerv1.ConfigMapLister, ns string) {
	t.Helper()
	err := retry.Until(func() bool {
		_, err := client.ConfigMaps(ns).Get(CACertNamespaceConfigMap)
		return err == nil
	}, retry.Timeout(time.Second*2))

	if err == nil {
		t.Fatalf("%s namespace should not have istio-ca-root-cert configmap.", ns)
	}
}
