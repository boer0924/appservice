package resources

import (
	appv1beta1 "github.com/boerlabs/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func NewService(app *appv1beta1.AppService, svc *corev1.Service) {
	svc.Spec = corev1.ServiceSpec{
		Type:  corev1.ServiceTypeClusterIP,
		Ports: app.Spec.Ports,
		Selector: map[string]string{
			"app": app.Name,
		},
	}
}
