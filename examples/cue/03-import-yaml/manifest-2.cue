package podinfo

KubernetesResources: horizontalpodautoscaler: podinfo: deploymentEnv: common: {
	apiVersion: "autoscaling/v2beta2"
	kind:       "HorizontalPodAutoscaler"
	metadata: name: "podinfo"
	spec: {
		maxReplicas: 4
		metrics: [{
			resource: {
				name: "cpu"
				target: {
					averageUtilization: 99
					type:               "Utilization"
				}
			}
			type: "Resource"
		}]
		minReplicas: 2
		scaleTargetRef: {
			apiVersion: "apps/v1"
			kind:       "Deployment"
			name:       "podinfo"
		}
	}
}
