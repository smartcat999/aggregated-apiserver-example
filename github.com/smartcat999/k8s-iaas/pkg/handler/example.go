package handler

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	regsitryrequest "k8s.io/apiserver/pkg/endpoints/request"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	regsitryrest "k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/rest"

	dbv1alpha1 "github.com/smartcat999/k8s-iaas/pkg/apis/db/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ rest.ResourceHandlerProvider = ExampleHandlerProvider

func ExampleHandlerProvider(s *runtime.Scheme, _ genericregistry.RESTOptionsGetter) (regsitryrest.Storage, error) {
	obj := dbv1alpha1.Instance{}
	return &ExampleHandler{
		DefaultStrategy: builder.DefaultStrategy{
			Object:      &obj,
			ObjectTyper: s,
			TableConvertor: regsitryrest.NewDefaultTableConvertor(
				obj.GetGroupVersionResource().GroupResource()),
		},
		MemStorage: &map[string]*map[string]*map[string]runtime.Object{},
	}, nil
}

var _ regsitryrest.Getter = &ExampleHandler{}
var _ regsitryrest.Lister = &ExampleHandler{}
var _ regsitryrest.CreaterUpdater = &ExampleHandler{}
var _ regsitryrest.GracefulDeleter = &ExampleHandler{}

type ExampleHandler struct {
	builder.DefaultStrategy
	MemStorage *map[string]*map[string]*map[string]runtime.Object
}

func GetMetaObj(obj runtime.Object) (metav1.Object, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	return accessor, nil
}

func (e *ExampleHandler) Create(ctx context.Context, obj runtime.Object,
	createValidation regsitryrest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	klog.V(4).Infof("create obj: %v", obj)

	metaObj, err := GetMetaObj(obj)
	if err != nil {
		return nil, err
	}
	namespace := metaObj.GetNamespace()
	name := metaObj.GetName()
	regsitryrest.FillObjectMetaSystemFields(metaObj)

	kind := obj.GetObjectKind().GroupVersionKind().String()
	if nsStorage, ok := (*e.MemStorage)[namespace]; !ok {
		(*e.MemStorage)[namespace] = &map[string]*map[string]runtime.Object{
			kind: {
				name: obj,
			},
		}
	} else {
		if kindStorage, ok := (*nsStorage)[kind]; !ok {
			(*nsStorage)[kind] = &map[string]runtime.Object{
				name: obj,
			}
		} else {
			if _, ok := (*kindStorage)[name]; !ok {
				(*kindStorage)[name] = obj
			}
		}
	}
	klog.V(4).Infof("mem obj: %v", e.MemStorage)
	return obj, nil
}

func (e *ExampleHandler) Update(
	ctx context.Context, name string, objInfo regsitryrest.UpdatedObjectInfo,
	createValidation regsitryrest.ValidateObjectFunc, updateValidation regsitryrest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	namespace := regsitryrequest.NamespaceValue(ctx)
	klog.V(4).Infof("update %v in namespace %v", name, namespace)

	kind := schema.GroupVersionKind{
		Group:   "animal.agg.io",
		Version: "v1alpha1",
		Kind:    "Cat",
	}.String()

	if nsStorage, ok := (*e.MemStorage)[namespace]; ok {
		if kindStorage, ok := (*nsStorage)[kind]; ok {
			if obj, ok := (*kindStorage)[name]; ok {
				newObj, err := objInfo.UpdatedObject(ctx, obj)
				if err != nil {
					return nil, false, nil
				}
				(*kindStorage)[name] = newObj
				klog.V(4).Infof("newObj: %v", newObj)
				return newObj, true, nil
			}
		}
	}

	return nil, true, nil
}

func (e *ExampleHandler) NewList() runtime.Object {
	return &dbv1alpha1.InstanceList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    nil,
	}
}

func (e *ExampleHandler) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	namespace := regsitryrequest.NamespaceValue(ctx)
	klog.V(4).Infof("list obj in namespace %v", namespace)

	ret := dbv1alpha1.InstanceList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    []dbv1alpha1.Instance{},
	}
	kind := schema.GroupVersionKind{
		Group:   "animal.agg.io",
		Version: "v1alpha1",
		Kind:    "Cat",
	}.String()

	if namespace == metav1.NamespaceAll {
		for _, nsStorage := range *e.MemStorage {
			if kindStorage, ok := (*nsStorage)[kind]; ok {
				for _, val := range *kindStorage {
					if cat, ok := val.(*dbv1alpha1.Instance); ok {
						ret.Items = append(ret.Items, *cat)
					}
				}
			}
		}
	} else if nsStorage, ok := (*e.MemStorage)[namespace]; ok {
		if kindStorage, ok := (*nsStorage)[kind]; ok {
			for _, val := range *kindStorage {
				if cat, ok := val.(*dbv1alpha1.Instance); ok {
					ret.Items = append(ret.Items, *cat)
				}
			}
		}
	}

	return &ret, nil
}

func (e *ExampleHandler) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	namespace := regsitryrequest.NamespaceValue(ctx)
	klog.V(4).Infof("get %v in namespace %v", name, namespace)

	kind := schema.GroupVersionKind{
		Group:   "animal.agg.io",
		Version: "v1alpha1",
		Kind:    "Cat",
	}.String()

	if nsStorage, ok := (*e.MemStorage)[namespace]; ok {
		if kindStorage, ok := (*nsStorage)[kind]; ok {
			if obj, ok := (*kindStorage)[name]; ok {
				return obj, nil
			}
		}
	}

	return nil, apierrors.NewNotFound((&dbv1alpha1.Instance{}).GetGroupVersionResource().GroupResource(), name)
}

func (e *ExampleHandler) Delete(ctx context.Context, name string, deleteValidation regsitryrest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {

	namespace := regsitryrequest.NamespaceValue(ctx)
	klog.V(4).Infof("delete %v in namespace %v", name, namespace)

	kind := schema.GroupVersionKind{
		Group:   "animal.agg.io",
		Version: "v1alpha1",
		Kind:    "Cat",
	}.String()

	if nsStorage, ok := (*e.MemStorage)[namespace]; ok {
		if kindStorage, ok := (*nsStorage)[kind]; ok {
			if obj, ok := (*kindStorage)[name]; ok {
				delete(*kindStorage, name)
				klog.V(4).Infof("delete %v in namespace %v success", name, namespace)
				return obj, true, nil
			}
		}
	}
	return nil, false, apierrors.NewNotFound((&dbv1alpha1.Instance{}).GetGroupVersionResource().GroupResource(), name)
}

func (e *ExampleHandler) New() runtime.Object {
	return &dbv1alpha1.Instance{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       dbv1alpha1.InstanceSpec{},
		Status:     dbv1alpha1.InstanceStatus{},
	}
}

func (e *ExampleHandler) Destroy() {}
