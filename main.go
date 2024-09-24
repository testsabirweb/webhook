package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func main() {
	var err error

	// Initialize the Kubernetes clientset once
	clientset, err = initializeClientset()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/mutate", handleMutate)

	if err := r.RunTLS(":8080", "/etc/certs/tls.crt", "/etc/certs/tls.key"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server started on https://localhost:8080/")
}

func initializeClientset() (*kubernetes.Clientset, error) {
	// Use the in-cluster config to connect to Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
	}

	// Create a new Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return clientset, nil
}

func handleMutate(c *gin.Context) {
	labels := getLabelsFromConfigMap(clientset)
	log.Printf("labels fetched from configmap %v", labels)
	admissionReview := v1.AdmissionReview{}
	var err error

	if err = c.BindJSON(&admissionReview); err != nil {
		log.Printf("Failed to bind AdmissionReview: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid AdmissionReview payload"})
		return
	}

	admissionReviewReq := admissionReview.Request
	if admissionReviewReq == nil {
		log.Println("AdmissionReview request is nil")
		c.JSON(http.StatusBadRequest, gin.H{"error": "AdmissionReview request is missing"})
		return
	}

	log.Printf("Incoming request UID: %s, Namespace: %s, Name: %s", admissionReviewReq.UID, admissionReviewReq.Namespace, admissionReviewReq.Name)

	var pod corev1.Pod
	if err = json.Unmarshal(admissionReviewReq.Object.Raw, &pod); err != nil {
		log.Printf("Failed to unmarshal pod object: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal pod object"})
		return
	}

	response := v1.AdmissionResponse{
		UID: admissionReviewReq.UID,
	}

	patch, err := addLabel(&pod, labels)
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  "Failure",
			Message: fmt.Sprintf("Error while adding label: %v", err),
		}
		log.Printf("Error while adding label: %v", err)
	} else {
		response.Allowed = true
		response.PatchType = func(pt v1.PatchType) *v1.PatchType { return &pt }(v1.PatchTypeJSONPatch)
		response.Patch = patch
		response.Result = &metav1.Status{
			Status: "Success",
		}
		log.Printf("Label added successfully to pod %s in namespace %s", pod.Name, pod.Namespace)
	}

	admissionReview.Response = &response
	c.JSON(http.StatusOK, admissionReview)
}

func addLabel(pod *corev1.Pod, labels map[string]string) ([]byte, error) {
	var patch []map[string]interface{}

	for key, value := range labels {
		patch = append(patch, map[string]interface{}{
			"op":    "add",
			"path":  fmt.Sprintf("/metadata/labels/%s", key),
			"value": value,
		})
	}

	log.Printf("Adding labels to pod %s", pod.Name)

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}
	return patchBytes, nil
}

func getLabelsFromConfigMap(clientset *kubernetes.Clientset) map[string]string {
	// Get the ConfigMap from the 'default' namespace with name 'label-config'
	configMap, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "label-config", metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get ConfigMap: %v", err)
	}
	return configMap.Data
}
