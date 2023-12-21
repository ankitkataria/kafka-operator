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
	KAFKA_PORT int32 = 9092
)

func (r *KafkaReconciler) CreateKafka(ctx context.Context, kafka *v1alpha1.Kafka) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Creating Kafka")
	err := r.createOrUpdateKafkaDeployment(ctx, kafka)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Creating Kafka service")
	err = r.createKafkaService(ctx, kafka)

	if err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *KafkaReconciler) DeleteKafka(ctx context.Context, app *v1alpha1.Kafka) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Deleting Kafka")

	err := r.Update(ctx, app)

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing kafka %v", err)
	}

	return ctrl.Result{}, nil
}

func (r *KafkaReconciler) createOrUpdateKafkaDeployment(ctx context.Context, kafka *v1alpha1.Kafka) error {
	var depl appsv1.Deployment

	deplName := types.NamespacedName{Name: kafka.ObjectMeta.Name + "-deployment", Namespace: kafka.ObjectMeta.Namespace}

	if err := r.Get(ctx, deplName, &depl); err != nil {
		// Creation
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("unable to fetch Deployment: %v", err)
		}

		if apierrors.IsNotFound(err) {
			r.getkafkaDeployment(kafka, &depl)
			err = r.Create(ctx, &depl)

			if err != nil {
				return fmt.Errorf("unable to create Deployment: %v", err)
			}

			return nil
		}
	}

	depl.Spec.Replicas = &kafka.Spec.Size

	err := r.Update(ctx, &depl)

	if err != nil {
		return fmt.Errorf("unable to update deployment: %v", err)
	}

	return nil
}

func (r *KafkaReconciler) getkafkaDeployment(app *v1alpha1.Kafka, depl *appsv1.Deployment) {
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
							Image: "confluentinc/cp-kafka:" + app.Spec.Version,
							Ports: []v1.ContainerPort{
								{
									Name:          "http-kf",
									ContainerPort: 9092,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "KAFKA_BROKER_ID",
									Value: "1",
								},
								{
									Name:  "KAFKA_ZOOKEEPER_CONNECT",
									Value: "zookeeper-sample-service:2181",
								},
								{
									Name:  "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP",
									Value: "PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT",
								},
								{
									Name:  "KAFKA_ADVERTISED_LISTENERS",
									Value: "PLAINTEXT://localhost:29092,PLAINTEXT_INTERNAL://kafka-sample-service:9092",
								},
								{
									Name:  "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR",
									Value: "1",
								},
								{
									Name:  "KAFKA_TRANSACTION_STATE_LOG_MIN_ISR",
									Value: "1",
								},
								{
									Name:  "KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR",
									Value: "1",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *KafkaReconciler) createKafkaService(ctx context.Context, app *v1alpha1.Kafka) error {
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
					Name:       "tcp",
					Protocol:   v1.ProtocolTCP,
					Port:       KAFKA_PORT,
					TargetPort: intstr.FromInt(int(KAFKA_PORT)),
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
