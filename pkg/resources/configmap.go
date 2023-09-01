package resources

import (
	"context"
	"fmt"
	"reflect"

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

func NewConfigMap(cr *bigdatav1alpha1.Superset, scheme *runtime.Scheme) *corev1.ConfigMap {

	supersetConfigPy := `
import os
from cachelib.redis import RedisCache

def env(key, default=None):
	return os.getenv(key, default)

MAPBOX_API_KEY = env('MAPBOX_API_KEY', '')
CACHE_CONFIG = {
		'CACHE_TYPE': 'redis',
		'CACHE_DEFAULT_TIMEOUT': 300,
		'CACHE_KEY_PREFIX': 'superset_',
		'CACHE_REDIS_HOST': env('REDIS_HOST'),
		'CACHE_REDIS_PORT': env('REDIS_PORT'),
		'CACHE_REDIS_PASSWORD': env('REDIS_PASSWORD'),
		'CACHE_REDIS_DB': env('REDIS_DB', 1),
}
DATA_CACHE_CONFIG = CACHE_CONFIG

SQLALCHEMY_DATABASE_URI = f"postgresql+psycopg2://{env('DB_USER')}:{env('DB_PASS')}@{env('DB_HOST')}:{env('DB_PORT')}/{env('DB_NAME')}"
SQLALCHEMY_TRACK_MODIFICATIONS = True
SECRET_KEY = env('SECRET_KEY', 'thisISaSECRET_1234')

SESSION_COOKIE_SAMESITE = "None"
SESSION_COOKIE_SECURE = False
SESSION_COOKIE_HTTPONLY = False

class CeleryConfig(object):
	CELERY_IMPORTS = ('superset.sql_lab', )
	CELERY_ANNOTATIONS = {'tasks.add': {'rate_limit': '10/s'}}
	BROKER_URL = f"redis://{env('REDIS_HOST')}:{env('REDIS_PORT')}/0"
	CELERY_RESULT_BACKEND = f"redis://{env('REDIS_HOST')}:{env('REDIS_PORT')}/0"

CELERY_CONFIG = CeleryConfig
RESULTS_BACKEND = RedisCache(
		host=env('REDIS_HOST'),
		port=env('REDIS_PORT'),
		key_prefix='superset_results'
)
`

	supersetInitSh := `
	#!/bin/sh
    set -eu
    echo "Upgrading DB schema..."
    superset db upgrade
    superset set_database_uri -d Impala -u impala://impala-master-0.impala-master-svc.default.svc.cluster.local:21050
    echo "Initializing roles..."
    superset init
    
    echo "Creating admin user..."
    superset fab create-admin \
                    --username $ADMIN_USER \
                    --firstname $ADMIN_FIRSTNAME \
                    --lastname $ADMIN_LASTNAME \
                    --email $ADMIN_EMAIL \
                    --password $ADMIN_PASS \
                    || true

    if [ -f "/app/configs/import_datasources.yaml" ]; then
      echo "Importing database connections.... "
      superset import_datasources -p /app/configs/import_datasources.yaml
    fi
	`

	supersetBootstrapSh := `
    #!/bin/bash
    rm -rf /var/lib/apt/lists/* && \
    pip install \
      psycopg2-binary==2.9.7 \
      redis==4.6.0 \
      impyla \
      pyhive==0.7.0 \
      sqlalchemy-solr && \
    if [ ! -f ~/bootstrap ]; then echo "Running Superset with uid 0" > ~/bootstrap; fi
	`

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "superset-config",
			Namespace: cr.Namespace,
			Labels:    util.LabelsForSuperset(cr, "superset"),
		},
		Data: map[string]string{
			"superset_config.py":    supersetConfigPy,
			"superset_init.sh":      supersetInitSh,
			"superset_bootstrap.sh": supersetBootstrapSh,
		},
	}

	controllerutil.SetControllerReference(cr, configMap, scheme)

	return configMap
}

func ReconcileConfigMap(ctx context.Context, client runtimeClient.Client, desired *corev1.ConfigMap) error {

	log := log.FromContext(ctx)
	log.Info("Reconciling ConfigMap")
	// Get the current ConfigMap
	current := &corev1.ConfigMap{}
	key := runtimeClient.ObjectKeyFromObject(desired)
	err := client.Get(ctx, key, current)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// The ConfigMap does not exist yet, so we'll create it
			log.Info("Creating ConfigMap")
			err = client.Create(ctx, desired)
			if err != nil {
				log.Error(err, "unable to create ConfigMap")
				return fmt.Errorf("unable to create ConfigMap: %v", err)
			} else {
				return nil
			}
		}

		log.Error(err, "error getting ConfigMap")
		return fmt.Errorf("error getting ConfigMap: %v", err)
	}

	// Check if the current ConfigMap matches the desired one
	if !reflect.DeepEqual(current.Data, desired.Data) || !reflect.DeepEqual(current.BinaryData, desired.BinaryData) || !reflect.DeepEqual(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
		// If it doesn't match, we'll update the current ConfigMap to match the desired one
		current.Data = desired.Data
		current.BinaryData = desired.BinaryData
		current.ObjectMeta.Labels = desired.ObjectMeta.Labels
		log.Info("Updating ConfigMap")
		err = client.Update(ctx, current)
		if err != nil {
			log.Error(err, "unable to update ConfigMap")
			return fmt.Errorf("unable to update ConfigMap: %v", err)
		}
	}

	// If we reach here, it means the ConfigMap is in the desired state
	return nil
}
