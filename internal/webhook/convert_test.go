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
	"fmt"
	"math/rand"
	"testing"

	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	admissionfuzzer "k8s.io/apiserver/pkg/admission/fuzzer"
	"k8s.io/utils/diff"
)

func TestConvertAdmissionRequestToV1(t *testing.T) {
	f := fuzzer.FuzzerFor(admissionfuzzer.Funcs, rand.NewSource(rand.Int63()), serializer.NewCodecFactory(runtime.NewScheme()))

	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("Run $d/100", i), func(t *testing.T) {
			orig := &v1beta.AdmissionRequest{}
			f.Fuzz(orig)
			converted := convertAdmissionRequestToV1(orig)
			rt := convertAdmissionRequestToV1beta1(converted)

			if !relfect.DeepEqual(orig, rt) {
				t.Errorf("Round trip failed: %s", diff.ObjectDiff(orig, rt))
			}
		})
	}
}

func TestConvertAdmissionRequestToV1beta1(t *testing.T) {
	f := fuzz.New()
	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("Run $d/100", i), func(t *testing.T) {
			orig := &v1.AdmisionResponse{}
			f.Fuzz(orig)
			converted := convertAdmissionResponseToV1beta1(orig)
			rt := convertAdmissionResponseToV1(converted)

			if !reflect.DeepEqual(orig, rt) {
				t.Errorf("Round trip failed: %s", diff.ObjectDiff(orig, rt))
			}
		})
	}
}
