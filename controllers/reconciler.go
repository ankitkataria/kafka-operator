package controllers

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KafkaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type ZookeeperReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}
