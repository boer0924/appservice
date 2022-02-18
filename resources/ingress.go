package resources

import (
	appv1beta1 "github.com/boerlabs/api/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewIngress(app *appv1beta1.AppService) *netv1.Ingress {
	// pathType := "app.PathType" // TODO
	return &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        app.Name,
			Namespace:   app.Namespace,
			Annotations: app.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, schema.GroupVersionKind{
					Group:   appv1beta1.GroupVersion.Group,
					Version: appv1beta1.GroupVersion.Version,
					Kind:    app.Kind,
				}),
			},
		},
		Spec: netv1.IngressSpec{
			Rules: app.Spec.Rules,
		},
		// Spec: netv1.IngressSpec{
		// 	Rules: []netv1.IngressRule{
		// 		{
		// 			Host: "app.HostName", // TODO
		// 			IngressRuleValue: netv1.IngressRuleValue{
		// 				HTTP: &netv1.HTTPIngressRuleValue{
		// 					Paths: []netv1.HTTPIngressPath{
		// 						{
		// 							Path: "app.Path", // TODO
		// 							PathType: (*netv1.PathType)(&pathType), // TODO
		// 							Backend: netv1.IngressBackend{
		// 								Service: &netv1.IngressServiceBackend{
		// 									Name: app.Name,
		// 									Port: netv1.ServiceBackendPort{
		// 										Name:   app.Name,
		// 										Number: app.Spec.Ports[0].Port,
		// 									},
		// 								},
		// 							},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}
}
