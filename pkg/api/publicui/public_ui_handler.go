package publicui

import (
	"net/http"

	"github.com/oneblock-ai/oneblock/pkg/settings"
	"github.com/oneblock-ai/oneblock/pkg/utils"
)

type Handler struct {
}

func NewPublicHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	utils.ResponseOKWithBody(rw, map[string]string{
		settings.UIPlSettingName:                  settings.UIPl.Get(),
		settings.UISourceSettingName:              getUISource(),
		settings.DefaultNotebookImagesSettingName: settings.NotebookDefaultImages.Get(),
		settings.DefaultRayVersion:                settings.RayVersion.Get(),
		settings.DefaultRayClusterImage:           settings.RayClusterImage.Get(),
		settings.FirstLoginSettingName:            settings.FirstLogin.Get(),
	})
}

func getUISource() string {
	uiSource := settings.UISource.Get()
	if uiSource == "auto" {
		if !settings.IsRelease() {
			uiSource = "external"
		} else {
			uiSource = "bundled"
		}
	}
	return uiSource
}
