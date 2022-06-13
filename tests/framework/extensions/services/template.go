package services

import (
	corev1 "k8s.io/api/core/v1"
)

// NewServiceTemplate is a constructor that creates the service template for services
func NewServiceTemplate(serviceType corev1.ServiceType, ports []corev1.ServicePort, selector map[string]string) corev1.ServiceSpec {
	return corev1.ServiceSpec{
		Type:     serviceType,
		Ports:    ports,
		Selector: selector,
	}
}
