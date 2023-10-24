package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	appv1 "sanmuyan.com/app-operator/api/v1"
)

func addCreatedByLabel(labels map[string]string) {
	labels["app.kubernetes.io/created-by"] = "app-operator"
}

func getDeployStatus(t appv1.DeployType, status []appv1.DeployStatus) (appv1.DeployStatus, bool) {
	for _, s := range status {
		if s.Type == t {
			return s, true
		}
	}
	return appv1.DeployStatus{}, false
}

func getCondition(t appsv1.DeploymentConditionType, dcs []appsv1.DeploymentCondition) (appsv1.DeploymentCondition, bool) {
	for _, s := range dcs {
		if s.Type == t {
			return s, true
		}
	}
	return appsv1.DeploymentCondition{}, false
}

func getContainer(n string, cs []corev1.Container) (corev1.Container, bool) {
	for _, s := range cs {
		if s.Name == n {
			return s, true
		}
	}
	return corev1.Container{}, false
}

func setContainerImage(n, image string, cs []corev1.Container) {
	for i, s := range cs {
		if s.Name == n {
			cs[i].Image = image
		}
	}
}
