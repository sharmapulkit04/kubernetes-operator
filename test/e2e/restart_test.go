package e2e

import (
	"context"
	"time"

	"github.com/jenkinsci/kubernetes-operator/api/v1alpha2"
	jenkinsclient "github.com/jenkinsci/kubernetes-operator/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func configureAuthorizationToUnSecure(namespace, configMapName string) {
	limitRange := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"set-unsecured-authorization.groovy": `
import hudson.security.*

def jenkins = jenkins.model.Jenkins.getInstance()

def strategy = new AuthorizationStrategy.Unsecured()
jenkins.setAuthorizationStrategy(strategy)
jenkins.save()
`,
		},
	}

	Expect(k8sClient.Create(context.TODO(), limitRange)).Should(Succeed())
}

func checkIfAuthorizationStrategyUnsecuredIsSet(jenkinsClient jenkinsclient.Jenkins) {
	By("checking if Authorization Strategy Unsecured is set")

	logs, err := jenkinsClient.ExecuteScript(`
	import hudson.security.*

	def jenkins = jenkins.model.Jenkins.getInstance()

	if (!(jenkins.getAuthorizationStrategy() instanceof AuthorizationStrategy.Unsecured)) {
	  throw new Exception('AuthorizationStrategy.Unsecured is not set')
	}
	`)
	Expect(err).NotTo(HaveOccurred(), logs)
}

func checkBaseConfigurationCompleteTimeIsNotSet(jenkins *v1alpha2.Jenkins) {
	By("checking that Base Configuration's complete time is not set")
	/*
		jenkinsStatus := &v1alpha2.Jenkins{}
		namespaceName := types.NamespacedName{Namespace: jenkins.Namespace, Name: jenkins.Name}
		err := k8sClient.Get(context.TODO(), namespaceName, jenkinsStatus)
		Expect(err).NotTo(HaveOccurred())
		if jenkinsStatus.Status.BaseConfigurationCompletedTime != nil {
			Fail(fmt.Sprintf("Status.BaseConfigurationCompletedTime is set after pod restart, status %+v", jenkinsStatus.Status))
		}
	*/

	Eventually(func() (bool, error) {
		actualJenkins := &v1alpha2.Jenkins{}
		err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: jenkins.Name, Namespace: jenkins.Namespace}, actualJenkins)
		if err != nil {
			return false, err
		}
		return actualJenkins.Status.BaseConfigurationCompletedTime == nil, nil
	}, time.Duration(110)*retryInterval, time.Second).Should(BeTrue())
	//	_, _ = fmt.Fprintf(GinkgoWriter, "Jenkins instance is up and ready\n")
}
