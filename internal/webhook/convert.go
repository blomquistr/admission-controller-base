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
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
	Converts a v1beta1 Admission Request to a v1 Admission Request - handy translation

stolen from the API tests of Kubernetes, but something that would've been useful on
certain versions of k8s and good to remember for versioning APIs in the future; do stuff
like this to make your developers' lives easier
*/
func convertAdmissionRequestToV1(r *v1beta1.AdmissionRequest) *v1.AdmissionRequest {
	return &v1.AdmissionRequest{
		Kind:               r.Kind,
		Namespace:          r.Namespace,
		Name:               r.Name,
		Object:             r.Object,
		Resource:           r.Resource,
		Operation:          v1.Operation(r.Operation),
		UID:                r.UID,
		DryRun:             r.DryRun,
		OldObject:          r.OldObject,
		Options:            r.Options,
		RequestKind:        r.RequestKind,
		RequestResource:    r.RequestResource,
		RequestSubResource: r.RequestSubResource,
		SubResource:        r.SubResource,
		UserInfo:           r.UserInfo,
	}
}

/*
	Converts a v1 admission request to a v1beta1 admission request - like the previous method

this isn't strictly a necessary functionality anymore in k8s because the v1beta1 API was
removed several versions ago, but it's a good pattern to implement because it keeps the idea
in mind for future use cases where we're maintaining versioned APIs
*/
func convertAdmissionRequestToV1beta1(r *v1.AdmissionRequest) *v1beta1.AdmissionRequest {
	return &v1beta1.AdmissionRequest{
		Kind:               r.Kind,
		Namespace:          r.Namespace,
		Name:               r.Name,
		Object:             r.Object,
		Resource:           r.Resource,
		Operation:          v1beta1.Operation(r.Operation),
		UID:                r.UID,
		DryRun:             r.DryRun,
		OldObject:          r.OldObject,
		Options:            r.Options,
		RequestKind:        r.RequestKind,
		RequestResource:    r.RequestResource,
		RequestSubResource: r.RequestSubResource,
		SubResource:        r.SubResource,
		UserInfo:           r.UserInfo,
	}
}

/*
	Converts a v1 admission response to a v1beta1 admission response - this *also* won't come

up anymore with the removal of the v1beta1 API, but (I feel like a broken record repeating this)
it's a good pattern that bears reinforcement in this exercise.
*/
func convertAdmissionResponseToV1beta1(r *v1.AdmissionResponse) *v1beta1.AdmissionResponse {
	var pt *v1beta1.PatchType
	if r.PatchType != nil {
		t := v1beta1.PatchType(*r.PatchType)
		pt = &t
	}
	return &v1beta1.AdmissionResponse{
		UID:              r.UID,
		Allowed:          r.Allowed,
		AuditAnnotations: r.AuditAnnotations,
		Patch:            r.Patch,
		PatchType:        pt,
		Result:           r.Result,
		Warnings:         r.Warnings,
	}
}

/*
	A method to convert an error into a v1AdmissionResponse object to respond to Kubernetes

even when we encounter an error. This is something the Python webhooks that we have now should
do.
*/
func toV1AdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
