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

package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Instance
// +k8s:openapi-gen=true
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

// InstanceList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Instance `json:"items"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	AlarmStatus      string   `json:"alarm_status,omitempty"`
	CpuTopology      string   `json:"cpu_topology,omitempty"`
	CreateTime       string   `json:"create_time,omitempty"`
	Description      string   `json:"description,omitempty"`
	Device           string   `json:"device,omitempty"`
	DnsAliases       []string `json:"dns_aliases,omitempty"`
	Eip              string   `json:"eip,omitempty"`
	Extra            string   `json:"extra,omitempty"`
	GraphicsPasswd   string   `json:"graphics_passwd,omitempty"`
	GraphicsProtocol string   `json:"graphics_protocol,omitempty"`
	Image            string   `json:"image,omitempty"`
	InstanceClass    int      `json:"instance_class,omitempty"`
	InstanceId       string   `json:"instance_id,omitempty"`
	InstanceName     string   `json:"instance_name,omitempty"`
	InstanceType     string   `json:"instance_type,omitempty"`
	KeypairIds       []string `json:"keypair_ids,omitempty"`
	MemoryCurrent    int      `json:"memory_current,omitempty"`
	Repl             string   `json:"repl,omitempty"`
	SecurityGroup    string   `json:"security_group,omitempty"`
	Status           string   `json:"status,omitempty"`
	StatusTime       string   `json:"status_time,omitempty"`
	SubCode          int      `json:"sub_code,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	TransitionStatus string   `json:"transition_status,omitempty"`
	VcpusCurrent     int      `json:"vcpus_current,omitempty"`
	VolumeIds        []string `json:"volume_ids,omitempty"`
	Volumes          []string `json:"volumes,omitempty"`
	Vxnets           []string `json:"vxnets,omitempty"`
	ZoneId           string   `json:"zone_id,omitempty"`
	CPU              uint8    `json:"cpu,omitempty"`
	MEM              string   `json:"mem,omitempty"`
}

var _ resource.Object = &Instance{}
var _ resourcestrategy.Validater = &Instance{}

func (in *Instance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Instance) NamespaceScoped() bool {
	return true
}

func (in *Instance) New() runtime.Object {
	return &Instance{}
}

func (in *Instance) NewList() runtime.Object {
	return &InstanceList{}
}

func (in *Instance) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "db.pitrix.qingcloud.com",
		Version:  "v1alpha1",
		Resource: "instances",
	}
}

func (in *Instance) IsStorageVersion() bool {
	return true
}

func (in *Instance) Validate(ctx context.Context) field.ErrorList {
	// TODO(user): Modify it, adding your API validation here.
	return nil
}

var _ resource.ObjectList = &InstanceList{}

func (in *InstanceList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	ResourceId string `json:"resource_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

func (in InstanceStatus) SubResourceName() string {
	return "status"
}

// Instance implements ObjectWithStatusSubResource interface.
var _ resource.ObjectWithStatusSubResource = &Instance{}

func (in *Instance) GetStatus() resource.StatusSubResource {
	return in.Status
}

// InstanceStatus{} implements StatusSubResource interface.
var _ resource.StatusSubResource = &InstanceStatus{}

func (in InstanceStatus) CopyTo(parent resource.ObjectWithStatusSubResource) {
	parent.(*Instance).Status = in
}
