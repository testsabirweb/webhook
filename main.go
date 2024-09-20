package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
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

func handleMutate(c *gin.Context) {
	admissionReview := v1.AdmissionReview{}
	var err error

	// Error handling for JSON binding
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

	// Add label to pod
	patch, err := addLabel(&pod)
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

func addLabel(pod *corev1.Pod) ([]byte, error) {
	var patch []map[string]interface{}

	// Get existing labels from the Pod
	labels := pod.GetLabels()
	_, isCustomLabelPresent := labels["custom_label"]

	// If custom_label is not present, add it
	if !isCustomLabelPresent {
		patch = append(patch, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/custom_label",
			"value": "custom_value",
		})
		log.Printf("Adding custom_label to pod %s", pod.Name)
	} else {
		log.Printf("custom_label already exists on pod %s", pod.Name)
	}

	// Convert the patch to JSON
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}
	return patchBytes, nil
}
