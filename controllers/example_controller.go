/*
Copyright 2021.

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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	examplev1 "mmertdogann/example-operator/api/v1"
)

// ExampleReconciler reconciles a Example object
type ExampleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=example.example.com,resources=examples,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.example.com,resources=examples/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=example.example.com,resources=examples/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Example object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ExampleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	lg := log.FromContext(ctx)
	lg.Info("Enter Reconcile", "req", req)

	example := &examplev1.Example{}

	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, example); err != nil {
		lg.Error(err, "Unable to fetch Example resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	result, err := r.checkDeployment(lg, example, r.nginxDeployment(example))
	if result != nil {
		return *result, err
	}

	return ctrl.Result{}, nil
}

func (r *ExampleReconciler) checkDeployment(logger logr.Logger,
	instance *examplev1.Example,
	dep *appsv1.Deployment,
) (*reconcile.Result, error) {

	found := &appsv1.Deployment{}

	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)

	if err != nil && errors.IsNotFound(err) {

		logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		instance.Status.Name = dep.Name
		err = r.Client.Create(context.TODO(), dep)

		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		}
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return &ctrl.Result{}, err
	}

	return &ctrl.Result{}, nil

}

func (r *ExampleReconciler) nginxDeployment(cr *examplev1.Example) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "default",
			Name:            cr.Spec.Name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cr, examplev1.GroupVersion.WithKind("Example"))},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &(cr.Spec.Size),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.Spec.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Labels: map[string]string{
						"app": cr.Spec.Name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  cr.Spec.Name,
							Image: "nginx:alpine",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *ExampleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.Example{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
