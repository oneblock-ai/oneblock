package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	k8sdashboardjwe "github.com/kubernetes/dashboard/src/app/backend/auth/jwe"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/oneblock-ai/oneblock/pkg/auth"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/settings"
)

const (
	privateKey = "priv"
	publicKey  = "pub"
)

func WatchSecret(ctx context.Context, mgmt *config.Management) error {
	name := settings.AuthSecretName.Get()
	secrets := mgmt.CoreFactory.Core().V1().Secret()
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	}
	watcher, err := secrets.Watch(mgmt.Namespace, opts)
	if err != nil {
		//logrus.Errorf("Failed to watch secret %s:%s, %v", mgmt.Namespace, name, err)
		return fmt.Errorf("failed to watch secret %s:%s, %v", mgmt.Namespace, name, err)
	}

	for {
		select {
		case watchEvent := <-watcher.ResultChan():
			if watch.Modified == watchEvent.Type {
				if sec, ok := watchEvent.Object.(*corev1.Secret); ok {
					if err := refreshKeyInTokenManager(sec, mgmt); err != nil {
						logrus.Errorf("Failed to update tokenManager with secret %s:%s, %v", mgmt.Namespace, name, err)
						continue
					}
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func refreshKeyInTokenManager(sec *corev1.Secret, mgmt *config.Management) (err error) {
	// handle panic from calling kubernetes dashboard tokenManager.Decrypt
	defer func() {
		if recoveryMessage := recover(); recoveryMessage != nil {
			err = fmt.Errorf("%v", recoveryMessage)
			logrus.Errorf("Failed to decrypt generated token with key from secret %s/%s, %v", sec.Namespace, sec.Name, recoveryMessage)
		}
	}()

	priv, err := k8sdashboardjwe.ParseRSAKey(string(sec.Data[privateKey]), string(sec.Data[publicKey]))
	if err != nil {
		return errors.Wrapf(err, "Failed to parse rsa key from secret %s/%s", sec.Namespace, sec.Name)
	}

	encrypter, err := jose.NewEncrypter(jose.A256GCM, jose.Recipient{Algorithm: jose.RSA_OAEP_256, Key: &priv.PublicKey}, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to create jose encrypter")
	}

	add, err := getAdd()
	if err != nil {
		return err
	}

	jwtEncryption, err := encrypter.EncryptWithAuthData([]byte(`{}`), add)
	if err != nil {
		return errors.Wrapf(err, "Failed to encrypt with key from secret %s/%s", sec.Namespace, sec.Name)
	}

	// token manager will refresh the key if decrypt failed
	_, err = mgmt.TokenManager.Decrypt(jwtEncryption.FullSerialize())
	if err != nil {
		return errors.Wrapf(err, "Failed to decrypt generated token with key from secret %s/%s", sec.Namespace, sec.Name)
	}
	return
}

func getAdd() ([]byte, error) {
	now := time.Now()
	claim := map[string]string{
		"iat": now.Format(time.RFC3339),
		"exp": now.Add(auth.GetTokenMaxTTL()).Format(time.RFC3339),
	}
	add, err := json.Marshal(claim)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to marshal jwe claim")
	}
	return add, nil
}
