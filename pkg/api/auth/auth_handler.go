package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	dashboardauthapi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	"github.com/oneblock-ai/apiserver/v2/pkg/apierror"
	"github.com/pkg/errors"
	"github.com/rancher/wrangler/v2/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	managementv1 "github.com/oneblock-ai/oneblock/pkg/apis/management.oneblock.ai/v1"
	ctlmanagementv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/management.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/indexeres"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils"
)

type Login struct {
	// local auth uses username and password
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	cookieName       = "OB_SESS"
	actionQuery      = "action"
	loginActionName  = "login"
	logoutActionName = "logout"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Handler struct {
	userCache    ctlmanagementv1.UserCache
	tokenManager dashboardauthapi.TokenManager
}

func NewAuthHandler(mgmt *config.Management) *Handler {
	return &Handler{
		userCache:    mgmt.OneBlockMgmtFactory.Management().V1().User().Cache(),
		tokenManager: mgmt.TokenManager,
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ResponseError(rw, http.StatusMethodNotAllowed, fmt.Errorf("only POST method is supported"))
		return
	}

	action := strings.ToLower(r.URL.Query().Get(actionQuery))
	if action == logoutActionName {
		// erase the cookie
		tokenCookie := &http.Cookie{
			Name:    cookieName,
			Value:   "",
			Path:    "/",
			MaxAge:  -1,
			Expires: time.Unix(1, 0), //January 1, 1970 UTC
		}
		http.SetCookie(rw, tokenCookie)
		utils.ResponseOKWithBody(rw, []byte("success logout"))
		return
	}

	if action != loginActionName {
		rw.WriteHeader(http.StatusBadRequest)
		utils.ResponseError(rw, http.StatusBadRequest, fmt.Errorf("unsupported action"))
		return
	}

	var input Login
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.ResponseError(rw, http.StatusBadRequest, fmt.Errorf("failed to decode request body, %s", err.Error()))
		return
	}

	tokenResp, err := h.login(&input)
	var header int
	if err != nil {
		var e *apierror.APIError
		if errors.As(err, &e) {
			header = e.Code.Status
		} else {
			header = http.StatusInternalServerError
		}
		utils.ResponseError(rw, header, err)
		return
	}

	tokenCookie := &http.Cookie{
		Name:  cookieName,
		Value: tokenResp,
		Path:  "/",
	}

	http.SetCookie(rw, tokenCookie)
	rw.Header().Set("Content-type", "application/json")
	utils.ResponseOKWithBody(rw, "login success")
}

func (h *Handler) login(input *Login) (token string, err error) {
	// handle panic from calling tokenManager.Generate
	defer func() {
		if recoveryMessage := recover(); recoveryMessage != nil {
			logrus.Errorf("failed to generate token: %v", recoveryMessage)
		}
	}()

	authInfo, err := h.userLogin(input)
	if err != nil {
		return "", err
	}

	token, err = h.tokenManager.Generate(*authInfo)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to generate token")
	}

	escapedToken := url.QueryEscape(token)
	return escapedToken, nil
}

func (h *Handler) userLogin(input *Login) (*clientcmdapi.AuthInfo, error) {
	username := input.Username
	pwd := input.Password

	user, err := h.getUser(username)
	if err != nil {
		return nil, apierror.NewAPIError(validation.Unauthorized, err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pwd)); err != nil {
		logrus.Warnf("invalid password, error: %v", err)
		return nil, apierror.NewAPIError(validation.Unauthorized, "authentication failed")
	}

	if user.IsAdmin {
		return &clientcmdapi.AuthInfo{
			Impersonate: user.Name,
		}, nil
	}

	return &clientcmdapi.AuthInfo{ImpersonateGroups: []string{"system:unauthenticated"}}, nil
}

func (h *Handler) getUser(username string) (*managementv1.User, error) {
	objs, err := h.userCache.GetByIndex(indexeres.UserNameIndex, username)
	if err != nil {
		return nil, err
	}
	if len(objs) == 0 {
		return nil, errors.New("authentication failed")
	}
	if len(objs) > 1 {
		return nil, errors.New("found more than one users with username " + username)
	}
	return objs[0], nil
}
