package controller_test

import (
	"path/filepath"
	"testing"

	secretsv1 "github.com/krystal/generated-secrets/api/v1"
	"github.com/krystal/generated-secrets/internal/controller"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type (
	testHarness struct {
		Manager ctrl.Manager
		testEnv *envtest.Environment
	}
)

func setupHarness(t *testing.T) *testHarness {
	t.Helper()

	harness := &testHarness{
		testEnv: &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
			CRDInstallOptions: envtest.CRDInstallOptions{
				CleanUpAfterUse: true,
			},
		},
	}

	cfg, err := harness.testEnv.Start()
	if err != nil {
		t.Fatalf("failed to start test environment: %v", err)
	}

	err = secretsv1.AddToScheme(scheme.Scheme)
	if err != nil {
		t.Fatalf("failed to add secretsv1 to scheme: %v", err)
	}

	//+kubebuilder:scaffold:scheme

	harness.Manager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})

	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = (&controller.GeneratedSecretReconciler{
		Client: harness.Manager.GetClient(),
		Scheme: harness.Manager.GetScheme(),
	}).SetupWithManager(harness.Manager)
	if err != nil {
		t.Fatalf("failed to setup reconciler: %v", err)
	}

	return harness
}
