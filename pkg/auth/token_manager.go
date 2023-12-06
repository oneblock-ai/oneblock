package auth

import (
	"strconv"
	"time"

	dashboardapi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	dashboardjwt "github.com/kubernetes/dashboard/src/app/backend/auth/jwe"
	ctlcorev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"

	"github.com/oneblock-ai/oneblock/pkg/settings"
)

func NewJWETokenManager(secrets ctlcorev1.SecretClient, namespace string) (tokenManager dashboardapi.TokenManager, err error) {
	//handle panic from token manager
	defer func() {
		if recoveryMessage := recover(); recoveryMessage != nil {
			logrus.Fatalf("failed to create token manager: %v", recoveryMessage)
		}
	}()

	synchronizer := NewSecretSynchronizer(secrets, namespace, settings.AuthSecretName.Get())
	keyHolder := dashboardjwt.NewRSAKeyHolder(synchronizer)
	tokenManager = dashboardjwt.NewJWETokenManager(keyHolder)
	tokenManager.SetTokenTTL(GetTokenMaxTTL())
	return tokenManager, nil
}

func GetTokenMaxTTL() time.Duration {
	ttlStr := settings.AuthTokenMaxTTLMinutes.Get()
	ttl, err := strconv.ParseInt(ttlStr, 10, 32)
	if err != nil {
		ttl = 720
	}
	return time.Duration(ttl) * time.Minute
}
