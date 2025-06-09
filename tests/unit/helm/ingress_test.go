//go:build kubeall || helm || unit || unitIngress
// +build kubeall helm unit unitIngress

package unit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	networkingv1 "k8s.io/api/networking/v1"
)

type ingressTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestIngressTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &ingressTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/ingress.yaml"},
	})
}

func (s *ingressTemplateTest) TestDisabled() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-test-fiftyone-teams-app",
		},
		{
			"defaultValuesIngressDisabled",
			map[string]string{
				"ingress.enabled": "false",
			},
			"",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/ingress.yaml in chart")

				var ingress networkingv1.Ingress
				helm.UnmarshalK8SYaml(subT, output, &ingress)

				s.Empty(ingress.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var ingress networkingv1.Ingress
				helm.UnmarshalK8SYaml(subT, output, &ingress)

				s.Equal(testCase.expected, ingress.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

// TODO: Unit Test with different k8s versions
// Given kubernetes version  >=1.14 and <1.19-0, when ingress is enabled then `apiVersion: networking.k8s.io/v1beta1`
// Given kubernetes version  <1.14 , when ingress is enabled then `apiVersion: extensions/v1beta1`
func (s *ingressTemplateTest) TestApiVersion() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"networking.k8s.io/v1",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			s.Equal(testCase.expected, ingress.TypeMeta.APIVersion, "API Version should be equal")
		})
	}
}

func (s *ingressTemplateTest) TestMetadataAnnotations() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"overrideAnnotations",
			map[string]string{
				"ingress.annotations.annotation-1": "annotation-1-value",
			},
			map[string]string{
				"annotation-1": "annotation-1-value",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			if testCase.expected == nil {
				s.Nil(ingress.ObjectMeta.Annotations, "Annotations should be nil")
			} else {
				for key, value := range testCase.expected {
					foundValue := ingress.ObjectMeta.Annotations[key]
					s.Equal(value, foundValue, "Annotations should contain all set annotations.")
				}
			}
		})
	}
}

func (s *ingressTemplateTest) TestMetadataLabels() {
	// Get chart info (to later obtain the chart's appVersion)
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	// Get appVersion from chart info
	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

	// Get version from chart info
	chartVersion, exists := cInfo["version"]
	s.True(exists, "failed to get version from chart info")

	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "fiftyone-test-fiftyone-teams-app",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
		{
			"overrideIngressLabels",
			map[string]string{
				"ingress.labels.test-label-key": "test-label-value",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "fiftyone-test-fiftyone-teams-app",
				"app.kubernetes.io/instance":   "fiftyone-test",
				"test-label-key":               "test-label-value",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			for key, value := range testCase.expected {
				foundValue := ingress.ObjectMeta.Labels[key]
				s.Equal(value, foundValue, "Labels should contain all set labels.")
			}
		})
	}
}

// .Chart.Name = "fiftyone-teams-app"
// .Release.Name = "fiftyone-test"
func (s *ingressTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-test-fiftyone-teams-app",
		},
		{
			"overrideFullnameOverride",
			map[string]string{
				"fullnameOverride": "test-service-name",
			},
			"test-service-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			s.Equal(testCase.expected, ingress.ObjectMeta.Name, "Ingress name should be equal.")
		})
	}
}

func (s *ingressTemplateTest) TestMetadataNamespace() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"namespace.name": "test-namespace-name",
			},
			"test-namespace-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			s.Equal(testCase.expected, ingress.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *ingressTemplateTest) TestIngressClassName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			"overrideIngressClassName",
			map[string]string{
				"ingress.className": "nginx",
			},
			"nginx",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			if testCase.expected == "" {
				s.Nil(ingress.Spec.IngressClassName)
			} else {
				s.Equal(testCase.expected, *ingress.Spec.IngressClassName, "Ingress class name should be equal.")
			}
		})
	}
}

func (s *ingressTemplateTest) TestTls() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(tls []networkingv1.IngressTLS)
	}{
		{
			"defaultValues",
			nil,
			func(tls []networkingv1.IngressTLS) {
				expectedJSON := `[
          {
            "hosts": [
              ""
            ],
            "secretName": "fiftyone-teams-tls-secret"
          }
        ]`
				var expectedTls []networkingv1.IngressTLS
				err := json.Unmarshal([]byte(expectedJSON), &expectedTls)
				s.NoError(err)
				s.Equal(expectedTls, tls, "TLS should be equal")
			},
		},
		{
			"overrideTlsEnabled",
			map[string]string{
				"ingress.tlsEnabled": "false",
			},
			func(tls []networkingv1.IngressTLS) {
				s.Nil(tls, "TLS should be nil")
			},
		},
		{
			"overrideDnsNamesAndTlsSecretName",
			map[string]string{
				"apiSettings.dnsName":      "teams-api.fiftyone.ai",
				"teamsAppSettings.dnsName": "teams-app.fiftyone.ai",
				"ingress.tlsSecretName":    "test-secret",
			},
			func(tls []networkingv1.IngressTLS) {
				expectedJSON := `[
          {
            "hosts": [
              "teams-app.fiftyone.ai",
              "teams-api.fiftyone.ai"
            ],
            "secretName": "test-secret"
          }
        ]`
				var expectedTls []networkingv1.IngressTLS
				err := json.Unmarshal([]byte(expectedJSON), &expectedTls)
				s.NoError(err)
				s.Equal(expectedTls, tls, "TLS should be equal")
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			testCase.expected(ingress.Spec.TLS)
		})
	}
}

func (s *ingressTemplateTest) TestRules() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(tls []networkingv1.IngressRule)
	}{
		{
			"defaultValues",
			nil,
			func(tls []networkingv1.IngressRule) {
				expectedJSON := `[
          {
            "host": "",
            "http": {
              "paths": [
                {
                  "path": "/cas",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-cas",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/*",
                  "pathType": "ImplementationSpecific",
                  "backend": {
                    "service": {
                      "name": "teams-app",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                }
              ]
            }
          }
        ]`
				var expectedRules []networkingv1.IngressRule
				err := json.Unmarshal([]byte(expectedJSON), &expectedRules)
				s.NoError(err)
				s.Equal(expectedRules, tls, "Rules should be equal")
			},
		},
		{
			"overridePaths",
			map[string]string{
				"ingress.paths[0].path":        "/test-cas",
				"ingress.paths[0].pathType":    "ImplementationSpecific",
				"ingress.paths[0].serviceName": "test-teams-cas",
				"ingress.paths[0].servicePort": "81",
				"ingress.paths[1].path":        "/**",
				"ingress.paths[1].pathType":    "Prefix",
				"ingress.paths[1].serviceName": "test-teams-app",
				"ingress.paths[1].servicePort": "82",
			},
			func(tls []networkingv1.IngressRule) {
				expectedJSON := `[
          {
            "host": "",
            "http": {
              "paths": [
                {
                  "path": "/test-cas",
                  "pathType": "ImplementationSpecific",
                  "backend": {
                    "service": {
                      "name": "test-teams-cas",
                      "port": {
                        "number": 81
                      }
                    }
                  }
                },
                {
                  "path": "/**",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "test-teams-app",
                      "port": {
                        "number": 82
                      }
                    }
                  }
                }
              ]
            }
          }
        ]`
				var expectedRules []networkingv1.IngressRule
				err := json.Unmarshal([]byte(expectedJSON), &expectedRules)
				s.NoError(err)
				s.Equal(expectedRules, tls, "Rules should be equal")
			},
		},
		{
			"overrideTeamsAppSettingsDnsName",
			map[string]string{
				"teamsAppSettings.dnsName": "teams-app.fiftyone.ai",
			},
			func(tls []networkingv1.IngressRule) {
				expectedJSON := `[
          {
            "host": "teams-app.fiftyone.ai",
            "http": {
              "paths": [
                {
                  "path": "/cas",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-cas",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/*",
                  "pathType": "ImplementationSpecific",
                  "backend": {
                    "service": {
                      "name": "teams-app",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                }
              ]
            }
          }
        ]`
				var expectedRules []networkingv1.IngressRule
				err := json.Unmarshal([]byte(expectedJSON), &expectedRules)
				s.NoError(err)
				s.Equal(expectedRules, tls, "Rules should be equal")
			},
		},
		{
			"overrideApiSettingsDnsName",
			map[string]string{
				"apiSettings.dnsName": "teams-api.fiftyone.ai",
			},
			func(tls []networkingv1.IngressRule) {
				expectedJSON := `[
          {
            "host": "",
            "http": {
              "paths": [
                {
                  "path": "/cas",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-cas",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/*",
                  "pathType": "ImplementationSpecific",
                  "backend": {
                    "service": {
                      "name": "teams-app",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                }
              ]
            }
          },
          {
            "host": "teams-api.fiftyone.ai",
            "http": {
              "paths": [
                {
                  "path": "/*",
                  "pathType": "ImplementationSpecific",
                  "backend": {
                    "service": {
                      "name": "teams-api",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                }
              ]
            }
          }
        ]`
				var expectedRules []networkingv1.IngressRule
				err := json.Unmarshal([]byte(expectedJSON), &expectedRules)
				s.NoError(err)
				s.Equal(expectedRules, tls, "Rules should be equal")
			},
		},
		{
			"overrideBothDnsNames",
			map[string]string{
				"apiSettings.dnsName":      "teams-api.fiftyone.ai",
				"apiSettings.service.name": "test-service-name-teams-api",
				"apiSettings.service.port": "81",
				"ingress.api.path":         "/test-api-path",
				"ingress.api.pathType":     "prefix",
				"teamsAppSettings.dnsName": "teams-app.fiftyone.ai",
			},
			func(tls []networkingv1.IngressRule) {
				expectedJSON := `[
          {
            "host": "teams-app.fiftyone.ai",
            "http": {
              "paths": [
                {
                  "path": "/cas",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-cas",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/*",
                  "pathType": "ImplementationSpecific",
                  "backend": {
                    "service": {
                      "name": "teams-app",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                }
              ]
            }
          },
          {
            "host": "teams-api.fiftyone.ai",
            "http": {
              "paths": [
                {
                  "path": "/test-api-path",
                  "pathType": "prefix",
                  "backend": {
                    "service": {
                      "name": "test-service-name-teams-api",
                      "port": {
                        "number": 81
                      }
                    }
                  }
                }
              ]
            }
          }
        ]`
				var expectedRules []networkingv1.IngressRule
				err := json.Unmarshal([]byte(expectedJSON), &expectedRules)
				s.NoError(err)
				s.Equal(expectedRules, tls, "Rules should be equal")
			},
		},
		{
			"overridePathsWithPathBasedRouting",
			map[string]string{
				"teamsAppSettings.dnsName":     "teams-app.fiftyone.ai",
				"ingress.paths[0].path":        "/cas",
				"ingress.paths[0].pathType":    "Prefix",
				"ingress.paths[0].serviceName": "teams-cas",
				"ingress.paths[0].servicePort": "80",
				"ingress.paths[1].path":        "/_pymongo",
				"ingress.paths[1].pathType":    "Prefix",
				"ingress.paths[1].serviceName": "teams-api",
				"ingress.paths[1].servicePort": "80",
				"ingress.paths[2].path":        "/health",
				"ingress.paths[2].pathType":    "Prefix",
				"ingress.paths[2].serviceName": "teams-api",
				"ingress.paths[2].servicePort": "80",
				"ingress.paths[3].path":        "/graphql/v1",
				"ingress.paths[3].pathType":    "Prefix",
				"ingress.paths[3].serviceName": "teams-api",
				"ingress.paths[3].servicePort": "80",
				"ingress.paths[4].path":        "/file",
				"ingress.paths[4].pathType":    "Prefix",
				"ingress.paths[4].serviceName": "teams-api",
				"ingress.paths[4].servicePort": "80",
				"ingress.paths[5].path":        "/",
				"ingress.paths[5].pathType":    "Prefix",
				"ingress.paths[5].serviceName": "teams-app",
				"ingress.paths[5].servicePort": "80",
			},
			func(tls []networkingv1.IngressRule) {
				expectedJSON := `[
          {
            "host": "teams-app.fiftyone.ai",
            "http": {
              "paths": [
                {
                  "path": "/cas",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-cas",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/_pymongo",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-api",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/health",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-api",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/graphql/v1",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-api",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/file",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-api",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                },
                {
                  "path": "/",
                  "pathType": "Prefix",
                  "backend": {
                    "service": {
                      "name": "teams-app",
                      "port": {
                        "number": 80
                      }
                    }
                  }
                }
              ]
            }
          }
        ]`
				var expectedRules []networkingv1.IngressRule
				err := json.Unmarshal([]byte(expectedJSON), &expectedRules)
				s.NoError(err)
				s.Equal(expectedRules, tls, "Rules should be equal")
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(subT, output, &ingress)

			testCase.expected(ingress.Spec.Rules)
		})
	}
}
