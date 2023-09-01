package resources

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bigdatav1alpha1 "github.com/kubernetesbigdataeg/superset-operator/api/v1alpha1"
	"github.com/kubernetesbigdataeg/superset-operator/pkg/util"
)

func NewMasterDeployment(cr *bigdatav1alpha1.Superset, scheme *runtime.Scheme) *appsv1.Deployment {

	labels := util.LabelsForSuperset(cr, "superset")

	replicas := util.GetMasterSize(cr)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-master",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers:     getMasterContainers(cr),
					InitContainers: getMasterInitContainer(cr),
					Volumes:        util.GetVolume(),
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, deployment, scheme)

	return deployment
}

func getMasterContainers(cr *bigdatav1alpha1.Superset) []corev1.Container {

	startup, readiness, liveness := getMasterProbes()

	container := []corev1.Container{
		{
			Image:           util.GetImage(cr),
			Name:            cr.Name,
			ImagePullPolicy: corev1.PullAlways,
			Command:         []string{"/bin/sh", "-c", ". /app/pythonpath/superset_bootstrap.sh; /usr/bin/run-server.sh"},
			Resources:       getResources(cr),
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 8088,
				},
			},
			Env:            util.GetEnvVars(cr),
			VolumeMounts:   util.GetVolumeMount(),
			StartupProbe:   startup,
			ReadinessProbe: readiness,
			LivenessProbe:  liveness,
		},
	}

	return container
}

func getMasterInitContainer(cr *bigdatav1alpha1.Superset) []corev1.Container {

	initContainer := []corev1.Container{
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
	}

	return initContainer

}

func NewWorkerDeployment(cr *bigdatav1alpha1.Superset, scheme *runtime.Scheme) *appsv1.Deployment {

	labels := util.LabelsForSuperset(cr, "superset-worker")

	replicas := util.GetWorkerSize(cr)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-worker",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers:     getWorkerContainers(cr),
					InitContainers: getWorkerInitContainer(cr),
					Volumes:        util.GetVolume(),
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, deployment, scheme)

	return deployment
}

func getWorkerContainers(cr *bigdatav1alpha1.Superset) []corev1.Container {

	container := []corev1.Container{
		{
			Image:           util.GetImage(cr),
			Name:            cr.Name,
			ImagePullPolicy: corev1.PullAlways,
			Command:         []string{"/bin/sh", "-c", ". /app/pythonpath/superset_bootstrap.sh; celery --app=superset.tasks.celery_app:app worker"},
			Resources:       getResources(cr),
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 8088,
				},
			},
			Env:           util.GetEnvVars(cr),
			VolumeMounts:  util.GetVolumeMount(),
			LivenessProbe: getWorkerProbes(),
		},
	}

	return container
}

func getWorkerInitContainer(cr *bigdatav1alpha1.Superset) []corev1.Container {

	initContainer := []corev1.Container{
		{
			Name:            "wait-for-postgres-redis",
			Image:           "jwilder/dockerize:latest",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"/bin/sh",
				"-c",
				`dockerize -wait "tcp://$DB_HOST:$DB_PORT" -wait "tcp://$REDIS_HOST:$REDIS_PORT" -timeout 120s`,
			},
			Env: util.GetEnvVars(cr),
		},
	}

	return initContainer

}

func getMasterProbes() (startupProbe, readinessProbe, livenessProbe *corev1.Probe) {

	httpGetConfig := &corev1.HTTPGetAction{
		Path: "/health",
		Port: intstr.FromString("http"),
	}

	startupProbe = &corev1.Probe{
		FailureThreshold: 60,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: httpGetConfig,
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}

	readinessProbe = &corev1.Probe{
		FailureThreshold: 3,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: httpGetConfig,
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       15,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}

	livenessProbe = &corev1.Probe{
		FailureThreshold: 3,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: httpGetConfig,
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       15,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}

	return
}

func getWorkerProbes() *corev1.Probe {

	execConfig := &corev1.ExecAction{
		Command: []string{
			"sh",
			"-c",
			"celery -A superset.tasks.celery_app:app inspect ping -d celery@$HOSTNAME",
		},
	}

	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: execConfig,
		},
		FailureThreshold:    3,
		InitialDelaySeconds: 120,
		PeriodSeconds:       60,
		SuccessThreshold:    1,
		TimeoutSeconds:      60,
	}
}

func getResources(cr *bigdatav1alpha1.Superset) corev1.ResourceRequirements {

	if len(cr.Spec.Deployment.Resources.Requests) > 0 || len(cr.Spec.Deployment.Resources.Limits) > 0 {
		return cr.Spec.Deployment.Resources
	} else {
		return corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(util.CpuRequest),
				corev1.ResourceMemory: resource.MustParse(util.MemoryRequest),
			},
		}
	}

}

func ReconcileDeployment(ctx context.Context, client runtimeClient.Client, desired *appsv1.Deployment) error {

	log := log.FromContext(ctx)
	log.Info("Reconciling Deployment " + desired.ObjectMeta.Name)
	// Get the current Deployment
	current := &appsv1.Deployment{}
	key := runtimeClient.ObjectKeyFromObject(desired)
	err := client.Get(ctx, key, current)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// The Deployment does not exist yet, so we'll create it
			log.Info("Creating Deployment " + desired.ObjectMeta.Name)
			err = client.Create(ctx, desired)
			if err != nil {
				log.Error(err, "unable to create Deployment "+desired.ObjectMeta.Name)
				return fmt.Errorf("unable to create Deployment "+desired.ObjectMeta.Name+": %v", err)
			} else {
				return nil
			}
		} else {
			return fmt.Errorf("error getting Deployment "+desired.ObjectMeta.Name+": %v", err)
		}
	}

	// TODO: Check if it has changed in more specific fields.
	// Check if the current Deployment matches the desired one
	if !reflect.DeepEqual(current.Spec, desired.Spec) {
		// If it doesn't match, we'll update the current Deployment to match the desired one
		current.Spec.Replicas = desired.Spec.Replicas
		current.Spec.Template = desired.Spec.Template
		log.Info("Updating Deployment" + desired.ObjectMeta.Name)
		err = client.Update(ctx, current)
		if err != nil {
			log.Error(err, "unable to update Deployment"+desired.ObjectMeta.Name)
			return fmt.Errorf("unable to update Deployment "+desired.ObjectMeta.Name+": %v", err)
		}
	}

	// If we reach here, it means the Deployment is in the desired state
	return nil
}
