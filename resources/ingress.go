package resources

import (
	appv1beta1 "github.com/boerlabs/api/v1beta1"
	netv1 "k8s.io/api/networking/v1"
)

func NewIngress(app *appv1beta1.AppService, ingress *netv1.Ingress) {
	// ingress.ObjectMeta = metav1.ObjectMeta{
	// 	Name:        app.Name,
	// 	Namespace:   app.Namespace,
	// 	Annotations: app.Annotations,
	// }
	ingress.Spec = netv1.IngressSpec{
		Rules: app.Spec.Rules,
	}
}
