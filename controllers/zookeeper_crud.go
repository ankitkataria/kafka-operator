package controllers

import (
	"context"
	"fmt"

	"github.com/kafkaop/kafka-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ZOOKEEPER_PORT      int32 = 2181
	ZOOKEEPER_TICK_PORT int32 = 2000
)

func (r *ZookeeperReconciler) CreateZookeeper(ctx context.Context, zookeeper *v1alpha1.Zookeeper) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Creating Zookeeper Service")
	err := r.createOrUpdateZookeeperDeployment(ctx, zookeeper)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating Zookeeper Service")
	err = r.createZookeeperService(ctx, zookeeper)

	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ZookeeperReconciler) DeleteZookeeper(ctx context.Context, app *v1alpha1.Zookeeper) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Deleting Zookeeper")

	err := r.Update(ctx, app)

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing zookeeper %v", err)
	}

	return ctrl.Result{}, nil
}

func (r *ZookeeperReconciler) createZookeeperService(ctx context.Context, app *v1alpha1.Zookeeper) error {
	srv := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.ObjectMeta.Name + "-service",
			Namespace: app.ObjectMeta.Namespace,
			Labels:    map[string]string{"app": app.ObjectMeta.Name},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: app.APIVersion,
					Kind:       app.Kind,
					Name:       app.Name,
					UID:        app.UID,
				},
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{"app": app.ObjectMeta.Name},
			Ports: []v1.ServicePort{
				{
					Name:       "tcp-zookeeper",
					Protocol:   v1.ProtocolTCP,
					Port:       ZOOKEEPER_PORT,
					TargetPort: intstr.FromInt(int(ZOOKEEPER_PORT)),
				},
			},
		},
		Status: v1.ServiceStatus{},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &srv, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create Service: %v", err)
	}
	return nil
}

func (r *ZookeeperReconciler) createOrUpdateZookeeperDeployment(ctx context.Context, zookeeper *v1alpha1.Zookeeper) error {
	var depl appsv1.Deployment

	deplName := types.NamespacedName{Name: zookeeper.ObjectMeta.Name + "-deployment", Namespace: zookeeper.ObjectMeta.Namespace}

	if err := r.Get(ctx, deplName, &depl); err != nil {
		// Creation
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("unable to fetch Deployment: %v", err)
		}

		if apierrors.IsNotFound(err) {
			r.getZookeeperDeployment(zookeeper, &depl)
			err = r.Create(ctx, &depl)

			if err != nil {
				return fmt.Errorf("unable to create Deployment: %v", err)
			}

			return nil
		}
	}

	depl.Spec.Replicas = &zookeeper.Spec.Size

	err := r.Update(ctx, &depl)

	if err != nil {
		return fmt.Errorf("unable to update deployment: %v", err)
	}

	return nil
}

func (r *ZookeeperReconciler) getZookeeperDeployment(app *v1alpha1.Zookeeper, depl *appsv1.Deployment) {
	*depl = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        app.ObjectMeta.Name + "-deployment",
			Namespace:   app.ObjectMeta.Namespace,
			Labels:      map[string]string{"label": app.ObjectMeta.Name, "app": app.ObjectMeta.Name},
			Annotations: map[string]string{"imageregistry": "https://hub.docker.com/"},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: app.APIVersion,
					Kind:       app.Kind,
					Name:       app.Name,
					UID:        app.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &app.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"label": app.ObjectMeta.Name},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"label": app.ObjectMeta.Name, "app": app.ObjectMeta.Name},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  app.ObjectMeta.Name + "-container",
							Image: "confluentinc/cp-zookeeper:" + app.Spec.Version,
							Ports: []v1.ContainerPort{
								{
									Name:          "http-zk",
									ContainerPort: ZOOKEEPER_PORT,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "ZOOKEEPER_CLIENT_PORT",
									Value: fmt.Sprintf("%d", ZOOKEEPER_PORT),
								},
								{
									Name:  "ZOOKEEPER_TICK_TIME",
									Value: fmt.Sprintf("%d", ZOOKEEPER_TICK_PORT),
								},
							},
						},
					},
				},
			},
		},
	}
}
