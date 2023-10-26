package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddOtherAnnotation(o metav1.Object, k, v string) {
	if o.GetAnnotations() == nil {
		o.SetAnnotations(make(map[string]string))
	}
	o.GetAnnotations()[k] = v
}

func AddAnnotation(o metav1.Object, k, v string) {
	if o.GetAnnotations() == nil {
		o.SetAnnotations(make(map[string]string))
	}
	o.GetAnnotations()[LabelPrefix+"/"+k] = v
}

func AddOtherLabel(o metav1.Object, k, v string) {
	if o.GetLabels() == nil {
		o.SetLabels(make(map[string]string))
	}
	o.GetLabels()[k] = v
}

func AddLabel(o metav1.Object, k, v string) {
	if o.GetLabels() == nil {
		o.SetLabels(make(map[string]string))
	}
	o.GetLabels()[LabelPrefix+"/"+k] = v
}

func GetAnnotation(o metav1.Object, k string) string {
	if o.GetAnnotations() == nil {
		return ""
	}
	if annotation, ok := o.GetAnnotations()[LabelPrefix+"/"+k]; ok {
		return annotation
	}
	return ""
}
