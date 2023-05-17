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

package dns

import (
	"context"
	"errors"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/config"
	"github.com/xzzpig/k8s-dns-manager/pkg/provider"
	"github.com/xzzpig/k8s-dns-manager/util"
)

// DNSRecordReconciler reconciles a DNSRecord object
type DNSRecordReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsrecords,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsrecords/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsrecords/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DNSRecord object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DNSRecordReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dnsRecord dnsv1.DNSRecord
	if err := r.Get(ctx, req.NamespacedName, &dnsRecord); err != nil {
		logger.Error(err, "unable to fetch DNSRecord")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	dnsRecordOrigin := dnsRecord.DeepCopy()

	logger = logger.WithValues("dnsrecord", dnsRecord.Name)

	status := &dnsRecord.Status
	defer func() {
		if !reflect.DeepEqual(dnsRecordOrigin.Status, dnsRecord.Status) {
			if err := r.Status().Update(ctx, &dnsRecord); err != nil && dnsRecord.DeletionTimestamp.IsZero() {
				// logger.Error(err, "unable to update DNSRecord status")
				status = status.DeepCopy()
				if err := r.Get(ctx, req.NamespacedName, &dnsRecord); err != nil {
					logger.Error(err, "unable to fetch DNSRecord for update status")
				}
				dnsRecord.Status = *status
				if err := r.Status().Update(ctx, &dnsRecord); err != nil {
					logger.Error(err, "unable to update DNSRecord status")
				}
			}
		}
	}()

	showResult := func(message string, err error) {
		if err != nil {
			logger.Error(err, message)
			status.Message = message + err.Error()
			r.recorder.Event(&dnsRecord, "Warning", "Error", message+err.Error())
		} else {
			logger.Info(message)
			status.Message = message
			r.recorder.Event(&dnsRecord, "Normal", "Info", message)
		}
	}

	if status.Status == "" {
		status.Status = dnsv1.DNSRecordStatusPhasePending
		showResult("start reconciling", nil)
		return ctrl.Result{Requeue: true}, nil
	}

	if status.ProviderRef.Name == "" {
		if status.Status != dnsv1.DNSRecordStatusPhaseMatching {
			status.Status = dnsv1.DNSRecordStatusPhaseMatching
			showResult("start matching provider for "+dnsRecord.Spec.Name, nil)
			return ctrl.Result{Requeue: true}, nil
		}
		var providerList dnsv1.DNSProviderList
		if err := r.List(ctx, &providerList); err != nil {
			showResult("unable to list DNSProvider", err)
			return ctrl.Result{}, err
		}
		for _, provider := range providerList.Items {
			if dnsRecord.Match(&provider.Spec) {
				status.ProviderRef.Name = provider.Name
				status.ProviderRef.Namespace = provider.Namespace
				showResult("provider found for "+dnsRecord.Spec.Name+" provider: "+provider.Name, nil)
				return ctrl.Result{Requeue: true}, nil
			}
		}
		logger.Error(nil, "no provider found for "+dnsRecord.Spec.Name)
		status.Message = "no provider found for " + dnsRecord.Spec.Name
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	logger = logger.WithValues("provider", status.ProviderRef.Name)

	var dnsProvider dnsv1.DNSProvider
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: status.ProviderRef.Namespace,
		Name:      status.ProviderRef.Name,
	}, &dnsProvider); err != nil {
		status.ProviderRef.Namespace = ""
		status.ProviderRef.Name = ""
		showResult("unable to fetch DNSProvider", err)
		return ctrl.Result{Requeue: true}, nil
	}

	if !dnsRecord.Match(&dnsProvider.Spec) {
		status.ProviderRef.Namespace = ""
		status.ProviderRef.Name = ""
		showResult("provider not match for "+dnsRecord.Spec.Name, errors.New("provider mismatch"))
		return ctrl.Result{Requeue: true}, nil
	}

	if !dnsProvider.Status.Valid {
		status.Message = "wait for provider to be valid"
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	iprovider, err := provider.New(ctx, &dnsProvider.Spec)
	if err != nil {
		status.ProviderRef.Namespace = ""
		status.ProviderRef.Name = ""
		showResult("unable to create provider", err)

		dnsProvider.Status.Valid = false
		dnsProvider.Status.Message = err.Error()
		r.Status().Update(ctx, &dnsProvider)

		return ctrl.Result{Requeue: true}, nil
	}

	if dnsRecord.Spec.TTL == nil {
		dnsRecord.Spec.TTL = &config.GetConfig().Default.Record.TTL
	}

	if status.Status != dnsv1.DNSRecordStatusPhaseSyncing {
		status.Status = dnsv1.DNSRecordStatusPhaseSyncing
		logger.Info("start syncing " + dnsRecord.Spec.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	recordID, ok, err := iprovider.SearchRecord(ctx, &dnsRecord)
	if err != nil {
		showResult("unable to search record", err)
		status.Status = dnsv1.DNSRecordStatusPhaseFailed
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	if dnsRecord.DeletionTimestamp.IsZero() {
		if ok {
			if err := iprovider.UpdateRecord(ctx, &dnsRecord, &recordID); err != nil {
				showResult("unable to update record", err)
				status.Status = dnsv1.DNSRecordStatusPhaseFailed
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
			status.Status = dnsv1.DNSRecordStatusPhaseSuccess
			showResult("synced", nil)
			if err := r.addFinalizer(ctx, dnsRecordOrigin); err != nil {
				logger.Error(err, "unable to add finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		} else {
			recordID, err := iprovider.CreateRecord(ctx, &dnsRecord)
			if err != nil {
				showResult("unable to create record", err)
				status.Status = dnsv1.DNSRecordStatusPhaseFailed
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
			status.Status = dnsv1.DNSRecordStatusPhaseSuccess
			status.RecordID = recordID
			showResult("synced", nil)
			if err := r.addFinalizer(ctx, dnsRecordOrigin); err != nil {
				logger.Error(err, "unable to add finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		if ok {
			if err := iprovider.DeleteRecord(ctx, &dnsRecord, &recordID); err != nil {
				showResult("unable to delete record", err)
				status.Status = dnsv1.DNSRecordStatusPhaseFailed
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
			status.Status = dnsv1.DNSRecordStatusPhaseSuccess
			showResult("deleted", nil)
			if err := r.removeFinalizer(ctx, dnsRecordOrigin); err != nil {
				logger.Error(err, "unable to remove finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		} else {
			status.Status = dnsv1.DNSRecordStatusPhaseSuccess
			showResult("deleted", nil)
			if err := r.removeFinalizer(ctx, dnsRecordOrigin); err != nil {
				logger.Error(err, "unable to remove finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}
}

func (r *DNSRecordReconciler) addFinalizer(ctx context.Context, dnsRecord *dnsv1.DNSRecord) error {
	if !util.ContainsString(dnsRecord.GetFinalizers(), util.FinalizerName) {
		dnsRecord.SetFinalizers(append(dnsRecord.GetFinalizers(), util.FinalizerName))
		return r.Update(ctx, dnsRecord)
	}
	return nil
}

func (r *DNSRecordReconciler) removeFinalizer(ctx context.Context, dnsRecord *dnsv1.DNSRecord) error {
	if util.ContainsString(dnsRecord.GetFinalizers(), util.FinalizerName) {
		dnsRecord.SetFinalizers(util.RemoveString(dnsRecord.GetFinalizers(), util.FinalizerName))
		return r.Update(ctx, dnsRecord)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSRecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("DNSRecord")
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1.DNSRecord{}).
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
