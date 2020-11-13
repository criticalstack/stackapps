package controllers

import (
	"bytes"
	"reflect"
	"strings"
	"text/template"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/jsonpath"
)

func checkHealth(hc featuresv1alpha1.HealthCheck, appRevision *featuresv1alpha1.AppRevision, health *featuresv1alpha1.AppRevisionCondition) error {
	switch hc.Type {
	case featuresv1alpha1.HealthCheckTypeJSONPath:
		j := jsonpath.New("healthcheck")
		if err := j.Parse(hc.Value); err != nil {
			return errors.Wrap(err, "failed to parse JSONPath string")
		}
		res, err := j.FindResults(appRevision)
		if err != nil {
			return err
		}
		for _, vv := range res {
			for _, v := range vv {
				if isHealthy(v) {
					if health.Status == corev1.ConditionUnknown {
						health.Status = corev1.ConditionTrue
					}
				} else {
					health.Status = corev1.ConditionFalse
				}
			}
		}
	case featuresv1alpha1.HealthCheckTypeGoTemplate:
		t, err := template.New("healthcheck").Parse(hc.Value)
		if err != nil {
			return errors.Wrap(err, "failed to parse go-template string")
		}
		buf := new(bytes.Buffer)
		if err := t.Execute(buf, appRevision); err != nil {
			return errors.Wrap(err, "failed to execute go-template")
		}
		if buf.Len() == 0 {
			if health.Status == corev1.ConditionUnknown {
				health.Status = corev1.ConditionTrue
				health.Reason = ""
			}
		} else {
			health.Status = corev1.ConditionFalse
			health.Reason = "HealthCheck"
			if hc.Name != "" {
				health.Reason = hc.Name
			}
			health.Message = buf.String()
		}
	default:
		return errors.Errorf("unrecognized health check type %q", hc.Type)
	}
	return nil
}

func isHealthy(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	x := v.Interface()
	switch t := x.(type) {
	case bool:
		return t
	case corev1.ConditionStatus:
		return t == corev1.ConditionTrue
	case string:
		return strings.ToLower(strings.TrimSpace(t)) == "true"
	case error:
		return t == nil
	}
	switch v.Kind() {
	case reflect.Struct:
		return isHealthy(v.FieldByName("Status"))
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if !isHealthy(v.Index(i)) {
				return false
			}
		}
		return v.Len() > 0
	}
	if v.Kind() == reflect.Struct {
	}
	return false
}
