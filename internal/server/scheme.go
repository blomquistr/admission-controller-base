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

// Copied from https://github.com/kubernetes/kubernetes/blob/release-1.26/test/images/agnhost/webhook/scheme.go
package server

/*
comment via Github Copilot
1. The init() function is a special function that is called when the program starts. It is used to initialize the package.
2. The runtime.NewScheme() function creates a new runtime.Scheme object.
3. The serializer.NewCodecFactory() function creates a new serializer.CodecFactory object. The CodecFactory object is used to create a decoder.
4. The decoder is then used to decode the admission review request and response.
*/
import (
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	addToScheme(scheme)
}

/*
comment via Github Copilot
1. The function addToScheme() is a helper function that adds the following APIs to the scheme:

The corev1 package contains the core API objects.
The admissionv1 package contains the admission.k8s.io/v1 API objects.
The admissionv1beta1 package contains the admission.k8s.io/v1beta1 API objects.
The admissionregistrationv1 package contains the admissionregistration.k8s.io/v1 API objects.
The admissionregistrationv1beta1 package contains the admissionregistration.k8s.io/v1beta1 API objects.

2. The utilruntime.Must() function is a wrapper for the panic() function that allows you to pass the error object as an argument.
*/
func addToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(admissionv1.AddToScheme(scheme))
	utilruntime.Must(admissionv1beta1.AddToScheme(scheme))
	utilruntime.Must(admissionregistrationv1.AddToScheme(scheme))
	utilruntime.Must(admissionregistrationv1beta1.AddToScheme(scheme))
}
