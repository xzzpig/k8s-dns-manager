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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/generator"
)

// DNSGeneratorReconciler reconciles a DNSGenerator object
type DNSGeneratorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsgenerators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsgenerators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dns.xzzpig.com,resources=dnsgenerators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DNSGenerator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DNSGeneratorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dnsGenerator dnsv1.DNSGenerator
	if err := r.Get(ctx, req.NamespacedName, &dnsGenerator); err != nil {
		logger.Error(err, "unable to fetch DNSGenerator")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := generator.New(&dnsGenerator, ctx)
	if err != nil {
		logger.Error(err, "unable to generate DNS")
		dnsGenerator.Status.Valid = false
		dnsGenerator.Status.Message = err.Error()
		if err := r.Status().Update(ctx, &dnsGenerator); err != nil {
			logger.Error(err, "unable to update DNSGenerator status")
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	dnsGenerator.Status.Valid = true
	dnsGenerator.Status.Message = "ok"
	if err := r.Status().Update(ctx, &dnsGenerator); err != nil {
		logger.Error(err, "unable to update DNSGenerator status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSGeneratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1.DNSGenerator{}).
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
