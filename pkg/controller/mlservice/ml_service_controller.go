package mlservice

import (
	"context"
	"fmt"
	"reflect"

	"github.com/rancher/wrangler/v2/pkg/condition"
	ctlcorev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	ctloneblockv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ml.oneblock.ai/v1"
	ctlrayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils"
)

const (
	mlServiceControllerOnChange   = "mlService.onChange"
	mlServiceControllerCreatePVC  = "mlService.createPVCFromAnnotation"
	mlServiceControllerSyncStatus = "mlService.syncRayServiceStatus"
)

type Handler struct {
	ctx                       context.Context
	releaseName               string
	mlService                 ctloneblockv1.MLServiceController
	mlServiceCache            ctloneblockv1.MLServiceCache
	rayService                ctlrayv1.RayServiceController
	rayServiceCache           ctlrayv1.RayServiceCache
	configmap                 ctlcorev1.ConfigMapController
	configmapCache            ctlcorev1.ConfigMapCache
	secret                    ctlcorev1.SecretController
	secretCache               ctlcorev1.SecretCache
	modelTemplateVersion      ctloneblockv1.ModelTemplateVersionController
	modelTemplateVersionCache ctloneblockv1.ModelTemplateVersionCache
	pvcHandler                *utils.PVCHandler
}

func Register(ctx context.Context, mgmt *config.Management) error {
	mlService := mgmt.OneBlockMLFactory.Ml().V1().MLService()
	templateVersion := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplateVersion()
	rayService := mgmt.KubeRayFactory.Ray().V1().RayService()
	pvcs := mgmt.CoreFactory.Core().V1().PersistentVolumeClaim()
	configmaps := mgmt.CoreFactory.Core().V1().ConfigMap()
	secrets := mgmt.CoreFactory.Core().V1().Secret()
	handler := &Handler{
		ctx:                       ctx,
		releaseName:               mgmt.ReleaseName,
		mlService:                 mlService,
		mlServiceCache:            mlService.Cache(),
		rayService:                rayService,
		rayServiceCache:           rayService.Cache(),
		configmap:                 configmaps,
		configmapCache:            configmaps.Cache(),
		secret:                    secrets,
		secretCache:               secrets.Cache(),
		modelTemplateVersion:      templateVersion,
		modelTemplateVersionCache: templateVersion.Cache(),
		pvcHandler:                utils.NewPVCHandler(pvcs, pvcs.Cache()),
	}

	mlService.OnChange(ctx, mlServiceControllerOnChange, handler.OnChange)
	mlService.OnChange(ctx, mlServiceControllerCreatePVC, handler.createMLServicePVCs)
	rayService.OnChange(ctx, mlServiceControllerSyncStatus, handler.syncRayServiceStatus)
	return nil
}

// OnChange method will help to serve the LLM model inference
// 1. sync required resources like model config and secrets to the local NS
// 2. serve and reconcile the serving parameters using the RayService
func (h *Handler) OnChange(_ string, mlService *mlv1.MLService) (*mlv1.MLService, error) {
	if mlService == nil || mlService.DeletionTimestamp != nil {
		return mlService, nil
	}

	modelRef := mlService.Spec.ModelTemplateVersionRef
	// get modelRef spec from the modelRef template version
	modelTmpVersion, err := h.modelTemplateVersionCache.Get(modelRef.Namespace, modelRef.Name)
	if err != nil {
		if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, err.Error()); err != nil {
			return mlService, err
		}
		return mlService, err
	}

	if !mlv1.ModelTemplateVersionConfigured.IsTrue(modelTmpVersion) {
		message := fmt.Sprintf("skip serving, model template version %s:%s is not configured correctly", modelRef.Name, modelRef.Namespace)
		if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, message); err != nil {
			return mlService, err
		}
		return mlService, err
	}

	// get the modelRef template version and save it as a configmap
	if _, err = h.createModelConfigMap(modelTmpVersion, mlService.Namespace); err != nil {
		if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, err.Error()); err != nil {
			return mlService, err
		}
		return mlService, err
	}

	// sync HF secret to the local ns
	if mlService.Spec.HFSecretRef != nil {
		if err = h.SyncClusterSecretsToLocalNS(mlService.Spec.HFSecretRef, mlService.Namespace); err != nil {
			if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, err.Error()); err != nil {
				return mlService, err
			}
			return mlService, err
		}
	}

	// serve the LLM model use RayService
	raySvc, err := h.rayServiceCache.Get(mlService.Namespace, mlService.Name)
	if err != nil && !errors.IsNotFound(err) {
		if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, err.Error()); err != nil {
			return mlService, err
		}
		return mlService, err
	}

	// ensuring ML cluster, create a new one by RayService if not exist
	if raySvc == nil {
		owners := generateMLServiceOwnerReference(mlService)
		rayService, err := getRayServiceConfig(mlService, modelTmpVersion, owners, h.releaseName)
		if err != nil {
			if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, err.Error()); err != nil {
				return mlService, err
			}
			return mlService, err
		}

		raySvc, err = h.rayService.Create(rayService)
		if err != nil {
			if err = h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, false, err.Error()); err != nil {
				return mlService, err
			}
			return mlService, err
		}

		return mlService, h.updateMLServiceCondition(mlService, mlv1.MLServiceCreated, true, "")
	}

	// updating the RayService if it is modified
	raySvcCpy := raySvc.DeepCopy()
	SetRayClusterImage(mlService, raySvcCpy)
	SetRayClusterHeadConfig(mlService, raySvcCpy)
	SetRayClusterWorkerGroupConfig(mlService, raySvcCpy)
	if !reflect.DeepEqual(raySvc.Spec, raySvcCpy.Spec) {
		logrus.Debugf("updating RayService: %s, spec:%v", raySvcCpy.Name, raySvcCpy.Spec)
		if _, err = h.rayService.Update(raySvcCpy); err != nil {
			return mlService, err
		}
	}

	return nil, nil
}

func (h *Handler) createModelConfigMap(modelTemplateVersion *mlv1.ModelTemplateVersion, namespace string) (*corev1.ConfigMap, error) {
	modelCfg, err := h.configmapCache.Get(modelTemplateVersion.Namespace, modelTemplateVersion.Name)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	// since model reference cannot be modified, if the configmap already exists, just return it
	if modelCfg != nil {
		return modelCfg, nil
	}

	// create a new configmap
	modelCfg = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      modelTemplateVersion.Name,
			Namespace: namespace,
		},
		Data: map[string]string{
			GetModelConfigMapKey(modelTemplateVersion.Name): modelTemplateVersion.Status.GeneratedModelConfig,
		},
	}
	return h.configmap.Create(modelCfg)
}

func (h *Handler) createMLServicePVCs(_ string, mlService *mlv1.MLService) (*mlv1.MLService, error) {
	if mlService == nil || mlService.DeletionTimestamp != nil {
		return mlService, nil
	}

	var volumes = make([]mlv1.Volume, 0)
	var volumeNames = make([]string, 0)
	headGroupVol := mlService.Spec.MLClusterRef.RayClusterSpec.HeadGroupSpec.Volume
	if headGroupVol != nil {
		volumes = append(volumes, *headGroupVol)
		volumeNames = append(volumeNames, headGroupVol.Name)
	}

	for _, wg := range mlService.Spec.MLClusterRef.RayClusterSpec.WorkerGroupSpec {
		if wg.Volume != nil {
			volumes = append(volumes, *wg.Volume)
			volumeNames = append(volumeNames, wg.Name)
		}
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	// skip PVC garbage collection by not setting its ownerRefs
	err := h.pvcHandler.CreatePVCByVolume(volumes, mlService.Namespace, nil)
	if err != nil {
		return nil, err
	}

	return mlService, nil
}

func (h *Handler) SyncClusterSecretsToLocalNS(hfRef *mlv1.HFSecretRef, namespace string) error {
	hfSecret, err := h.secretCache.Get(hfRef.Namespace, hfRef.Name)
	if err != nil {
		return fmt.Errorf("fail to find the HF secret: %v", err)
	}

	// we only need to create 1 synced HF secret in the related local namespace
	nsSecret, err := h.secretCache.Get(namespace, hfRef.Name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if nsSecret == nil {
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      hfRef.Name,
				Namespace: namespace,
			},
			Data: hfSecret.Data,
		}
		if _, err = h.secret.Create(newSecret); err != nil {
			return fmt.Errorf("failed to sync HF secret to ns %s: %v", namespace, err)
		}
		return nil
	}

	if !reflect.DeepEqual(hfSecret.Data, nsSecret.Data) {
		secretCpy := nsSecret.DeepCopy()
		secretCpy.Data = hfSecret.Data

		if _, err = h.secret.Update(secretCpy); err != nil {
			return fmt.Errorf("failed to update secret %s:%s, %v", nsSecret.Name, nsSecret.Namespace, err)
		}
	}
	return nil
}

func generateMLServiceOwnerReference(mlService *mlv1.MLService) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion: mlService.APIVersion,
			Kind:       mlService.Kind,
			Name:       mlService.Name,
			UID:        mlService.UID,
		},
	}
}

func (h *Handler) updateMLServiceCondition(service *mlv1.MLService, cond condition.Cond, isTrue bool, message string) error {
	if isTrue {
		if cond.IsTrue(service) {
			return nil
		}
		cond.True(service)
	} else {
		if cond.IsFalse(service) && cond.GetMessage(service) == message {
			return nil
		}
		cond.False(service)
		cond.Message(service, message)
	}
	svcCpy := service.DeepCopy()
	if _, err := h.mlService.UpdateStatus(svcCpy); err != nil {
		return err
	}
	return nil
}
