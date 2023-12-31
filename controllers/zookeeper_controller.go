/*
Copyright 2023.

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kafkaop/kafka-operator/api/v1alpha1"
	kafkav1alpha1 "github.com/kafkaop/kafka-operator/api/v1alpha1"
)

//+kubebuilder:rbac:groups=kafka.kafkaop.com,resources=zookeepers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka.kafkaop.com,resources=zookeepers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafka.kafkaop.com,resources=zookeepers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Zookeeper object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ZookeeperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var zookeeper v1alpha1.Zookeeper

	if err := r.Get(ctx, req.NamespacedName, &zookeeper); err != nil {
		if apierrors.IsNotFound(err) {
			// Ignoring not found errors
			// Need a new creation request
			return ctrl.Result{}, nil
		}

		l.Error(err, "unable to fetch Zookeeper Application")
		return ctrl.Result{}, err
	}

	if !zookeeper.DeletionTimestamp.IsZero() {
		l.Info("Zookeeper is being deleted")
		return r.DeleteZookeeper(ctx, &zookeeper)
	}

	l.Info("Zookeeper is being created")
	return r.CreateZookeeper(ctx, &zookeeper)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZookeeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkav1alpha1.Zookeeper{}).
		Complete(r)
}
