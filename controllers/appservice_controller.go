/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"reflect"

	netv1 "k8s.io/api/networking/v1"

	"github.com/boerlabs/resources"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1beta1 "github.com/boerlabs/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.boer.xyz,resources=appservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.boer.xyz,resources=appservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.boer.xyz,resources=appservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *AppServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here
	var appService appv1beta1.AppService
	err := r.Get(ctx, req.NamespacedName, &appService)
	if err != nil {
		// AppService被删除时忽略错误
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	logger.Info("fetch appservice objects", "appservice", appService)

	// 如果不存在，则创建关联资源
	// 如果存在，判断是否需要更新
	// 如果需要更新，则直接更新
	// 如果不需要更新，则正常返回

	deploy := &appsv1.Deployment{}
	if err = r.Get(ctx, req.NamespacedName, deploy); err != nil && errors.IsNotFound(err) {
		// 1、创建Deployment
		deploy := resources.NewDeploy(&appService)
		if err = r.Create(ctx, deploy); err != nil {
			return ctrl.Result{}, err
		}
		// 2、创建Service
		service := resources.NewService(&appService)
		if err := r.Create(ctx, service); err != nil {
			return ctrl.Result{}, err
		}
		// 3、创建Ingress
		logger.Info("debug for ->", "rules", appService.Spec.Rules)
		if appService.Spec.Rules != nil {
			ingress := resources.NewIngress(&appService)
			if err = r.Create(ctx, ingress); err != nil {
				return ctrl.Result{}, err
			}
		}
		// 4、关联Annotations
		data, _ := json.Marshal(appService.Spec)
		if appService.Annotations != nil {
			appService.Annotations["spec"] = string(data)
		} else {
			appService.Annotations = map[string]string{"spec": string(data)}
		}

		if err = r.Update(ctx, &appService); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	oldspec := appv1beta1.AppServiceSpec{}
	if err := json.Unmarshal([]byte(appService.Annotations["spec"]), &oldspec); err != nil {
		return ctrl.Result{}, err
	}

	// 当前规范与旧的对象不一致，则需要更新
	if !reflect.DeepEqual(appService.Spec, oldspec) {
		// 更新关联资源
		newDeploy := resources.NewDeploy(&appService)
		oldDeploy := &appsv1.Deployment{}
		if err := r.Get(ctx, req.NamespacedName, oldDeploy); err != nil {
			return ctrl.Result{}, err
		}
		oldDeploy.Spec = newDeploy.Spec
		if err := r.Update(ctx, oldDeploy); err != nil {
			return ctrl.Result{}, err
		}

		newService := resources.NewService(&appService)
		oldService := &corev1.Service{}
		if err := r.Get(ctx, req.NamespacedName, oldService); err != nil {
			return ctrl.Result{}, err
		}
		// 需要指定 ClusterIP 为之前的，不然更新会报错
		newService.Spec.ClusterIP = oldService.Spec.ClusterIP
		oldService.Spec = newService.Spec
		if err := r.Update(ctx, oldService); err != nil {
			return ctrl.Result{}, err
		}

		newIngress := resources.NewIngress(&appService)
		oldIngress := &netv1.Ingress{}
		if err := r.Get(ctx, req.NamespacedName, oldIngress); err != nil {
			// 因rules,omitempty策略，创建时无ingress, 更新时有ingress
			if errors.IsNotFound(err) {
				if appService.Spec.Rules != nil {
					ingress := resources.NewIngress(&appService)
					if err = r.Create(ctx, ingress); err != nil {
						return ctrl.Result{}, err
					}
				}
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		oldIngress.Spec = newIngress.Spec
		if err := r.Update(ctx, oldIngress); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil

	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.AppService{}).
		Complete(r)
}
