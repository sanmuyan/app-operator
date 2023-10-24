package controller

import (
	corev1 "k8s.io/api/core/v1"
	appv1 "sanmuyan.com/app-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	acLog          = log.Log.WithName("appconfig-controller")
	ownerKey       = ".metadata.controller"
	apiGVStr       = appv1.GroupVersion.String()
	templateConfig = &corev1.ConfigMap{
		Data: make(map[string]string),
	}
)

const (
	apiKind          = "AppConfig"
	appContainerName = "app"
)
