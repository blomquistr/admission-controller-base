package webhook

import (
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
)

/* This section of the webhook server defines a struct that each admission control handler
must provide at instantiation, which will be used to route requests that are v1 and v1beta1
requests - each webhook must define a v1 compliant and v1beta1 compliant method to handle
requests, which is a clever scheme to force a new plugin to handle all possible paths of an
admission request */
type admitv1beta1Func func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse
type admitv1Func func(v1.AdmissionReview) *v1.AdmissionResponse

/* The struct that each webhook that registers with the controller must provide - notice
that it takes two kinds of funcs, a v1beta1 function and a v1 function, as properties. Each
webhook will have to define these things as part of registration. */
type AdmitHandler struct {
	V1beta1 admitv1beta1Func
	V1      admitv1Func
}

/* This method converst a v1beta1 admission request to a v1 admission request, which, again
is a slick way to achieve compliance between either type of requests with minimal handling by
the prospective end user who is just trying to write a mutate/admit method for their webhook */
func DelegateV1beta1AdmitToV1(f admitv1Func) admitv1beta1Func {
	return func(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
		in := v1.AdmissionReview{Request: convertAdmissionRequestToV1(review.Request)}
		out := f(in)
		return convertAdmissionResponseToV1beta1(out)
	}
}

func NewDelegateToV1AdmitHandler(f admitv1Func) AdmitHandler {
	return AdmitHandler{
		V1beta1: DelegateV1beta1AdmitToV1(f),
		V1:      f,
	}
}
