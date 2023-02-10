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
	"os"
	"strings"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var (
	cloudProvider string = os.Getenv("CLOUD_PROVIDER")
)

const (
	addAnnotationsAndAzureLabelPatch string = `[
		{ "op": "add", "path", "/metadata/annotations", "value": {"service.beta.kubernetes.io/azure-load-balancer-internal: "true"}}
	]`
	addAzureAnnotationPatch string = `[
		{ "op": "add", "path", "/metadata/annotations/service.beta.kubernetes.io/azure-load-balancer-internal", "value": "true" }
	]`
	setAzureAnnotationTruePatch string = `[
		{ "op": "replace", "path": "/metadata/labels/service.beta.kubernetes.io/azure-load-balancer-internal", "value": "true" }
	]`
)

/* This function checks that any service of type LoadBalancer has the Azure-specific annotation preventing them from
generating a public IP address. */
func NoExternalIpLoadBalancers(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("Testing load balancer for public IP addresses...")

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
	}
	pt := v1.PatchTypeJSONPatch
	var labelValue string
	hasLabel := false
	switch {
	case strings.ToLower(cloudProvider) == "azure":
		labelValue, hasLabel = obj.ObjectMeta.Annotations["service.beta.kubernetes.io/azure-load-balancer-internal"]
	default:
		err := fmt.Errorf("unexpected cloud provider configuration: got [%s], expected one of [Azure, AWS, GCP]", cloudProvider)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	switch {
	case len(obj.ObjectMeta.Annotations) == 0:
		reviewResponse.Patch = []byte(addAnnotationsAndAzureLabelPatch)
		reviewResponse.PatchType = &pt
	case !hasLabel:
		reviewResponse.Patch = []byte(addAzureAnnotationPatch)
		reviewResponse.PatchType = &pt
	case labelValue != "true":
		reviewResponse.Patch = []byte(setAzureAnnotationTruePatch)
		reviewResponse.PatchType = &pt
	default:
		// none - already set
	}
	return &reviewResponse
}
