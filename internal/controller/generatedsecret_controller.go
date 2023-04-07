package controller

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/jimeh/rands"
	secretsv1 "github.com/krystal/generated-secrets/api/v1"
)

// GeneratedSecretReconciler reconciles a GeneratedSecret object
type GeneratedSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=secrets.k8s.k.io,resources=generatedsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secrets.k8s.k.io,resources=generatedsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secrets.k8s.k.io,resources=generatedsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete

const (
	ownerAnnotationKey = "secrets.k8s.k.io/owner"
	finalizerName      = "secrets.k8s.k.io/finalizer"
)

func (r *GeneratedSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Firstly, let's get our generated secret from the API.
	var generatedSecret secretsv1.GeneratedSecret
	if err := r.Get(ctx, req.NamespacedName, &generatedSecret); err != nil {
		log.Error(err, "unable to fetch GeneratedSecret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Next, let's add a finalizer to our object if we don't have one because
	// we don't want this to be deleted until we've also tidied up the secret
	// if that was desirable.
	if !controllerutil.ContainsFinalizer(&generatedSecret, finalizerName) {
		log.Info("adding finalizer to GeneratedSecret")
		controllerutil.AddFinalizer(&generatedSecret, finalizerName)
		if err := r.Update(ctx, &generatedSecret); err != nil {
			log.Error(err, "unable to update GeneratedSecret")
			return ctrl.Result{}, err
		}
	}

	// Next, we need to see if we have an existing secret for this.
	// If not, we can create it but we do need to make sure we can copy over
	// any existing values when we do an update.
	var secret v1.Secret
	newSecret := false
	err := r.Get(ctx, req.NamespacedName, &secret)
	if errors.IsNotFound(err) {
		// If we don't already have a secret, we can create a new one
		// to get started with.
		newSecret = true
		secret = v1.Secret{
			ObjectMeta: ctrl.ObjectMeta{
				Name:      generatedSecret.Name,
				Namespace: generatedSecret.Namespace,
				Annotations: map[string]string{
					ownerAnnotationKey: fmt.Sprintf("%s/%s", generatedSecret.Namespace, generatedSecret.Name),
				},
			},
			Data: map[string][]byte{},
		}
	} else if err != nil {
		log.Error(err, "unable to fetch Secret")
		return ctrl.Result{}, err
	} else {
		// If an existing secret was found, we will be updating that one
		log.Info("Found existing secret")

		// If the generated secret was deleted, we need to delete the secret
		// if desired and remove the finalizer.
		if !generatedSecret.ObjectMeta.DeletionTimestamp.IsZero() {
			if generatedSecret.Spec.DeleteSecretOnDelete {
				err = r.Delete(ctx, &secret)
				if errors.IsNotFound(err) {
					log.Info("secret is aleady deleted")
				} else if err != nil {
					log.Error(err, "unable to delete Secret")
					return ctrl.Result{}, err
				} else {
					log.Info("deleted secret")
				}
			}

			log.Info("removing finalizer")
			controllerutil.RemoveFinalizer(&generatedSecret, finalizerName)
			if err := r.Update(ctx, &generatedSecret); err != nil {
				log.Info("removed finalizer")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	// Next, we can go through all the desired keys and set them
	// as appropriate. If they're already set, we will not do
	// anything as that value is fixed now.
	for _, keySpec := range generatedSecret.Spec.Keys {
		if secret.Data[keySpec.Name] != nil {
			log.Info("key already set, skipping", "key", keySpec.Name)
			continue
		}

		// Generaet a new random value for the key spec
		newValue, err := generateRandomValue(keySpec)
		if err != nil {
			log.Error(err, "unable to generate random value", "key", keySpec.Name)
			return ctrl.Result{}, err
		}

		// Store that in the secret
		secret.Data[keySpec.Name] = []byte(newValue)
		log.Info("setting value for key", "key", keySpec.Name)
	}

	// Save or create the secret as appropriate
	if newSecret {
		log.Info("creating secret")
		if err := r.Create(ctx, &secret); err != nil {
			log.Error(err, "unable to create Secret")
			return ctrl.Result{}, err
		}
	} else {
		log.Info("updating secret")
		if err := r.Update(ctx, &secret); err != nil {
			log.Error(err, "unable to update Secret")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GeneratedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1.GeneratedSecret{}).
		Complete(r)
}

func generateRandomValue(spec secretsv1.GeneratedSecretKeySpec) (string, error) {
	switch spec.Type {
	case secretsv1.Base64Type:
		return rands.Base64(spec.Length)
	case secretsv1.Base64URLType:
		return rands.Base64URL(spec.Length)
	case secretsv1.HexType:
		return rands.Hex(spec.Length)
	case secretsv1.AlphanumericType:
		return rands.Alphanumeric(spec.Length)
	case secretsv1.AlphabeticType:
		return rands.Alphabetic(spec.Length)
	case secretsv1.UpperType:
		return rands.Upper(spec.Length)
	case secretsv1.UpperNumericType:
		return rands.UpperNumeric(spec.Length)
	case secretsv1.LowerType:
		return rands.Lower(spec.Length)
	case secretsv1.LowerNumericType:
		return rands.LowerNumeric(spec.Length)
	case secretsv1.NumericType:
		return rands.Numeric(spec.Length)
	case secretsv1.UUIDType:
		return rands.UUID()
	case secretsv1.DNSLabelType:
		return rands.DNSLabel(spec.Length)
	case secretsv1.StringType:
		return rands.String(spec.Length, spec.String.Charset)
	default:
		return "", fmt.Errorf("invalid key type: %s", spec.Type)
	}
}
