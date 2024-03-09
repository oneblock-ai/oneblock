package modeltemplate

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	ctlmlv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	templateControllerSetDefaultVersion = "templateController.setDefaultVersion"
	templateControllerSyncLatestVersion = "templateController.syncLatestVersion"
	templateControllerAssignVersion     = "templateController.assignVersion"
)

type TemplateHandler struct {
	templateController    ctlmlv1.ModelTemplateController
	templateCache         ctlmlv1.ModelTemplateCache
	templateVersionClient ctlmlv1.ModelTemplateVersionClient
	templateVersionCache  ctlmlv1.ModelTemplateVersionCache

	latestVersionMap map[string]int
	mutex            *sync.Mutex
}

func TemplateRegister(ctx context.Context, mgmt *config.Management) error {
	templates := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplate()
	templateVersions := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplateVersion()
	h := &TemplateHandler{
		templateController:    templates,
		templateCache:         templates.Cache(),
		templateVersionClient: templateVersions,
		templateVersionCache:  templateVersions.Cache(),

		latestVersionMap: make(map[string]int),
		mutex:            &sync.Mutex{},
	}

	if err := h.initLatestVersionMap(); err != nil {
		return fmt.Errorf("failed to init model template version map: %w", err)
	}

	templates.OnChange(ctx, templateControllerSetDefaultVersion, h.SetDefaultVersion)
	templates.OnChange(ctx, templateControllerSyncLatestVersion, h.SyncLatestVersion)
	templates.OnRemove(ctx, templateControllerSyncLatestVersion, h.DeleteLatestVersion)
	templateVersions.OnChange(ctx, templateControllerAssignVersion, h.AssignVersion)

	return nil
}

// initLatestVersionMap initializes the latest version map by querying all the template versions
func (h *TemplateHandler) initLatestVersionMap() error {
	tvs, err := h.templateVersionClient.List("", metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, tv := range tvs.Items {
		ref := utils.NewRef(tv.Namespace, tv.Spec.TemplateName)
		// avoid gosec G601: Implicit memory aliasing from the loop variable
		if mlv1.ModelTemplateVersionAssigned.IsTrue(tv.DeepCopy()) && tv.Status.Version > h.latestVersionMap[ref] {
			//if tv.Status.Version > h.latestVersionMap[ref] {
			h.mutex.Lock()
			h.latestVersionMap[ref] = tv.Status.Version
			h.mutex.Unlock()
		}
	}

	return nil
}

// SetDefaultVersion sets the default version for the template
func (h *TemplateHandler) SetDefaultVersion(_ string, tp *mlv1.ModelTemplate) (*mlv1.ModelTemplate, error) {
	if tp == nil || tp.DeletionTimestamp != nil {
		return tp, nil
	}

	tpCopy := tp.DeepCopy()
	// set the first version as the default version if the default version id is empty.
	if tp.Spec.DefaultVersionID == "" {
		firstTemplateVersion, err := h.getFirstTemplateVersion(tp)
		if err != nil {
			return nil, err
		}
		if firstTemplateVersion == nil {
			return nil, nil
		}

		if mlv1.ModelTemplateVersionAssigned.IsFalse(firstTemplateVersion) {
			return nil, fmt.Errorf("the template version %s/%s of template %s/%s version number haven't been assigned yet",
				firstTemplateVersion.Namespace, firstTemplateVersion.Name, tp.Namespace, tp.Name)
		}
		tpCopy.Spec.DefaultVersionID = utils.NewRef(firstTemplateVersion.Namespace, firstTemplateVersion.Name)
		return h.templateController.Update(tpCopy)
	}
	ns, name := utils.Ref(tp.Spec.DefaultVersionID).Parse()
	tpv, err := h.templateVersionCache.Get(ns, name)
	if err != nil {
		if errors.IsNotFound(err) && !mlv1.ModelTemplateDefaultVersionAssigned.MatchesError(tpCopy, "", err) {
			mlv1.ModelTemplateDefaultVersionAssigned.SetError(tpCopy, "", err)
			if _, err = h.templateController.UpdateStatus(tpCopy); err != nil {
				return nil, err
			}
		}
		return tp, err
	}

	if mlv1.ModelTemplateVersionAssigned.IsFalse(tpv) {
		return nil, fmt.Errorf("the template version %s haven't been assigned a version number", tp.Spec.DefaultVersionID)
	}

	if tp.Status.DefaultVersion == tpv.Status.Version && mlv1.ModelTemplateDefaultVersionAssigned.IsTrue(tpCopy) {
		logrus.Debug("the default version is already set to the latest version")
		return nil, nil
	}
	tpCopy.Status.DefaultVersion = tpv.Status.Version
	tpCopy.Status.ObservedGeneration = tp.Generation
	mlv1.ModelTemplateDefaultVersionAssigned.SetError(tpCopy, "", nil)
	return h.templateController.UpdateStatus(tpCopy)
}

// getFirstTemplateVersion gets the first template version of the template
func (h *TemplateHandler) getFirstTemplateVersion(tp *mlv1.ModelTemplate) (*mlv1.ModelTemplateVersion, error) {
	selector := labels.Set(map[string]string{constant.LabelModelTemplateName: tp.Name}).AsSelector()
	tpvs, err := h.templateVersionCache.List(tp.Namespace, selector)
	if err != nil {
		return nil, err
	}
	if len(tpvs) == 0 {
		return nil, nil
	}

	sort.Sort(templateVersionByCreationTimestamp(tpvs))
	for _, v := range tpvs {
		if mlv1.ModelTemplateVersionAssigned.IsTrue(v) {
			return v, nil
		}
	}
	return nil, nil
}

func (h *TemplateHandler) DeleteLatestVersion(_ string, tp *mlv1.ModelTemplate) (*mlv1.ModelTemplate, error) {
	if tp == nil {
		return tp, nil
	}

	key := utils.NewRef(tp.Namespace, tp.Name)
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.latestVersionMap, key)

	return tp, nil
}

// SyncLatestVersion syncs the latest version from memory to the template CR
func (h *TemplateHandler) SyncLatestVersion(_ string, tp *mlv1.ModelTemplate) (*mlv1.ModelTemplate, error) {
	if tp == nil || tp.DeletionTimestamp != nil {
		return tp, nil
	}

	key := utils.NewRef(tp.Namespace, tp.Name)
	if tp.Status.LatestVersion == h.latestVersionMap[key] {
		return tp, nil
	}

	tpCopy := tp.DeepCopy()
	tpCopy.Status.LatestVersion = h.latestVersionMap[key]
	tpCopy.Status.ObservedGeneration = tp.Generation
	logrus.Debugf("delete the latest version record of %s/%s, %+v", tp.Namespace, tp.Name, h.latestVersionMap)

	return h.templateController.UpdateStatus(tpCopy)
}

// AssignVersion assigns a version number to the template version
func (h *TemplateHandler) AssignVersion(_ string, tpv *mlv1.ModelTemplateVersion) (*mlv1.ModelTemplateVersion, error) {
	if tpv == nil || tpv.DeletionTimestamp != nil || mlv1.ModelTemplateVersionAssigned.IsTrue(tpv) {
		return tpv, nil
	}

	tp, err := h.templateCache.Get(tpv.Namespace, tpv.Spec.TemplateName)
	if err != nil {
		return nil, err
	}

	tpvCopy := tpv.DeepCopy()

	// add template label
	if tpvCopy.Labels == nil {
		tpvCopy.Labels = make(map[string]string)
	}
	tpvCopy.Labels[constant.LabelModelTemplateName] = tp.Name

	// add owner reference
	if tpvCopy.OwnerReferences == nil {
		tpvCopy.OwnerReferences = []metav1.OwnerReference{
			{
				Name:       tp.Name,
				APIVersion: tp.APIVersion,
				UID:        tp.UID,
				Kind:       tp.Kind,
			},
		}
	}

	if _, err = h.templateVersionClient.Update(tpvCopy); err != nil {
		return tpv, err
	}

	// assign version
	tpvCopy.Status.Version = h.latestVersionMap[utils.NewRef(tp.Namespace, tp.Name)] + 1
	mlv1.ModelTemplateVersionAssigned.True(&tpvCopy.Status)
	if tpv, err = h.templateVersionClient.UpdateStatus(tpvCopy); err != nil {
		return tpv, err
	}

	h.mutex.Lock()
	h.latestVersionMap[utils.NewRef(tp.Namespace, tp.Name)]++
	h.mutex.Unlock()
	logrus.Debugf("assigned new version: %v", h.latestVersionMap)

	// trigger the template controller to sync the latest version
	h.templateController.Enqueue(tp.Namespace, tp.Name)
	return nil, nil
}

// templateVersionByCreationTimestamp sorts a list of TemplateVersion by creation timestamp, using their names as a tie breaker.
type templateVersionByCreationTimestamp []*mlv1.ModelTemplateVersion

func (o templateVersionByCreationTimestamp) Len() int      { return len(o) }
func (o templateVersionByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o templateVersionByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[j].CreationTimestamp.Before(&o[i].CreationTimestamp)
}
