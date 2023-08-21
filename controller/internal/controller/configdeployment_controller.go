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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	kapps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	webappv1 "felipem1210/realworks-test/api/v1"
)

const (
	configMapField = ".spec.configMap"
)

// ConfigDeploymentReconciler reconciles a ConfigDeployment object
type ConfigDeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.tutorial.kubebuilder.io,resources=configdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.tutorial.kubebuilder.io,resources=configdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.tutorial.kubebuilder.io,resources=configdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *ConfigDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var configDeployment webappv1.ConfigDeployment
	if err := r.Get(ctx, req.NamespacedName, &configDeployment); err != nil {
		log.Error(err, "unable to fetch ConfigDeployment")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var configMapVersion string
	configMapName := configDeployment.Spec.ConfigMap
	foundConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: configDeployment.Namespace}, foundConfigMap)
	if err != nil {
		// If a configMap name is provided, then it must exist
		// You will likely want to create an Event for the user to understand why their reconcile is failing.
		return ctrl.Result{}, err
	}

	// Get the ConfigMap version
	configMapVersion = foundConfigMap.ResourceVersion

	// Get all the Deployments that are using this ConfigMap
	deployments, err := r.getDeploymentsUsingConfigMap(ctx, foundConfigMap, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, deployment := range deployments {
		// Modify pod template annotations this will trigger a rolling update
		if deployment.Spec.Template.Annotations["configMapVersion"] == configMapVersion {
			log.Info("The Deployment already using ConfigMap version", "configMapVersion", configMapVersion)
		} else {
			log.Info("ConfigMap changed", "configMapName", configMapName)
			log.Info("New ConfigMap version detected", "configMapVersion", configMapVersion)
			log.Info("Updating deployment", "name", deployment.Name, "namespace", deployment.Namespace)
			// Update the Deployments with the new ConfigMap version
			deployment.Spec.Template.Annotations["configMapVersion"] = configMapVersion
			if err := r.Update(ctx, &deployment); err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Updated Deployment with new ConfigMap version", "configMapVersion", configMapVersion)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ConfigDeploymentReconciler) getDeploymentsUsingConfigMap(ctx context.Context, configMap *corev1.ConfigMap, ns string) ([]kapps.Deployment, error) {
	// Search all Deployments that have the annotation referencing the ConfigMap
	deploymentList := &kapps.DeploymentList{}
	listOptions := &client.ListOptions{
		Namespace: ns,
	}

	if err := r.List(ctx, deploymentList, listOptions); err != nil {
		return nil, err
	}

	// Filter the Deployments that have the annotation referencing the ConfigMap
	var deploymentsUsingConfigMap []kapps.Deployment
	for _, deployment := range deploymentList.Items {
		if deployment.Annotations["configMapUsed"] == configMap.Name {
			deploymentsUsingConfigMap = append(deploymentsUsingConfigMap, deployment)
		}
	}

	return deploymentsUsingConfigMap, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	/*
		The `configMap` field must be indexed by the manager, so that we will be able to lookup `ConfigDeployments` by a referenced `ConfigMap` name.
		This will allow for quickly answer the question:
		- If ConfigMap _x_ is updated, which ConfigDeployments are affected?
	*/

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &webappv1.ConfigDeployment{}, configMapField, func(rawObj client.Object) []string {
		// Extract the ConfigMap name from the ConfigDeployment Spec, if one is provided
		configDeployment := rawObj.(*webappv1.ConfigDeployment)
		if configDeployment.Spec.ConfigMap == "" {
			return nil
		}
		return []string{configDeployment.Spec.ConfigMap}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.ConfigDeployment{}).
		Owns(&kapps.Deployment{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

// findObjectsForConfigMap will find all the ConfigDeployments that reference a given ConfigMap name
// That way we can enqueue reconcile requests for all the ConfigDeployments that reference a ConfigMap that has been updated
func (r *ConfigDeploymentReconciler) findObjectsForConfigMap(ctx context.Context, configMap client.Object) []reconcile.Request {
	attachedConfigDeployments := &webappv1.ConfigDeploymentList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(configMapField, configMap.GetName()),
		Namespace:     configMap.GetNamespace(),
	}
	err := r.List(context.TODO(), attachedConfigDeployments, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedConfigDeployments.Items))
	for i, item := range attachedConfigDeployments.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}
