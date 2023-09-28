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
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	// +kubebuilder:scaffold:resource-imports
	animalv1alpha1 "github.com/smartcat999/k8s-aggregated/pkg/apis/animal/v1alpha1"
	"github.com/smartcat999/k8s-aggregated/pkg/handler"
)

func main() {
	fg := flag.NewFlagSet("k8s-aggregated", flag.PanicOnError)
	klog.InitFlags(fg)
	fg.Set("v", "4")
	command, err := builder.APIServer.
		// +kubebuilder:scaffold:resource-register
		//WithResource(&animalv1alpha1.Cat{}).
		WithResourceAndHandler(&animalv1alpha1.Cat{}, handler.ExampleHandlerProvider).
		WithoutEtcd().
		Build()

	//WithResourceAndStorage(&animalv1alpha1.Cat{}, mysql.NewMysqlStorageProvider(
	//	"", // mysql host name		e.g. "127.0.0.1"
	//	0,  // mysql password 		e.g. 3306
	//	"", // mysql username 		e.g. "mysql"
	//	"", // mysql password 		e.g. "password"
	//	"", // mysql database name 	e.g. "mydb"
	//)).Build()
	if err != nil {
		klog.Fatal(err)
	}
	err = command.Execute()
	if err != nil {
		klog.Fatal(err)
	}
}
