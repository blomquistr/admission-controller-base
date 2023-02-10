package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"

	webhook "github.com/blomquistr/admission-controller-base/internal/webhook"
)

/* Method to handle the HTTP portion of requests from the API server, before it hands
the request off to the admitHandler `admit` */
func serve(w http.ResponseWriter, r *http.Request, admit webhook.AdmitHandler) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType) // TODO: clean this up with a better log message, as if this weren't an acceptance test
		return
	}

	klog.V(2).Info(fmt.Sprintf("Handling request [%s]", body))

	/* finding this was a pain, because the example in the k8s tests uses
	an object from their tests library instead of using the client-go
	package, which is what an actual client would use.
	EXAMPLE: https://dx13.co.uk/articles/2021/01/15/kubernetes-types-using-go/

	This doesn't load any custom resource types; this must be done manually.
	EXAMPLE: https://github.com/bitnami-labs/sealed-secrets/blob/ce399099886139edbfcb7a16f3c693a62dbe9475/pkg/apis/sealed-secrets/v1alpha1/register.go#L32-L39 */

	/* The UniversalDeserializer is an object that decodes byte code into
	Go objects to be operated on by the rest of your application. It loads
	all the known client-go object types for Kubernetes, but won't work on
	custom resources */
	deserializer := scheme.Codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	/* This section of the server constructs a response object based on the group,
	version, and Kind of the request - we'll send a v1 response if the request is
	a v1 request, and a v1beta1 response if the request is a v1beta1 request. In
	either case, responseObj will be valid for the request that's been submitted.
	
	Our response object is the admit handler passed to the initial method call. It
	must return a valid response object, formatted appropriately for either the v1
	or v1beta1 APIs, in jsonpatch format. We'll call the Struct's v1 endpoint in
	the v1 path, or the struct's v1beta1 endpoint in the v1beta1 path.*/
	var responseObj runtime.Object
	switch *gvk {
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1beta1.AdmissionReview but got [%T]", obj)
			return
		}
		responseAdmissionReview := &v1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.V1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1.AdmissionReview but got: %T", obj)
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.V1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	/* Finally, we'll encode our response  */
	klog.V(2).Info(fmt.Sprintf("sending response: [%v]", responseObj))
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}
