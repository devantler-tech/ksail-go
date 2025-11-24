package configmanager

import (
	"fmt"
	"reflect"
	"time"

	mapstructure "github.com/go-viper/mapstructure/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// metav1DurationDecodeHook converts duration strings (e.g. "1m", "30s") into metav1.Duration values
// so that string values in ksail.yaml or environment variables are accepted.
func metav1DurationDecodeHook() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		durationType := reflect.TypeOf(metav1.Duration{})
		pointerDurationType := reflect.TypeOf(&metav1.Duration{})

		if to != durationType && to != pointerDurationType {
			return data, nil
		}

		if from.Kind() != reflect.String {
			return data, nil
		}

		raw, ok := data.(string)
		if !ok {
			return data, nil
		}

		if raw == "" {
			if to == pointerDurationType {
				return &metav1.Duration{}, nil
			}

			return metav1.Duration{}, nil
		}

		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parse duration %q: %w", raw, err)
		}

		durationValue := metav1.Duration{Duration: parsed}

		if to == pointerDurationType {
			return &durationValue, nil
		}

		return durationValue, nil
	}
}
