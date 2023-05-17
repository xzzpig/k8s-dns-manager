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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/provider"
)

// DNSProviderReconciler reconciles a DNSProvider object
type DNSProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsproviders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsproviders/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DNSProvider object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DNSProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dnsProvider dnsv1.DNSProvider
	if err := r.Get(ctx, req.NamespacedName, &dnsProvider); err != nil {
		logger.Error(err, "unable to fetch DNSProvider")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if dnsProvider.Spec.Selector != nil {
		_, err := metav1.LabelSelectorAsSelector(dnsProvider.Spec.Selector)
		if err != nil {
			logger.Error(err, "unable to parse DNSProvider selector")
			dnsProvider.Status.Valid = false
			dnsProvider.Status.Message = err.Error()
			if err := r.Status().Update(ctx, &dnsProvider); err != nil {
				logger.Error(err, "unable to update DNSProvider status")
			}
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	}

	_, err := provider.New(ctx, &dnsProvider.Spec)
	if err != nil {
		logger.Error(err, "unable to create provider")
		dnsProvider.Status.Valid = false
		dnsProvider.Status.Message = err.Error()
		if err := r.Status().Update(ctx, &dnsProvider); err != nil {
			logger.Error(err, "unable to update DNSProvider status")
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	dnsProvider.Status.Valid = true
	dnsProvider.Status.Message = "ok"
	if err := r.Status().Update(ctx, &dnsProvider); err != nil {
		logger.Error(err, "unable to update DNSProvider status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1.DNSProvider{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldGeneration := e.ObjectOld.GetGeneration()
				newGeneration := e.ObjectNew.GetGeneration()
				// Generation is only updated on spec changes (also on deletion),
				// not metadata or status
				// Filter out events where the generation hasn't changed to
				// avoid being triggered by status updates

				return oldGeneration != newGeneration
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
