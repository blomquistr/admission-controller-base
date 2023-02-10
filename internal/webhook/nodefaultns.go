/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.

 * This file is based on the links found at https://github.com/kubernetes/kubernetes/tree/release-1.21/test/images/agnhost/webhook
 * and is meant as a basic "I'm learning how to Go via mutating admission controller" type project - according to the Kubernetes
 * docs at https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#write-an-admission-webhook-server
 * this is a working and valid example of how to make a Golang webserver that handles admission control requests, and can serve
 * as a model for picking this activity up. This is **not** production ready, this is **not** something you can plug and play.
 */

package webhook

import (
	"strings"
	"encoding/json"
	"k8s.io/api/admission/v1"
	"k8s.io/klog/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdmitNoDefault(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("Validating object is not in the default namespace")

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	obj := struct{
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}
	
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	if strings.ToLower(obj.ObjectMeta.Namespace) == "default" {
		reviewResponse.Allowed = false
		reviewResponse.Result = &metav1.Status{Message: "default namespace not allowed!"}
	}
	return &reviewResponse
}
