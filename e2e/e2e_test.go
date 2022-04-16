package e2e_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	cfg *rest.Config
)

var _ = BeforeSuite(func() {
	configfile := os.Getenv("KUBECONFIG")
	if configfile == "" {
		configfile = "$HOME/.kube/config"
	}
	restConfig, err := clientcmd.BuildConfigFromFlags("", os.ExpandEnv(configfile))
	Expect(err).ShouldNot(HaveOccurred())

	cfg = restConfig

	client, err := kubernetes.NewForConfig(restConfig)
	Expect(err).ShouldNot(HaveOccurred())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err = waitUntilReady(ctx, client)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = Describe("E2E", func() {
	Describe("Run example job", func() {
		AfterEach(func() {
			client, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				panic(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			err = cleanup(ctx, client)
			if err != nil {
				panic(err)
			}
		})
		Context("Cleanup job when all", func() {
			It("Job succeeded", func() {
				cmd := exec.Command("../kube-job", "run", "--template-file", "../example/job.yaml", "--args", "pwd", "--container", "alpine")
				out, err := cmd.Output()
				Expect(err).To(BeNil())
				Expect(string(out)).To(Equal("/\n"))

				time.Sleep(10 * time.Second)
				client, err := kubernetes.NewForConfig(cfg)
				Expect(err).To(BeNil())
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
					LabelSelector: "app=example-job",
				})
				Expect(err).To(BeNil())
				Expect(len(jobList.Items)).To(Equal(0))
			})
			It("Job failed", func() {
				cmd := exec.Command("../kube-job", "run", "--template-file", "../example/job.yaml", "--args", "hoge", "--container", "alpine")
				_, err := cmd.Output()
				Expect(err).To(HaveOccurred())

				time.Sleep(10 * time.Second)
				client, err := kubernetes.NewForConfig(cfg)
				Expect(err).To(BeNil())
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
					LabelSelector: "app=example-job",
				})
				Expect(err).To(BeNil())
				Expect(len(jobList.Items)).To(Equal(0))
			})
		})
		Context("Cleanup job only succeeded", func() {
			It("Job succeeded", func() {
				cmd := exec.Command("../kube-job", "run", "--template-file", "../example/job.yaml", "--args", "pwd", "--container", "alpine", "--cleanup", "succeeded")
				out, err := cmd.Output()
				Expect(err).To(BeNil())
				Expect(string(out)).To(Equal("/\n"))

				time.Sleep(10 * time.Second)
				client, err := kubernetes.NewForConfig(cfg)
				Expect(err).To(BeNil())
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
					LabelSelector: "app=example-job",
				})
				Expect(err).To(BeNil())
				Expect(len(jobList.Items)).To(Equal(0))
			})
			It("Job failed", func() {
				cmd := exec.Command("../kube-job", "run", "--template-file", "../example/job.yaml", "--args", "hoge", "--container", "alpine", "--cleanup", "succeeded")
				_, err := cmd.Output()
				Expect(err).To(HaveOccurred())

				time.Sleep(10 * time.Second)
				client, err := kubernetes.NewForConfig(cfg)
				Expect(err).To(BeNil())
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
					LabelSelector: "app=example-job",
				})
				Expect(err).To(BeNil())
				Expect(len(jobList.Items)).To(Equal(1))
			})
		})
		Context("Cleanup job only failed", func() {
			It("Job succeeded", func() {
				cmd := exec.Command("../kube-job", "run", "--template-file", "../example/job.yaml", "--args", "pwd", "--container", "alpine", "--cleanup", "failed")
				out, err := cmd.Output()
				Expect(err).To(BeNil())
				Expect(string(out)).To(Equal("/\n"))

				time.Sleep(10 * time.Second)
				client, err := kubernetes.NewForConfig(cfg)
				Expect(err).To(BeNil())
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
					LabelSelector: "app=example-job",
				})
				Expect(err).To(BeNil())
				Expect(len(jobList.Items)).To(Equal(1))
			})
			It("Job failed", func() {
				cmd := exec.Command("../kube-job", "run", "--template-file", "../example/job.yaml", "--args", "hoge", "--container", "alpine", "--cleanup", "failed")
				_, err := cmd.Output()
				Expect(err).To(HaveOccurred())

				time.Sleep(10 * time.Second)
				client, err := kubernetes.NewForConfig(cfg)
				Expect(err).To(BeNil())
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
					LabelSelector: "app=example-job",
				})
				Expect(err).To(BeNil())
				Expect(len(jobList.Items)).To(Equal(0))
			})
		})
	})
})

func waitUntilReady(ctx context.Context, client *kubernetes.Clientset) error {
	klog.Info("Waiting until kubernetes cluster is ready")
	err := wait.Poll(10*time.Second, 10*time.Minute, func() (bool, error) {
		nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to list nodes: %v", err)
		}
		if len(nodeList.Items) == 0 {
			klog.Warningf("node does not exist yet")
			return false, nil
		}
		for i := range nodeList.Items {
			n := &nodeList.Items[i]
			if !nodeIsReady(n) {
				klog.Warningf("node %s is not ready yet", n.Name)
				return false, nil
			}
		}
		klog.Info("all nodes are ready")
		return true, nil
	})
	return err
}

func nodeIsReady(node *corev1.Node) bool {
	for i := range node.Status.Conditions {
		con := &node.Status.Conditions[i]
		if con.Type == corev1.NodeReady && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func cleanup(ctx context.Context, client *kubernetes.Clientset) error {
	klog.Info("Waiting until all jobs are deleted")
	if err := cleanupJobs(ctx, client); err != nil {
		return err
	}
	return cleanupPods(ctx, client)
}

func cleanupJobs(ctx context.Context, client *kubernetes.Clientset) error {
	return wait.PollImmediate(3*time.Second, 1*time.Minute, func() (bool, error) {
		jobList, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
			LabelSelector: "app=example-job",
		})
		if err != nil && kerrors.IsNotFound(err) {
			klog.Info("jobs do not found")
			return true, nil
		} else if err != nil {
			klog.Errorf("failed to list pods: %v", err)
			return false, err
		}
		if len(jobList.Items) == 0 {
			klog.Info("all jobs are deleted")
			return true, nil
		}
		for i := range jobList.Items {
			job := &jobList.Items[i]
			klog.Infof("Job is still living, so deleting %s/%s", job.Namespace, job.Name)

			err := client.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.Errorf("failed to delete job: %v", err)
				return false, nil
			}
		}
		return false, nil
	})
}

func cleanupPods(ctx context.Context, client *kubernetes.Clientset) error {
	return wait.PollImmediate(3*time.Second, 1*time.Minute, func() (bool, error) {
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
			LabelSelector: "app=example",
		})
		if err != nil && kerrors.IsNotFound(err) {
			klog.Info("pods do not found")
			return true, nil
		} else if err != nil {
			klog.Errorf("failed to list pods: %v", err)
			return false, err
		}
		if len(podList.Items) == 0 {
			klog.Info("all pods are deleted")
			return true, nil
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			klog.Infof("Pod is still living %s/%s, so deleting", pod.Namespace, pod.Name)
			err := client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.Errorf("failed to delete pod: %v", err)
				return false, nil
			}
		}
		return false, nil
	})
}
