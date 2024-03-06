package unit

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"gopkg.in/yaml.v3"
)

func chartInfo(t *testing.T, chartPath string) (map[string]interface{}, error) {
	// Get chart info.
	output, err := helm.RunHelmCommandAndGetOutputE(t, &helm.Options{}, "show", "chart", chartPath)
	if err != nil {
		return nil, err
	}
	cInfo := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(output), &cInfo)
	return cInfo, err
}
