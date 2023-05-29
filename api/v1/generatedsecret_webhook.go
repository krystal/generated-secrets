package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var generatedsecretlog = logf.Log.WithName("generatedsecret-resource")

func (r *GeneratedSecret) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-secrets-k8s-k-io-v1-generatedsecret,mutating=false,failurePolicy=fail,sideEffects=None,groups=secrets.k8s.k.io,resources=generatedsecrets,verbs=create;update,versions=v1,name=vgeneratedsecret.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &GeneratedSecret{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GeneratedSecret) ValidateCreate() error {
	generatedsecretlog.Info("validate create", "name", r.Name)

	return validateGeneratedSecret(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GeneratedSecret) ValidateUpdate(old runtime.Object) error {
	generatedsecretlog.Info("validate update", "name", r.Name)

	return validateGeneratedSecret(r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GeneratedSecret) ValidateDelete() error {
	generatedsecretlog.Info("validate delete", "name", r.Name)

	return nil
}

func validateGeneratedSecret(secret *GeneratedSecret) error {
	allErrs := field.ErrorList{}

	for i, key := range secret.Spec.Keys {
		if key.Name == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("keys").Index(i).Child("name"), "must be set"))
		}

		if key.Type == UndefinedType {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("keys").Index(i).Child("type"), "must be set"))
		} else if !validateTypeIsSupported(key.Type) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("keys").Index(i).Child("type"), key.Type, "must be one of the supported types"))
		}

		if typeNeedsLength(key.Type) == false && key.Length != 0 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("keys").Index(i).Child("length"), key.Length, "must not be set for type"))
		}

		if typeNeedsLength(key.Type) && key.Length == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("keys").Index(i).Child("length"), "must be set"))
		}

		if key.Type == StringType && (key.String == nil || key.String.Charset == "") {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("keys").Index(i).Child("string").Child("charset"), "must be set for string type"))
		}

	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "secrets.k8s.k.io", Kind: "GeneratedSecret"},
		secret.Name, allErrs)
}

func typeNeedsLength(t GeneratedSecretType) bool {
	switch t {
	case UUIDType, ECDSAKeyType:
		return false
	default:
		return true
	}
}

func validateTypeIsSupported(t GeneratedSecretType) bool {
	switch t {
	case Base64Type, Base64URLType, HexType, AlphanumericType, AlphabeticType, UpperType,
		UpperNumericType, LowerType, LowerNumericType, NumericType, UUIDType, DNSLabelType,
		StringType, ECDSAKeyType:
		return true
	default:
		return false
	}
}
