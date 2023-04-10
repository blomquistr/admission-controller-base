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
	"encoding/json"
	"strings"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// a bit of a silly function, really, but bar access to the `default` namespace
// for all requests; everyone should be in their own namespace rather than default
func AdmitNoDefault(ar v1.AdmissionReview) *v1.AdmissionResponse {
	// do some simple logging
	klog.V(2).Info("Validating object is not in the default namespace")

	// create an admission response object and
	// set the default response - we assume false
	reviewResponse := v1.AdmissionResponse{
		Allowed: false,
	}

	// create a new stub of an object that we're going to populate
	// from the request object body
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}

	// the raw variable is the raw String serialization of the
	// object under review by the admission controller
	raw := ar.Request.Object.Raw
	// using json.Unmarshal, we will unmarshal that raw string
	// into the obj struct; we'll drop any values that
	err := json.Unmarshal(raw, &obj)
	// if there is an error, fail the request and pass it to a
	// special handler to convert an Error golang object into
	// a Kubernetes V1AdmissionResponse
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	if strings.ToLower(obj.ObjectMeta.Namespace) == "default" {
		reviewResponse.Result = &metav1.Status{Message: "default namespace not allowed!"}
	} else {
		reviewResponse.Allowed = true
	}
	return &reviewResponse
}
