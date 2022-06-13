package ingresses

import (
	networkingv1 "k8s.io/api/networking/v1"
)

// NewIngressTemplate is a constructor that creates the ingress template for ingresses
func NewIngressTemplate(hostName string, paths []networkingv1.HTTPIngressPath) networkingv1.IngressSpec {
	return networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				Host: hostName,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: paths,
					},
				},
			},
		},
	}
}
