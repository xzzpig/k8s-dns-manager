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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

	_, err := provider.New(&dnsProvider.Spec)
	if err != nil {
		logger.Error(err, "unable to create provider")
		dnsProvider.Status.Valid = false
		dnsProvider.Status.Message = err.Error()
		if err := r.Status().Update(ctx, &dnsProvider); err != nil {
			logger.Error(err, "unable to update DNSProvider status")
		}
		return ctrl.Result{}, err
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
		Complete(r)
}
