package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GeneratedSecretType string

const (
	UndefinedType    GeneratedSecretType = ""
	Base64Type       GeneratedSecretType = "Base64"
	Base64URLType    GeneratedSecretType = "Base64URL"
	HexType          GeneratedSecretType = "Hex"
	AlphanumericType GeneratedSecretType = "Alphanumeric"
	AlphabeticType   GeneratedSecretType = "Alphabetic"
	UpperType        GeneratedSecretType = "Upper"
	UpperNumericType GeneratedSecretType = "UpperNumeric"
	LowerType        GeneratedSecretType = "Lower"
	LowerNumericType GeneratedSecretType = "LowerNumeric"
	NumericType      GeneratedSecretType = "Numeric"
	UUIDType         GeneratedSecretType = "UUID"
	DNSLabelType     GeneratedSecretType = "DNSLabel"
	StringType       GeneratedSecretType = "String"
)

// GeneratedSecretSpec defines the desired state of GeneratedSecret
type GeneratedSecretSpec struct {
	Keys                 []GeneratedSecretKeySpec `json:"keys,omitempty"`
	DeleteSecretOnDelete bool                     `json:"deleteSecretOnDelete,omitempty"`
}

type GeneratedSecretKeySpec struct {
	Name   string              `json:"name,omitempty"`
	Type   GeneratedSecretType `json:"type,omitempty"`
	Length int                 `json:"length,omitempty"`

	// Additional options can be provided for some types. Length
	// is supported for all types except it is ignored for Integers
	// and UUID.s
	String *GeneratedSecretKeySpecString `json:"string,omitempty"`
	Int    *GeneratedSecretKeySpecInt    `json:"int,omitempty"`
	Int64  *GeneratedSecretKeySpecInt64  `json:"int64,omitempty"`
}

type GeneratedSecretKeySpecString struct {
	Charset string `json:"charset,omitempty"`
}

type GeneratedSecretKeySpecInt struct {
	Max int `json:"max,omitempty"`
}

type GeneratedSecretKeySpecInt64 struct {
	Max int64 `json:"max,omitempty"`
}

// GeneratedSecretStatus defines the observed state of GeneratedSecret
type GeneratedSecretStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GeneratedSecret is the Schema for the generatedsecrets API
type GeneratedSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GeneratedSecretSpec   `json:"spec,omitempty"`
	Status GeneratedSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GeneratedSecretList contains a list of GeneratedSecret
type GeneratedSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GeneratedSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GeneratedSecret{}, &GeneratedSecretList{})
}
