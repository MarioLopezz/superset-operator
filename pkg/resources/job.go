package resources

import (
	"context"
	"fmt"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bigdatav1alpha1 "github.com/kubernetesbigdataeg/superset-operator/api/v1alpha1"
	"github.com/kubernetesbigdataeg/superset-operator/pkg/util"
)

func NewJob(cr *bigdatav1alpha1.Superset, scheme *runtime.Scheme) *batchv1.Job {

	labels := util.LabelsForSuperset(cr, "superset-job")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "superset-init-db",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:            "wait-for-postgres",
							Image:           "jwilder/dockerize:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-c",
								`dockerize -wait "tcp://$DB_HOST:$DB_PORT" -timeout 120s`,
							},
							Env: util.GetEnvVars(cr),
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "superset-init-db",
							Image:           util.GetImage(cr),
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"/bin/sh",
								"-c",
								". /app/pythonpath/superset_bootstrap.sh; . /app/pythonpath/superset_init.sh",
							},
							Env:          util.GetEnvVars(cr),
							VolumeMounts: util.GetVolumeMount(),
						},
					},
					Volumes:       util.GetVolume(),
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, job, scheme)

	return job
}

func ReconcileJob(ctx context.Context, client runtimeClient.Client, desired *batchv1.Job) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling Job")

	// Get the current Job
	current := &batchv1.Job{}
	key := runtimeClient.ObjectKeyFromObject(desired)
	err := client.Get(ctx, key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "error getting Job")
		return fmt.Errorf("error getting Job: %v", err)
	}

	if current.Status.Succeeded > 0 {
		// Job has completed successfully, nothing to do.
		log.Info("Job has completed successfully")
		return nil
	}

	// Check if the Job is currently running
	if current.Status.Active > 0 {
		log.Info("Job is currently running")
		return nil
	}

	if apierrors.IsNotFound(err) {
		// The Job does not exist yet, so we'll create it
		log.Info("Creating Job")
		if err := client.Create(ctx, desired); err != nil {
			log.Error(err, "unable to create Job")
			return fmt.Errorf("unable to create Job: %v", err)
		}
		return nil
	}

	if !reflect.DeepEqual(current.Spec, desired.Spec) || !reflect.DeepEqual(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
		log.Info("Job spec or labels don't match, updating Job")

		// If it doesn't match, we'll update the current Job to match the desired one
		current.Spec = desired.Spec
		current.ObjectMeta.Labels = desired.ObjectMeta.Labels

		err = client.Update(ctx, current)
		if err != nil {
			log.Error(err, "unable to update Job")
			return fmt.Errorf("unable to update Job: %v", err)
		}
	}

	// If we reach here, it means the Job is in the desired state
	return nil
}
