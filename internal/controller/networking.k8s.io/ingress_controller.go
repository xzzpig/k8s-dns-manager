/*
Copyright 2023.

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

package networkingk8sio

import (
	"context"
	"reflect"
	"strings"
	"time"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/config"
	"github.com/xzzpig/k8s-dns-manager/pkg/generator"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	netv1 "k8s.io/api/networking/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var ingress netv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &ingress); err != nil {
		logger.Info("unable to fetch Ingress")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	ctx = context.WithValue(ctx, generator.ContextKeyIngress, &ingress)

	if !ingress.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	showResult := func(reason string, message string, err error) {
		if err != nil {
			logger.Error(err, message)
			r.recorder.Eventf(&ingress, "Warning", reason, "%s: %s", message, err.Error())
		} else {
			logger.Info(message)
			r.recorder.Event(&ingress, "Normal", reason, message)
		}
	}
	ctx = context.WithValue(ctx, generator.ContextKeyShowResultFunc, showResult)

	generatorType, ok := ingress.Annotations[generator.AnnotationKeyGenerator]
	if !ok {
		generatorType = config.GetConfig().Default.Generator.Type
	}
	if generatorType == "" {
		logger.V(1).Info("ingress ignore")
		return ctrl.Result{}, nil
	}
	logger = logger.WithValues("generator", generatorType)

	recordGenerator := generator.Get(generatorType)
	if recordGenerator == nil {
		showResult("Warning", "no generator found", nil)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	if !recordGenerator.Support(generator.DNSGeneratorSourceIngress) {
		showResult("Warning", "generator not support ingress", nil)
		return ctrl.Result{}, nil
	}

	records, err := recordGenerator.Generate(ctx, generator.DNSGeneratorSourceIngress)
	if err != nil {
		showResult("Error", "generator error", err)
		return ctrl.Result{}, err
	}

	var recordList dnsv1.DNSRecordList
	if err := r.List(ctx, &recordList, client.InNamespace(req.Namespace), client.MatchingFields{dnsOwnerKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}
	ownedRecordMap := make(map[string]dnsv1.DNSRecord)
	for _, record := range recordList.Items {
		ownedRecordMap[record.Spec.Name] = record
	}

	ingressAnnotationMap := make(map[string]string)
	for k, v := range ingress.Annotations {
		if strings.HasPrefix(k, generator.AnnotationKeyRecordPrefix) {
			ingressAnnotationMap[k] = v
		}
	}
	ingressLabelMap := make(map[string]string)
	for k, v := range ingress.Labels {
		if strings.HasPrefix(k, generator.AnnotationKeyRecordPrefix) {
			ingressLabelMap[k] = v
		}
	}

	oldLogger := logger
	for _, record := range records {
		logger = oldLogger.WithValues("dns", record)

		delete(ownedRecordMap, record.Name)

		dnsRecord := dnsv1.DNSRecord{}
		exists := true
		if err := r.Get(ctx, client.ObjectKey{Name: record.SpinalName(), Namespace: ingress.Namespace}, &dnsRecord); err != nil {
			if client.IgnoreNotFound(err) != nil {
				showResult("Error", "get dns record error", err)
				return ctrl.Result{}, err
			}
			exists = false
		}
		specEquals := reflect.DeepEqual(dnsRecord.Spec, record)
		annotationEquals := reflect.DeepEqual(dnsRecord.Annotations, ingressAnnotationMap)
		labelEquals := reflect.DeepEqual(dnsRecord.Labels, ingressLabelMap)
		if specEquals && annotationEquals && labelEquals {
			continue
		}
		dnsRecord.Spec = record
		dnsRecord.Name = record.SpinalName()
		dnsRecord.Namespace = ingress.Namespace
		dnsRecord.Annotations = ingressAnnotationMap
		dnsRecord.Labels = ingressLabelMap
		if err := ctrl.SetControllerReference(&ingress, &dnsRecord, r.Scheme); err != nil {
			showResult("Error", "set controller reference error", err)
			return ctrl.Result{}, err
		}
		if exists {
			if err := r.Update(ctx, &dnsRecord); err != nil {
				showResult("Error", "update dns record error", err)
				return ctrl.Result{}, err
			}
			showResult("Normal", "dns record updated", nil)
		} else {
			if err := r.Create(ctx, &dnsRecord); err != nil {
				showResult("Error", "create dns record error", err)
				return ctrl.Result{}, err
			}
		}
	}
	for _, record := range ownedRecordMap {
		logger = oldLogger.WithValues("dns", record)
		if err := r.Delete(ctx, &record); err != nil {
			showResult("Error", "delete dns record error", err)
			return ctrl.Result{}, err
		}
		showResult("Normal", "dns record deleted", nil)
	}
	logger = oldLogger

	logger.Info("ingress reconciled")

	return ctrl.Result{
		RequeueAfter: recordGenerator.RequeueAfter(ctx, generator.DNSGeneratorSourceIngress),
	}, nil
}

var (
	dnsOwnerKey = ".metadata.controller"
	apiGVStr    = netv1.SchemeGroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("IngressDNS")

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &dnsv1.DNSRecord{}, dnsOwnerKey, func(rawObj client.Object) []string {
		// grab the dnsRecord object, extract the owner...
		dnsRecord := rawObj.(*dnsv1.DNSRecord)
		owner := metav1.GetControllerOf(dnsRecord)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Ingress...
		if owner.APIVersion != apiGVStr || owner.Kind != "Ingress" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Ingress{}).
		Owns(&dnsv1.DNSRecord{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldGeneration := e.ObjectOld.GetGeneration()
				newGeneration := e.ObjectNew.GetGeneration()
				oldAnnotations := e.ObjectOld.GetAnnotations()
				newAnnotations := e.ObjectNew.GetAnnotations()
				oldLables := e.ObjectOld.GetLabels()
				newLables := e.ObjectNew.GetLabels()
				// Generation is only updated on spec changes (also on deletion),
				// not metadata or status
				// Filter out events where the generation hasn't changed to
				// avoid being triggered by status updates

				return oldGeneration != newGeneration ||
					!reflect.DeepEqual(oldAnnotations, newAnnotations) ||
					!reflect.DeepEqual(oldLables, newLables)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				// The reconciler adds a finalizer so we perform clean-up
				// when the delete timestamp is added
				// Suppress Delete events to avoid filtering them out in the Reconcile function
				return false
			},
		}).
		Complete(r)
}
