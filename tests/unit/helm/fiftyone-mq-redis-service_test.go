//go:build kubeall || helm || unit || unitFiftyoneMqRedisService
// +build kubeall helm unit unitFiftyoneMqRedisService

package unit

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
)

type fiftyoneMqRedisServiceTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestFiftyoneMqRedisServiceTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &fiftyoneMqRedisServiceTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates: []string{
			"templates/fiftyone-mq-redis-service.yaml",
		},
	})
}

// renderService renders the Service template and unmarshals it into the typed
// core/v1 struct. (mqRedisEnabled lives in
// fiftyone-mq-redis-deployment_test.go — both files share the `unit` and
// `helm` tags.)
func (s *fiftyoneMqRedisServiceTemplateTest) renderService(options *helm.Options) corev1.Service {
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	var svc corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &svc)
	return svc
}

func (s *fiftyoneMqRedisServiceTemplateTest) TestMetadata() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	svc := s.renderService(options)

	expectedName := fmt.Sprintf("%s-fiftyone-mq-redis", s.releaseName)
	s.Equal(expectedName, svc.ObjectMeta.Name, "Service name should be release-prefixed")
	s.True(strings.HasPrefix(svc.ObjectMeta.Name, s.releaseName+"-"),
		"Service name must carry the release-name prefix — the queue URL helper builds "+
			"FIFTYONE_MQ_REDIS_URL from this name, so two releases in one namespace must not collide")
	s.Equal("fiftyone-teams", svc.ObjectMeta.Namespace,
		"Service namespace should default to fiftyone-teams")
	s.Equal("fiftyone-mq-redis", svc.ObjectMeta.Labels["app.kubernetes.io/name"])
	s.Equal(s.releaseName, svc.ObjectMeta.Labels["app.kubernetes.io/instance"])
	s.Equal("fiftyone-mq-redis", svc.ObjectMeta.Labels["app.voxel51.com/component"])
}

func (s *fiftyoneMqRedisServiceTemplateTest) TestNamespaceOverride() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"namespace.name": "my-ns",
	})}

	svc := s.renderService(options)
	s.Equal("my-ns", svc.ObjectMeta.Namespace)
}

// TestMqDisabled ensures the bundled Redis Service is not rendered when
// fiftyone-mq is off, even with the redis sub-toggle on.
func (s *fiftyoneMqRedisServiceTemplateTest) TestMqDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"fiftyoneMq.enabled":       "false",
		"fiftyoneMq.redis.enabled": "true",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// TestRedisDisabled ensures the redis sub-toggle alone suppresses the bundled
// Service while fiftyone-mq stays enabled.
func (s *fiftyoneMqRedisServiceTemplateTest) TestRedisDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"fiftyoneMq.enabled":       "true",
		"fiftyoneMq.redis.enabled": "false",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// TestExternalUrlSkipsBundled ensures that pointing fiftyone-mq at an
// operator-managed Redis suppresses the bundled Service. Rendering it anyway
// would publish a ClusterIP with no matching pods.
func (s *fiftyoneMqRedisServiceTemplateTest) TestExternalUrlSkipsBundled() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.external.url": "redis://my-managed-redis:6379/0",
	})}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

func (s *fiftyoneMqRedisServiceTemplateTest) TestPorts() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	svc := s.renderService(options)

	s.Require().Len(svc.Spec.Ports, 1, "Service should expose exactly one port")
	s.EqualValues(6379, svc.Spec.Ports[0].Port)
	s.Equal(6379, svc.Spec.Ports[0].TargetPort.IntValue())
}

// TestSelectorMatchesDeploymentPodLabels ensures the Service selects the
// bundled Redis pods. The selector must be the selectorLabels only — adding
// the full label set (which includes chart/version labels) would break the
// Service across chart upgrades.
func (s *fiftyoneMqRedisServiceTemplateTest) TestSelectorMatchesDeploymentPodLabels() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	svc := s.renderService(options)

	s.Equal("fiftyone-mq-redis", svc.Spec.Selector["app.kubernetes.io/name"])
	s.Equal(s.releaseName, svc.Spec.Selector["app.kubernetes.io/instance"])
	s.Len(svc.Spec.Selector, 2,
		"Service selector should be the selectorLabels only, not the full label set")
}
