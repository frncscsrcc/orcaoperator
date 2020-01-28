/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "orcaoperator/pkg/apis/orca/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// IgnitorLister helps list Ignitors.
type IgnitorLister interface {
	// List lists all Ignitors in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Ignitor, err error)
	// Ignitors returns an object that can list and get Ignitors.
	Ignitors(namespace string) IgnitorNamespaceLister
	IgnitorListerExpansion
}

// ignitorLister implements the IgnitorLister interface.
type ignitorLister struct {
	indexer cache.Indexer
}

// NewIgnitorLister returns a new IgnitorLister.
func NewIgnitorLister(indexer cache.Indexer) IgnitorLister {
	return &ignitorLister{indexer: indexer}
}

// List lists all Ignitors in the indexer.
func (s *ignitorLister) List(selector labels.Selector) (ret []*v1alpha1.Ignitor, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Ignitor))
	})
	return ret, err
}

// Ignitors returns an object that can list and get Ignitors.
func (s *ignitorLister) Ignitors(namespace string) IgnitorNamespaceLister {
	return ignitorNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// IgnitorNamespaceLister helps list and get Ignitors.
type IgnitorNamespaceLister interface {
	// List lists all Ignitors in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Ignitor, err error)
	// Get retrieves the Ignitor from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Ignitor, error)
	IgnitorNamespaceListerExpansion
}

// ignitorNamespaceLister implements the IgnitorNamespaceLister
// interface.
type ignitorNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Ignitors in the indexer for a given namespace.
func (s ignitorNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Ignitor, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Ignitor))
	})
	return ret, err
}

// Get retrieves the Ignitor from the indexer for a given namespace and name.
func (s ignitorNamespaceLister) Get(name string) (*v1alpha1.Ignitor, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("ignitor"), name)
	}
	return obj.(*v1alpha1.Ignitor), nil
}
