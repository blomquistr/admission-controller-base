package webhook

import (
	"encoding/json"
	"testing"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAdmitNoDefault(t *testing.T) {
	testCases := []struct {
		name           string
		objectMeta     map[string]string
		expectedResult bool
	}{
		{
			name:           "Default namespace meta",
			objectMeta:     map[string]string{"namespace": "default"},
			expectedResult: false,
		},
		{
			name:           "Non-default namespace meta",
			objectMeta:     map[string]string{"namespace": "not-default"},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: tc.objectMeta["namespace"]}}
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}

			review := v1.AdmissionReview{Request: &v1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
			response := AdmitNoDefault(review)
			if response.Allowed != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, response.Allowed)
			}
		})
	}
}
