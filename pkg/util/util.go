package util

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsContainsAnnotation(o client.Object, k string) bool {
	if o.GetAnnotations() == nil {
		return false
	}
	_, ok := o.GetAnnotations()[k]
	return ok
}

func AddAnnotation(o client.Object, k, v string) {
	if o.GetAnnotations() == nil {
		o.SetAnnotations(make(map[string]string))
	}
	o.GetAnnotations()[k] = v
}

func GetAnnotation(o client.Object, k string) string {
	if o.GetAnnotations() == nil {
		return ""
	}
	if _, ok := o.GetAnnotations()[k]; ok {
		return o.GetAnnotations()[k]
	}
	return ""
}
