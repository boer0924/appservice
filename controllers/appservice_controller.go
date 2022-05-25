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

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/util/retry"

	"github.com/boerlabs/resources"
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

// +kubebuilder:rbac:groups="",resources=pods;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
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
	logger := log.FromContext(ctx).WithValues("appservice", req.NamespacedName)

	var appService appv1beta1.AppService
	if err := r.Get(ctx, req.NamespacedName, &appService); err != nil {
		// AppService被删除时忽略错误
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 已经获取到了 AppService 实例
	// 创建/更新对应的 Deployment, Service 以及 Ingress 对象
	// CreateOrUpdate
	// 调谐：观察当前的状态和期望的状态进行对比

	var deployment appsv1.Deployment
	deployment.Name = appService.Name
	deployment.Namespace = appService.Namespace
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r.Client, &deployment, func() error {
			resources.NewDeployment(&appService, &deployment)
			return ctrl.SetControllerReference(&appService, &deployment, r.Scheme)
		})
		logger.Info("CreateOrUpdate Result", "Deployment", or)
		// logger.Error(err, "CreateOrUpdate Error", "errorString", err.Error())
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	var service corev1.Service
	service.Name = appService.Name
	service.Namespace = appService.Namespace
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, _ := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
			resources.NewService(&appService, &service)
			return ctrl.SetControllerReference(&appService, &service, r.Scheme)
		})
		logger.Info("CreateOrUpdate Result", "Service", or)
		// @TODO spec.clusterIP: Invalid value: \"\": field is immutable"
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	}

	var ingress netv1.Ingress
	ingress.Name = appService.Name
	ingress.Namespace = appService.Namespace
	ingress.Annotations = appService.Annotations
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r.Client, &ingress, func() error {
			resources.NewIngress(&appService, &ingress)
			return ctrl.SetControllerReference(&appService, &ingress, r.Scheme)
		})
		logger.Info("CreateOrUpdate Result", "Ingress", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.AppService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&netv1.Ingress{}).
		Complete(r)
}
