package util

import (
	corev1 "k8s.io/api/core/v1"

	bigdatav1alpha1 "github.com/kubernetesbigdataeg/superset-operator/api/v1alpha1"
)

const (
	baseImage        = "docker.io/kubernetesbigdataeg/superset"
	baseImageVersion = "2.1.0-1"
	masterSize       = 1
	workerSize       = 1
	MemoryRequest    = "1G"
	CpuRequest       = "2"
	DefaultUser      = "admin"
	DefaultPass      = "admin"
	DefaultFirstname = "Superset"
	DefaultLastname  = "Admin"
	DefaultEmail     = "admin@superset.com"
)

// labelsForSuperset returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func LabelsForSuperset(cr *bigdatav1alpha1.Superset, app string) map[string]string {

	return map[string]string{
		"app.kubernetes.io/name":       "Superset",
		"app.kubernetes.io/instance":   cr.Name,
		"app.kubernetes.io/version":    GetImageVersion(cr),
		"app.kubernetes.io/part-of":    "superset-operator",
		"app.kubernetes.io/created-by": "controller-manager",
		"app":                          app,
	}

}

func GetImage(cr *bigdatav1alpha1.Superset) string {
	return baseImage + ":" + GetImageVersion(cr)
}

func GetImageVersion(cr *bigdatav1alpha1.Superset) string {
	if len(cr.Spec.BaseImageVersion) > 0 {
		return cr.Spec.BaseImageVersion
	}
	return baseImageVersion
}

func GetMasterSize(cr *bigdatav1alpha1.Superset) int32 {
	if cr.Spec.MasterSize > 0 {
		return cr.Spec.MasterSize
	}
	return masterSize
}

func GetWorkerSize(cr *bigdatav1alpha1.Superset) int32 {
	if cr.Spec.WorkerSize > 0 {
		return cr.Spec.WorkerSize
	}
	return workerSize
}

func GetEnvVars(cr *bigdatav1alpha1.Superset) []corev1.EnvVar {

	envVars := []corev1.EnvVar{
		{
			Name:  "SUPERSET_PORT",
			Value: "8088",
		},
		{
			Name:  "REDIS_HOST",
			Value: cr.Spec.DbConnection.RedisHost,
		},
		{
			Name:  "REDIS_PORT",
			Value: cr.Spec.DbConnection.RedisPort,
		},
		{
			Name:  "DB_HOST",
			Value: cr.Spec.DbConnection.DbHost,
		},
		{
			Name:  "DB_PORT",
			Value: cr.Spec.DbConnection.DbPort,
		},
		{
			Name:  "DB_USER",
			Value: cr.Spec.DbConnection.DbUser,
		},
		{
			Name:  "DB_PASS",
			Value: cr.Spec.DbConnection.DbPass,
		},
		{
			Name:  "DB_NAME",
			Value: cr.Spec.DbConnection.DbName,
		},
	}

	return append(envVars, getAdmin(cr)...)

}

func getAdmin(cr *bigdatav1alpha1.Superset) []corev1.EnvVar {

	if cr.Spec.Admin.User == "" {
		cr.Spec.Admin.User = DefaultUser
	}
	if cr.Spec.Admin.Pass == "" {
		cr.Spec.Admin.Pass = DefaultPass
	}
	if cr.Spec.Admin.Firstname == "" {
		cr.Spec.Admin.Firstname = DefaultFirstname
	}
	if cr.Spec.Admin.Lastname == "" {
		cr.Spec.Admin.Lastname = DefaultLastname
	}
	if cr.Spec.Admin.Email == "" {
		cr.Spec.Admin.Email = DefaultEmail
	}

	return []corev1.EnvVar{
		{
			Name:  "ADMIN_USER",
			Value: cr.Spec.Admin.User,
		},
		{
			Name:  "ADMIN_PASS",
			Value: cr.Spec.Admin.Pass,
		},
		{
			Name:  "ADMIN_FIRSTNAME",
			Value: cr.Spec.Admin.Firstname,
		},
		{
			Name:  "ADMIN_LASTNAME",
			Value: cr.Spec.Admin.Lastname,
		},
		{
			Name:  "ADMIN_EMAIL",
			Value: cr.Spec.Admin.Email,
		},
	}

}

func GetVolume() []corev1.Volume {

	return []corev1.Volume{
		{
			Name: "superset-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "superset-config",
					},
				},
			},
		},
	}
}

func GetVolumeMount() []corev1.VolumeMount {

	return []corev1.VolumeMount{
		{
			Name:      "superset-config",
			MountPath: "/app/pythonpath",
			ReadOnly:  true,
		},
	}

}
