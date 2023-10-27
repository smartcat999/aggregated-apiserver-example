/*
Copyright 2023.

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

package main

import (
	"flag"
	"github.com/smartcat999/k8s-iaas/pkg/handler"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	// +kubebuilder:scaffold:resource-imports
	dbv1alpha1 "github.com/smartcat999/k8s-iaas/pkg/apis/db/v1alpha1"
)

func main() {
	fg := flag.NewFlagSet("k8s-iaas", flag.PanicOnError)
	klog.InitFlags(fg)
	fg.Set("v", "4")
	err := builder.APIServer.
		// +kubebuilder:scaffold:resource-register
		WithResourceAndHandler(&dbv1alpha1.Instance{}, handler.ExampleHandlerProvider).
		Execute()
	if err != nil {
		klog.Fatal(err)
	}
}