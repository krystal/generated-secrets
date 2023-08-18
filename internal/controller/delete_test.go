package controller_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	secretsv1 "github.com/krystal/generated-secrets/api/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func randomKey(keyName string) *secretsv1.GeneratedSecretKeySpec {
	secretTypes := []secretsv1.GeneratedSecretType{
		secretsv1.Base64Type,
		secretsv1.Base64URLType,
		secretsv1.HexType,
		secretsv1.AlphanumericType,
		secretsv1.AlphabeticType,
		secretsv1.UpperType,
		secretsv1.UpperNumericType,
		secretsv1.LowerType,
		secretsv1.LowerNumericType,
		secretsv1.NumericType,
		secretsv1.UUIDType,
		secretsv1.StringType,
		secretsv1.DNSLabelType,
		secretsv1.ECDSAKeyType,
	}

	return &secretsv1.GeneratedSecretKeySpec{
		Name:   keyName,
		Type:   secretTypes[rand.Intn(len(secretTypes))],
		Length: rand.Intn(64),
		String: &secretsv1.GeneratedSecretKeySpecString{
			Charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		Int: &secretsv1.GeneratedSecretKeySpecInt{
			Max: rand.Intn(64),
		},
		Int64: &secretsv1.GeneratedSecretKeySpecInt64{
			Max: rand.Int63n(64),
		},
	}
}

func TestDeleteGeneratedSecret_Multi(t *testing.T) {
	harness := setupHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	const (
		timeout  = time.Second * 3
		interval = time.Millisecond * 250
	)

	t.Cleanup(func() {
		cancel()
		assert.True(t, assert.Eventually(t, func() bool {
			err := harness.testEnv.Stop()
			return err == nil
		}, timeout, interval))
	})

	go func() {
		if err := harness.Manager.Start(ctx); err != nil {
			log.Printf("failed to start manager: %v", err)
		}
	}()

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("DeleteGeneratedSecret%d", i), func(t *testing.T) {
			i := i
			t.Parallel()
			secret := &secretsv1.GeneratedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("deletegeneratedsecret%d", i),
					Namespace: "default",
				},
				Spec: secretsv1.GeneratedSecretSpec{
					DeleteSecretOnDelete: rand.Intn(2) == 0,
					Keys:                 []secretsv1.GeneratedSecretKeySpec{},
				},
			}

			for key := 0; key < rand.Intn(10)+1; key++ {
				secret.Spec.Keys = append(secret.Spec.Keys, *randomKey(fmt.Sprintf("key%d", key)))
			}
			testDeleteGeneratedSecret(ctx, t, harness, secret)
		})
	}
}

func testDeleteGeneratedSecret(ctx context.Context, t *testing.T, harness *testHarness, generatedSecret *secretsv1.GeneratedSecret) {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	client := harness.Manager.GetClient()

	// Create a GeneratedSecret object.
	assert.NoError(t, client.Create(ctx, generatedSecret))

	// Wait for the secret to be created.

	lookupKey := types.NamespacedName{
		Name:      generatedSecret.ObjectMeta.Name,
		Namespace: generatedSecret.ObjectMeta.Namespace,
	}

	assert.True(t, assert.Eventually(t, func() bool {
		var secret secretsv1.GeneratedSecret
		err := client.Get(ctx, lookupKey, &secret)
		return err == nil
	}, timeout, interval))

	// Check that the secret was created.

	assert.True(t, assert.Eventually(t, func() bool {
		var secret corev1.Secret
		err := client.Get(ctx, lookupKey, &secret)
		// log.Println(err)
		return err == nil
	}, timeout, interval))

	// Delete the GeneratedSecret object.

	assert.NoError(t, client.Delete(ctx, generatedSecret))

	// Wait for the secret to be deleted.

	assert.True(t, assert.Eventually(t, func() bool {
		var secret corev1.Secret
		err := client.Get(ctx, lookupKey, &secret)
		if generatedSecret.Spec.DeleteSecretOnDelete {
			return apierrors.IsNotFound(err)
		}

		return err == nil
	}, timeout, interval))

	// Check that the GeneratedSecret object is gone.
	assert.True(t, assert.Eventually(t, func() bool {
		var secret secretsv1.GeneratedSecret
		err := client.Get(ctx, lookupKey, &secret)
		return apierrors.IsNotFound(err)
	}, timeout, interval))
}
