package webhook

import (
	"encoding/json"
	"reflect"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNoExternalIpLoadBalancers(t *testing.T) {
	testCases := []struct {
		name                string
		initialAnnotations  map[string]string
		expectedAnnotations map[string]string
	}{
		{
			name:                "Azure Load Balancer with public IP",
			initialAnnotations:  map[string]string{},
			expectedAnnotations: map[string]string{"service.beta.kubernetes.io/azure-load-balancer-internal": "true"},
		},
		{
			name:                "Azure Load Balancer with false annotation value",
			initialAnnotations:  map[string]string{"service.beta.kubernetes.io/azure-load-balancer-internal": "false"},
			expectedAnnotations: map[string]string{"service.beta.kubernetes.io/azure-load-balancer-internal": "true"},
		},
		{
			name:                "Azure Load Balancer with true annotation value",
			initialAnnotations:  map[string]string{"service.beta.kubernetes.io/azure-load-balancer-internal": "true"},
			expectedAnnotations: map[string]string{"service.beta.kubernetes.io/azure-load-balancer-internal": "true"},
		},
		{
			name:                "AWS Load Balancer with public IP",
			initialAnnotations:  map[string]string{},
			expectedAnnotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal"},
		},
		{
			name:                "AWS Load Balancer with false annotation value",
			initialAnnotations:  map[string]string{"service.beta.kubernetes.io/aws-load-balancer-scheme": "false"},
			expectedAnnotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal"},
		},
		{
			name:                "AWS Load Balancer with true annotation value",
			initialAnnotations:  map[string]string{"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal"},
			expectedAnnotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal"},
		},
		{
			name:                "GCP Load Balancer with public IP",
			initialAnnotations:  map[string]string{},
			expectedAnnotations: map[string]string{"cloud.google.com/load-balancer-type": "internal"},
		},
		{
			name:                "GCP Load Balancer with false annotation value",
			initialAnnotations:  map[string]string{"cloud.google.com/load-balancer-type": "false"},
			expectedAnnotations: map[string]string{"cloud.google.com/load-balancer-type": "internal"},
		},
		{
			name:                "GCP Load Balancer with true annotation value",
			initialAnnotations:  map[string]string{"cloud.google.com/load-balancer-type": "internal"},
			expectedAnnotations: map[string]string{"cloud.google.com/load-balancer-type": "internal"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := corev1.Service{ObjectMeta: metav1.ObjectMeta{Annotations: tc.initialAnnotations}}
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}

			review := v1.AdmissionReview{Request: &v1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
			response := NoExternalIpLoadBalancers(review)
			if response.Patch != nil {
				patchObj, err := jsonpatch.DecodePatch(response.Patch)
				if err != nil {
					t.Fatal(err)
				}
				raw, err := patchObj.Apply(raw)
				if err != nil {
					t.Fatal(err)
				}
			}

			objType := reflect.TypeOf(request)
			objTest := reflect.New(objType).Interface()
			err = json.Unmarshal(raw, objTest)
			if err != nil {
				t.Fatal(err)
			}

			actual := objTest.(*corev1.Service)
			if !reflect.DeepEqual(actual.Annotations, tc.expectedAnnotations) {
				t.Errorf("expected %v, got %v", tc.expectedAnnotations, actual.Annotations)
			}
		})
	}
}
