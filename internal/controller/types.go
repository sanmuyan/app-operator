package controller

import (
	corev1 "k8s.io/api/core/v1"
	"os"
	appv1 "sanmuyan.com/app-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	acLog      = log.Log.WithName("appconfig-controller")
	ownerKey   = ".metadata.controller"
	apiGVStr   = appv1.GroupVersion.String()
	templateCM = &corev1.ConfigMap{
		Data: make(map[string]string),
	}
	templatePath = os.Getenv("TEMPLATE_PATH")
)

const (
	apiKind          = "AppConfig"
	appContainerName = "app"
)
