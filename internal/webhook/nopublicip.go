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
	"fmt"
	"strings"

	"github.com/spf13/viper"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	// AWS annotations and labels
	awsInternalLoadBalancerAnnotation      string = "service.beta.kubernetes.io/aws-load-balancer-scheme"
	awsInternalLoadBalancerAnnotationValue string = "internal"
	addAnnotationsAndAWSLabelPatch         string = `[
		{ "op": "add", "path", "/metadata/annotations", "value": {"service.beta.kubernetes.io/aws-load-balancer-scheme: "internal"}}
	]`
	addAWSAnnotationPatch string = `[
		{ "op": "add", "path", "/metadata/annotations/service.beta.kubernetes.io/aws-load-balancer-scheme", "value": "internal" }
	]`
	setAWSAnnotationTruePatch string = `[
		{ "op": "replace", "path": "/metadata/labels/service.beta.kubernetes.io/aws-load-balancer-scheme", "value": "internal" }
	]`

	// Azure annotations and labels
	azureInternalLoadBalancerAnnotation      string = "service.beta.kubernetes.io/azure-load-balancer-internal"
	azureInternalLoadBalancerAnnotationValue string = "true"
	addAnnotationsAndAzureLabelPatch         string = `[
		{ "op": "add", "path", "/metadata/annotations", "value": {"service.beta.kubernetes.io/azure-load-balancer-internal: "true"}}
	]`
	addAzureAnnotationPatch string = `[
		{ "op": "add", "path", "/metadata/annotations/service.beta.kubernetes.io/azure-load-balancer-internal", "value": "true" }
	]`
	setAzureAnnotationTruePatch string = `[
		{ "op": "replace", "path": "/metadata/labels/service.beta.kubernetes.io/azure-load-balancer-internal", "value": "true" }
	]`

	// GCP/GKE annotations and labels
	gcpInternalLoadBalancerAnnotation      string = "networking.gke.io/load-balancer-type"
	gcpInternalLoadBalancerAnnotationValue string = "internal"
	addAnnotationsAndGCPLabelPatch         string = `[
		{ "op": "add", "path", "/metadata/annotations", "value": {"networking.gke.io/load-balancer-type: "internal"}}
	]`
	addGCPAnnotationPatch string = `[
		{ "op": "add", "path", "/metadata/annotations/networking.gke.io/load-balancer-type", "value": "internal" }
	]`
	setGCPAnnotationTruePatch string = `[
		{ "op": "replace", "path": "/metadata/labels/networking.gke.io/load-balancer-type", "value": "internal" }
	]`
)

type cloudProviderAnnotation struct {
	Annotation             string
	AnnotationValue        string
	AddAnnotationsAndPatch string
	AddAnnotationPatch     string
	SetAnnotationPatch     string
}

var (
	// cloudProvider         string                  = os.Getenv("CLOUD_PROVIDER")
	cloudProvider         string                  = viper.GetString("cloudProvider")
	annotationsCollection cloudProviderAnnotation = newCloudProviderAnnotation()
)

func newCloudProviderAnnotation() cloudProviderAnnotation {
	annotationsCollection := cloudProviderAnnotation{}
	switch {
	case strings.ToLower(cloudProvider) == "aws":
		annotationsCollection.Annotation = awsInternalLoadBalancerAnnotation
		annotationsCollection.AnnotationValue = awsInternalLoadBalancerAnnotationValue
		annotationsCollection.AddAnnotationsAndPatch = addAnnotationsAndAWSLabelPatch
		annotationsCollection.AddAnnotationPatch = addAWSAnnotationPatch
		annotationsCollection.SetAnnotationPatch = setAWSAnnotationTruePatch
	case strings.ToLower(cloudProvider) == "azure":
		annotationsCollection.Annotation = azureInternalLoadBalancerAnnotation
		annotationsCollection.AnnotationValue = azureInternalLoadBalancerAnnotationValue
		annotationsCollection.AddAnnotationsAndPatch = addAnnotationsAndAzureLabelPatch
		annotationsCollection.AddAnnotationPatch = addAzureAnnotationPatch
		annotationsCollection.SetAnnotationPatch = setAzureAnnotationTruePatch
	case strings.ToLower(cloudProvider) == "gcp":
		annotationsCollection.Annotation = gcpInternalLoadBalancerAnnotation
		annotationsCollection.AnnotationValue = gcpInternalLoadBalancerAnnotationValue
		annotationsCollection.AddAnnotationsAndPatch = addAnnotationsAndGCPLabelPatch
		annotationsCollection.AddAnnotationPatch = addGCPAnnotationPatch
		annotationsCollection.SetAnnotationPatch = setGCPAnnotationTruePatch
	default:
		err := fmt.Errorf("unexpected cloud provider configuration: got [%s], expected one of [Azure, AWS, GCP]", cloudProvider)
		klog.Error(err)
	}
	return annotationsCollection
}

/*
	This function checks that any service of type LoadBalancer has the Azure-specific annotation preventing them from

generating a public IP address.
*/
func NoExternalIpLoadBalancers(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("Testing load balancer for public IP addresses...")

	// on each request, check that we're configured with a valid cloud provider and, if not, return an error
	if len(annotationsCollection.Annotation) == 0 {
		err := fmt.Errorf("unexpected cloud provider configuration: got [%s], expected one of [Azure, AWS, GCP]", cloudProvider)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	reviewResponse := v1.AdmissionResponse{
		Allowed: true,
	}

	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
	}
	pt := v1.PatchTypeJSONPatch
	labelValue, hasLabel := obj.ObjectMeta.Annotations[annotationsCollection.Annotation]
	switch {
	case len(obj.ObjectMeta.Annotations) == 0:
		reviewResponse.Patch = []byte(annotationsCollection.AddAnnotationsAndPatch)
		reviewResponse.PatchType = &pt
	case !hasLabel:
		reviewResponse.Patch = []byte(annotationsCollection.AddAnnotationPatch)
		reviewResponse.PatchType = &pt
	case labelValue != annotationsCollection.AnnotationValue:
		reviewResponse.Patch = []byte(annotationsCollection.SetAnnotationPatch)
		reviewResponse.PatchType = &pt
	default:
		// none - already set
	}
	return &reviewResponse
}
