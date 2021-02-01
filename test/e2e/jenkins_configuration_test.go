package e2e

import (
	"fmt"

	"github.com/jenkinsci/kubernetes-operator/api/v1alpha2"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Jenkins controller configuration", func() {

	const (
		jenkinsCRName            = e2e
		numberOfExecutors        = 6
		numberOfExecutorsEnvName = "NUMBER_OF_EXECUTORS"
		systemMessage            = "Configuration as Code integration works!!!"
		systemMessageEnvName     = "SYSTEM_MESSAGE"
		priorityClassName        = ""
	)

	var (
		namespace *corev1.Namespace
		jenkins   *v1alpha2.Jenkins
		mySeedJob = seedJobConfig{
			SeedJob: v1alpha2.SeedJob{
				ID:                    "jenkins-operator",
				CredentialID:          "jenkins-operator",
				JenkinsCredentialType: v1alpha2.NoJenkinsCredentialCredentialType,
				Targets:               "cicd/jobs/*.jenkins",
				Description:           "Jenkins Operator repository",
				RepositoryBranch:      "master",
				RepositoryURL:         "https://github.com/jenkinsci/kubernetes-operator.git",
				PollSCM:               "1 1 1 1 1",
				UnstableOnDeprecation: true,
				BuildPeriodically:     "1 1 1 1 1",
				FailOnMissingPlugin:   true,
				IgnoreMissingFiles:    true,
				//AdditionalClasspath: can fail with the seed job agent
				GitHubPushTrigger: true,
			},
		}
		groovyScripts = v1alpha2.GroovyScripts{
			Customization: v1alpha2.Customization{
				Configurations: []v1alpha2.ConfigMapRef{
					{
						Name: userConfigurationConfigMapName,
					},
				},
				Secret: v1alpha2.SecretRef{
					Name: userConfigurationSecretName,
				},
			},
		}
		casc = v1alpha2.ConfigurationAsCode{
			Customization: v1alpha2.Customization{
				Configurations: []v1alpha2.ConfigMapRef{
					{
						Name: userConfigurationConfigMapName,
					},
				},
				Secret: v1alpha2.SecretRef{
					Name: userConfigurationSecretName,
				},
			},
		}
		userConfigurationSecretData = map[string]string{
			systemMessageEnvName:     systemMessage,
			numberOfExecutorsEnvName: fmt.Sprintf("%d", numberOfExecutors),
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

		createUserConfigurationSecret(namespace.Name, userConfigurationSecretData)
		createUserConfigurationConfigMap(namespace.Name, numberOfExecutorsEnvName, fmt.Sprintf("${%s}", systemMessageEnvName))
		jenkins = createJenkinsCR(jenkinsCRName, namespace.Name, &[]v1alpha2.SeedJob{mySeedJob.SeedJob}, groovyScripts, casc, LivenessProbe, priorityClassName)
		createDefaultLimitsForContainersInNamespace(namespace.Name)
		createKubernetesCredentialsProviderSecret(namespace.Name, mySeedJob)
	})

	AfterEach(func() {
		destroyNamespace(namespace)
	})

	Context("when deploying CR to cluster", func() {
		It("creates Jenkins instance and configures it", func() {
			waitForJenkinsBaseConfigurationToComplete(jenkins)
			verifyJenkinsMasterPodAttributes(jenkins)
			verifyServices(jenkins)
			jenkinsClient, cleanUpFunc := verifyJenkinsAPIConnection(jenkins, namespace.Name)
			defer cleanUpFunc()
			verifyPlugins(jenkinsClient, jenkins)
			waitForJenkinsUserConfigurationToComplete(jenkins)
			verifyUserConfiguration(jenkinsClient, numberOfExecutors, systemMessage)
			verifyJenkinsSeedJobs(jenkinsClient, []seedJobConfig{mySeedJob})
		})
	})
})

var _ = Describe("Jenkins controller priority class", func() {

	const (
		jenkinsCRName     = "k8s-ete-priority-class-existing"
		priorityClassName = "system-cluster-critical"
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
			InitialDelaySeconds: int32(100),
			TimeoutSeconds:      int32(4),
			FailureThreshold:    int32(10),
			SuccessThreshold:    int32(1),
			PeriodSeconds:       int32(1),
		}
	)

	BeforeEach(func() {
		namespace = createNamespace()
		jenkins = createJenkinsCR(jenkinsCRName, namespace.Name, nil, groovyScripts, casc, LivenessProbe, priorityClassName)
	})

	AfterEach(func() {
		destroyNamespace(namespace)
	})

	Context("when deploying CR with priority class to cluster", func() {
		It("creates Jenkins instance and configures it", func() {
			waitForJenkinsBaseConfigurationToComplete(jenkins)
			verifyJenkinsMasterPodAttributes(jenkins)
		})
	})
})

var _ = Describe("Jenkins controller plugins test", func() {

	const (
		jenkinsCRName     = e2e
		priorityClassName = ""
		jobID             = "k8s-e2e"
	)

	var (
		namespace *corev1.Namespace
		jenkins   *v1alpha2.Jenkins
		mySeedJob = seedJobConfig{
			SeedJob: v1alpha2.SeedJob{
				ID:                    "jenkins-operator",
				CredentialID:          "jenkins-operator",
				JenkinsCredentialType: v1alpha2.NoJenkinsCredentialCredentialType,
				Targets:               "cicd/jobs/k8s.jenkins",
				Description:           "Jenkins Operator repository",
				RepositoryBranch:      "master",
				RepositoryURL:         "https://github.com/jenkinsci/kubernetes-operator.git",
			},
		}
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
		jenkins = createJenkinsCR(jenkinsCRName, namespace.Name, &[]v1alpha2.SeedJob{mySeedJob.SeedJob}, groovyScripts, casc, LivenessProbe, priorityClassName)
	})

	AfterEach(func() {
		destroyNamespace(namespace)
	})

	Context("when deploying CR with a SeedJob to cluster", func() {
		It("runs kubernetes plugin job successfully", func() {
			waitForJenkinsUserConfigurationToComplete(jenkins)
			jenkinsClient, cleanUpFunc := verifyJenkinsAPIConnection(jenkins, namespace.Name)
			defer cleanUpFunc()
			waitForJobCreation(jenkinsClient, jobID)
			verifyJobCanBeRun(jenkinsClient, jobID)
			verifyJobHasBeenRunCorrectly(jenkinsClient, jobID)
		})
	})
})
