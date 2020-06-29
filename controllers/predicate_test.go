package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestPredicate(t *testing.T) {
	p := filter{
		namespaces: []string{"default", "auth"},
	}

	t.Run("should disallow", func(t *testing.T) {

		unsupported := []metav1.Object{
			&metav1.ObjectMeta{Namespace: "system"},
			&metav1.ObjectMeta{Namespace: "testing"},
		}
		msg := "namespace \"%s\" should NOT be allowed"

		t.Run("CreateEvents outside the supported namespaces", func(t *testing.T) {
			for _, meta := range unsupported {
				assert.False(t, p.Create(event.CreateEvent{Meta: meta}), msg, meta.GetNamespace())
			}
		})

		t.Run("DeleteEvents outside the supported namespaces", func(t *testing.T) {
			for _, meta := range unsupported {
				assert.False(t, p.Delete(event.DeleteEvent{Meta: meta}), msg, meta.GetNamespace())
			}
		})

		t.Run("UpdateEvents outside the supported namespaces", func(t *testing.T) {
			for _, meta := range unsupported {
				assert.False(t, p.Update(event.UpdateEvent{MetaNew: meta}), msg, meta.GetNamespace())
			}
		})

		t.Run("GenericEvents outside the supported namespaces", func(t *testing.T) {
			for _, meta := range unsupported {
				assert.False(t, p.Generic(event.GenericEvent{Meta: meta}), msg, meta.GetNamespace())
			}
		})
	})

	t.Run("should allow", func(t *testing.T) {
		var supported []metav1.Object
		for _, namespace := range p.namespaces {
			supported = append(supported, &metav1.ObjectMeta{Namespace: namespace})
		}
		msg := "namespace \"%s\" should be allowed"

		t.Run("CreateEvents inside the supported namespaces", func(t *testing.T) {
			for _, meta := range supported {
				assert.True(t, p.Create(event.CreateEvent{Meta: meta}), msg, meta.GetNamespace())
			}
		})

		t.Run("DeleteEvents inside the supported namespaces", func(t *testing.T) {
			for _, meta := range supported {
				assert.True(t, p.Delete(event.DeleteEvent{Meta: meta}), msg, meta.GetNamespace())
			}
		})

		t.Run("UpdateEvents inside the supported namespaces", func(t *testing.T) {
			for _, meta := range supported {
				assert.True(t, p.Update(event.UpdateEvent{MetaNew: meta}), msg, meta.GetNamespace())
			}
		})

		t.Run("GenericEvents inside the supported namespaces", func(t *testing.T) {
			for _, meta := range supported {
				assert.True(t, p.Generic(event.GenericEvent{Meta: meta}), msg, meta.GetNamespace())
			}
		})
	})
}
