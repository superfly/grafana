package services

import (
	"fmt"

	"github.com/grafana/grafana/pkg/tsdb/cloudwatch/constants"
	"github.com/grafana/grafana/pkg/tsdb/cloudwatch/models"
)

var GetHardCodedDimensionKeysByNamespace = func(namespace string) ([]string, error) {
	var dimensionKeys []string
	exists := false
	if dimensionKeys, exists = constants.NamespaceDimensionKeysMap[namespace]; !exists {
		return nil, fmt.Errorf("unable to find dimensions for namespace '%q'", namespace)
	}
	return dimensionKeys, nil
}

var GetHardCodedMetricsByNamespace = func(namespace string) ([]models.Metric, error) {
	response := []models.Metric{}
	exists := false
	var metrics []string
	if metrics, exists = constants.NamespaceMetricsMap[namespace]; !exists {
		return nil, fmt.Errorf("unable to find metrics for namespace '%q'", namespace)
	}

	for _, metric := range metrics {
		response = append(response, models.Metric{Namespace: namespace, Name: metric})
	}

	return response, nil
}

var GetAllHardCodedMetrics = func() []models.Metric {
	response := []models.Metric{}
	for namespace, metrics := range constants.NamespaceMetricsMap {
		for _, metric := range metrics {
			response = append(response, models.Metric{Namespace: namespace, Name: metric})
		}
	}

	return response
}

var GetHardCodedNamespaces = func() []string {
	var namespaces []string
	for key := range constants.NamespaceMetricsMap {
		namespaces = append(namespaces, key)
	}

	return namespaces
}
