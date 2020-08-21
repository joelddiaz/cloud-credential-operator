/*
Copyright 2018 The OpenShift Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package codec

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	credreqv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	cloudtypesv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudtypes/v1"
)

// NewScheme creates a new Scheme
func NewScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	if err := cloudtypesv1.SchemeBuilder.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := credreqv1.SchemeBuilder.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}

// ProviderCodec is a runtime codec for providers.
// +k8s:deepcopy-gen=false
type ProviderCodec struct {
	encoder runtime.Encoder
	decoder runtime.Decoder
}

// NewCodec creates a serializer/deserializer for the provider configuration
func NewCodec() (*ProviderCodec, error) {
	scheme, err := NewScheme()
	if err != nil {
		return nil, err
	}
	codecFactory := serializer.NewCodecFactory(scheme)
	encoder, err := newEncoder(&codecFactory)
	if err != nil {
		return nil, err
	}
	codec := ProviderCodec{
		encoder: encoder,
		decoder: codecFactory.UniversalDecoder(cloudtypesv1.SchemeGroupVersion, credreqv1.SchemeGroupVersion),
	}
	return &codec, nil
}

// EncodeProvider serializes an object to the provider spec.
func (codec *ProviderCodec) EncodeProviderSpec(in runtime.Object) (*runtime.RawExtension, error) {
	var buf bytes.Buffer
	if err := codec.encoder.Encode(in, &buf); err != nil {
		return nil, fmt.Errorf("encoding failed: %v", err)
	}
	return &runtime.RawExtension{Raw: buf.Bytes()}, nil
}

// DecodeProviderSpec deserializes an object from the provider config.
func (codec *ProviderCodec) DecodeProviderSpec(providerConfig *runtime.RawExtension, out runtime.Object) error {
	_, _, err := codec.decoder.Decode(providerConfig.Raw, nil, out)
	if err != nil {
		return fmt.Errorf("decoding failure: %v", err)
	}
	return nil
}

// EncodeProviderStatus serializes the provider status.
func (codec *ProviderCodec) EncodeProviderStatus(in runtime.Object) (*runtime.RawExtension, error) {
	var buf bytes.Buffer
	if err := codec.encoder.Encode(in, &buf); err != nil {
		return nil, fmt.Errorf("encoding failed: %v", err)
	}
	return &runtime.RawExtension{Raw: buf.Bytes()}, nil
}

// DecodeProviderStatus deserializes the provider status.
func (codec *ProviderCodec) DecodeProviderStatus(providerStatus *runtime.RawExtension, out runtime.Object) error {
	if providerStatus != nil {
		_, _, err := codec.decoder.Decode(providerStatus.Raw, nil, out)
		if err != nil {
			return fmt.Errorf("decoding failure: %v", err)
		}
		return nil
	}
	return nil
}

func newEncoder(codecFactory *serializer.CodecFactory) (runtime.Encoder, error) {
	serializerInfos := codecFactory.SupportedMediaTypes()
	if len(serializerInfos) == 0 {
		return nil, fmt.Errorf("unable to find any serlializers")
	}

	encoder := codecFactory.EncoderForVersion(serializerInfos[0].Serializer, cloudtypesv1.SchemeGroupVersion)

	return encoder, nil
}
