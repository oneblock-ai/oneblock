package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	dashboardauthapi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

func NewMiddleware(management *config.Management) *Middleware {
	return &Middleware{
		tokenManager: management.TokenManager,
	}
}

type Middleware struct {
	tokenManager dashboardauthapi.TokenManager
}

func (m *Middleware) AuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		jweToken, err := extractJWETokenFromRequest(req)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(ResponseBody(ErrorResponse{Errors: []string{err.Error()}}))
			return
		}

		userInfo, err := m.getUserInfoFromToken(jweToken)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(ResponseBody(ErrorResponse{Errors: []string{err.Error()}}))
			return
		}

		ctx := request.WithUser(req.Context(), userInfo)
		req = req.WithContext(ctx)
		handler.ServeHTTP(rw, req)
	})
}

func extractJWETokenFromRequest(req *http.Request) (string, error) {
	tokenStr := req.Header.Get("Authorization")
	if strings.HasPrefix(tokenStr, "Bearer ") {
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	} else {
		tokenStr = ""
	}

	if tokenStr == "" {
		cookie, err := req.Cookie(cookieName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			return tokenStr, err
		} else if !errors.Is(err, http.ErrNoCookie) && len(cookie.Value) > 0 {
			tokenStr = cookie.Value
		}
	}

	if tokenStr == "" {
		return "", errors.New("failed to get cookie from request")
	}

	decodedToken, err := url.QueryUnescape(tokenStr)
	if err != nil {
		return "", errors.New("failed to parse cookie from request")
	}
	return decodedToken, nil
}

func (m *Middleware) getUserInfoFromToken(jweToken string) (userInfo user.Info, err error) {
	//handle panic from calling tokenManager.Decrypt
	defer func() {
		if recoveryMessage := recover(); recoveryMessage != nil {
			err = fmt.Errorf("%v", recoveryMessage)
		}
	}()

	authInfo, err := m.tokenManager.Decrypt(jweToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}
	return impersonateAuthInfoToUserInfo(authInfo), nil
}

func impersonateAuthInfoToUserInfo(authInfo *clientcmdapi.AuthInfo) user.Info {
	var userInfo user.DefaultInfo
	if authInfo.Impersonate != "" {
		userInfo.Name = authInfo.Impersonate
	}

	if len(authInfo.ImpersonateGroups) != 0 {
		userInfo.Groups = authInfo.ImpersonateGroups
	}

	return &userInfo
}
