/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Sourced from https://github.com/kubernetes/kubernetes/blob/release-1.26/test/images/agnhost/webhook/scheme.go
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
