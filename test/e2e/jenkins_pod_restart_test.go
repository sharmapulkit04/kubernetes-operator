package e2e

import (
	"github.com/jenkinsci/kubernetes-operator/api/v1alpha2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Jenkins controller", func() {

	const (
		jenkinsCRName     = e2e
		priorityClassName = ""
	)

	var (
		namespace     *corev1.Namespace
		jenkins       *v1alpha2.Jenkins
		groovyScripts = v1alpha2.GroovyScripts{
			Customization: v1alpha2.Customization{
				Configurations: []v1alpha2.ConfigMapRef{},
			},
		}
		casc = v1alpha2.ConfigurationAsCode{
			Customization: v1alpha2.Customization{
				Configurations: []v1alpha2.ConfigMapRef{},
			},
		}
		LivenessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/login",
					Port:   intstr.FromString("http"),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: int32(80),
			TimeoutSeconds:      int32(4),
			FailureThreshold:    int32(10),
			SuccessThreshold:    int32(1),
			PeriodSeconds:       int32(1),
		}
	)

	BeforeEach(func() {
		namespace = createNamespace()

		configureAuthorizationToUnSecure(namespace.Name, userConfigurationConfigMapName)
		jenkins = createJenkinsCR(jenkinsCRName, namespace.Name, nil, groovyScripts, casc, LivenessProbe, priorityClassName)
	})

	AfterEach(func() {
		destroyNamespace(namespace)
	})

	Context("when restarting Jenkins master pod", func() {
		It("new Jenkins Master pod should be created", func() {
			waitForJenkinsBaseConfigurationToComplete(jenkins)
			restartJenkinsMasterPod(jenkins)
			waitForRecreateJenkinsMasterPod(jenkins)
			checkBaseConfigurationCompleteTimeIsNotSet(jenkins)
			waitForJenkinsBaseConfigurationToComplete(jenkins)
		})
	})
})

var _ = Describe("Jenkins controller", func() {

	const (
		jenkinsCRName     = e2e
		priorityClassName = ""
	)

	var (
		namespace     *corev1.Namespace
		jenkins       *v1alpha2.Jenkins
		groovyScripts = v1alpha2.GroovyScripts{
			Customization: v1alpha2.Customization{
				Configurations: []v1alpha2.ConfigMapRef{
					{
						Name: userConfigurationConfigMapName,
					},
				},
			},
		}
		casc = v1alpha2.ConfigurationAsCode{
			Customization: v1alpha2.Customization{
				Configurations: []v1alpha2.ConfigMapRef{},
			},
		}
		LivenessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/login",
					Port:   intstr.FromString("http"),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: int32(100),
			TimeoutSeconds:      int32(4),
			FailureThreshold:    int32(12),
			SuccessThreshold:    int32(1),
			PeriodSeconds:       int32(5),
		}
	)

	BeforeEach(func() {
		namespace = createNamespace()

		configureAuthorizationToUnSecure(namespace.Name, userConfigurationConfigMapName)
		jenkins = createJenkinsCR(jenkinsCRName, namespace.Name, nil, groovyScripts, casc, LivenessProbe, priorityClassName)
	})

	AfterEach(func() {
		destroyNamespace(namespace)
	})

	Context("when running Jenkins safe restart", func() {
		It("authorization strategy is not overwritten", func() {
			waitForJenkinsBaseConfigurationToComplete(jenkins)
			waitForJenkinsUserConfigurationToComplete(jenkins)
			jenkinsClient, cleanUpFunc := verifyJenkinsAPIConnection(jenkins, namespace.Name)
			defer cleanUpFunc()
			checkIfAuthorizationStrategyUnsecuredIsSet(jenkinsClient)

			err := jenkinsClient.SafeRestart()
			Expect(err).NotTo(HaveOccurred())
			waitForJenkinsSafeRestart(jenkinsClient)

			checkIfAuthorizationStrategyUnsecuredIsSet(jenkinsClient)
		})
	})
})
