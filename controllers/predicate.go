package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = filter{}

type filter struct {
	namespaces []string
}

func (f filter) allowed(namespace string) bool {
	for _, n := range f.namespaces {
		if n == namespace {
			return true
		}
	}
	return false
}

func (f filter) Create(event event.CreateEvent) bool {
	return f.allowed(event.Meta.GetNamespace())
}

func (f filter) Delete(event event.DeleteEvent) bool {
	return f.allowed(event.Meta.GetNamespace())
}

func (f filter) Update(event event.UpdateEvent) bool {
	return f.allowed(event.MetaNew.GetNamespace())
}

func (f filter) Generic(event event.GenericEvent) bool {
	return f.allowed(event.Meta.GetNamespace())
}

