/*
Copyright 2020 The Flux authors

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

package metrics

import (
	"fmt"
	"time"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
)

// Recorder records the canary analysis as Prometheus metrics
type Recorder struct {
	info                          *prometheus.GaugeVec
	duration                      *prometheus.HistogramVec
	total                         *prometheus.GaugeVec
	status                        *prometheus.GaugeVec
	phase                         *prometheus.GaugeVec
	webhookConfirmRollout         *prometheus.GaugeVec
	webhookConfirmTrafficIncrease *prometheus.GaugeVec
	webhookConfirmPromotion       *prometheus.GaugeVec
	weight                        *prometheus.GaugeVec
}

type WebhookStatus int

const (
	WebhookStatusSuccess WebhookStatus = iota //0
	WebhookStatusFailed
)

// NewRecorder creates a new recorder and registers the Prometheus metrics
func NewRecorder(controller string, register bool) Recorder {
	info := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "info",
		Help:      "Flagger version and mesh provider information",
	}, []string{"version", "mesh_provider"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: controller,
		Name:      "canary_duration_seconds",
		Help:      "Seconds spent performing canary analysis.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"name", "namespace"})

	total := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_total",
		Help:      "Total number of canary object",
	}, []string{"namespace"})

	// 0 - running, 1 - successful, 2 - failed
	status := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_status",
		Help:      "Last canary analysis result",
	}, []string{"name", "namespace"})

	// see pkg/apis/flagger/v1beta1/status.go
	phase := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_phase",
		Help:      "Condition of a canary at the current time",
	}, []string{"name", "namespace"})

	webhookConfirmRollout := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_webhook_confirm_rollout",
		Help:      "greater than 0 if confirm_rollout webhook failed",
	}, []string{"name", "namespace"})

	webhookConfirmTrafficIncrease := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_webhook_confirm_traffic_increase",
		Help:      "greater than 0 if confirm_traffic_increase webhook failed",
	}, []string{"name", "namespace"})

	webhookConfirmPromotion := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_webhook_confirm_promotion",
		Help:      "greater than 0 if confirm_promotion webhook failed",
	}, []string{"name", "namespace"})

	weight := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: controller,
		Name:      "canary_weight",
		Help:      "The virtual service destination weight current value",
	}, []string{"workload", "namespace"})

	if register {
		prometheus.MustRegister(info)
		prometheus.MustRegister(duration)
		prometheus.MustRegister(total)
		prometheus.MustRegister(status)
		prometheus.MustRegister(phase)
		prometheus.MustRegister(webhookConfirmRollout)
		prometheus.MustRegister(webhookConfirmTrafficIncrease)
		prometheus.MustRegister(webhookConfirmPromotion)
		prometheus.MustRegister(weight)
	}

	return Recorder{
		info:                          info,
		duration:                      duration,
		total:                         total,
		status:                        status,
		phase:                         phase,
		webhookConfirmRollout:         webhookConfirmRollout,
		webhookConfirmTrafficIncrease: webhookConfirmTrafficIncrease,
		webhookConfirmPromotion:       webhookConfirmPromotion,
		weight:                        weight,
	}
}

// SetInfo sets the version and mesh provider labels
func (cr *Recorder) SetInfo(version string, meshProvider string) {
	cr.info.WithLabelValues(version, meshProvider).Set(1)
}

// SetDuration sets the time spent in seconds performing canary analysis
func (cr *Recorder) SetDuration(cd *flaggerv1.Canary, duration time.Duration) {
	cr.duration.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Observe(duration.Seconds())
}

// SetTotal sets the total number of canaries per namespace
func (cr *Recorder) SetTotal(namespace string, total int) {
	cr.total.WithLabelValues(namespace).Set(float64(total))
}

// SetStatus sets the last known canary analysis status
func (cr *Recorder) SetStatus(cd *flaggerv1.Canary, phase flaggerv1.CanaryPhase) {
	var status int
	switch phase {
	case flaggerv1.CanaryPhaseProgressing:
		status = 0
	case flaggerv1.CanaryPhaseFailed:
		status = 2
	default:
		status = 1
	}
	cr.status.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Set(float64(status))
}

//  sets the webhook status
func (cr *Recorder) SetWebhookConfirmTrafficIncrease(cd *flaggerv1.Canary, status WebhookStatus) {
	cr.webhookConfirmTrafficIncrease.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Set(float64(status))
}

//  sets the webhook status
func (cr *Recorder) SetWebhookConfirmRollout(cd *flaggerv1.Canary, status WebhookStatus) {
	cr.webhookConfirmRollout.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Set(float64(status))
}

//  sets the webhook status
func (cr *Recorder) SetWebhookConfirmPromotion(cd *flaggerv1.Canary, status WebhookStatus) {
	cr.webhookConfirmPromotion.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Set(float64(status))
}

// SetPhase sets the last known condition of a canary at the current time
func (cr *Recorder) SetPhase(cd *flaggerv1.Canary, phase flaggerv1.CanaryPhase) {
	type CanaryPhase int
	const (
		Initializing     CanaryPhase = iota //0
		Initialized                         //1
		Waiting                             //2
		Progressing                         //3
		WaitingPromotion                    //4
		Promoting                           //5
		Finalising                          //6
		Succeeded                           //7
		Failed                              //8
		Terminating                         //9
		Terminated                          //10
	)
	var canaryPhase CanaryPhase
	switch phase {
	case flaggerv1.CanaryPhaseInitializing:
		canaryPhase = Initializing
	case flaggerv1.CanaryPhaseInitialized:
		canaryPhase = Initialized
	case flaggerv1.CanaryPhaseWaiting:
		canaryPhase = Waiting
	case flaggerv1.CanaryPhaseProgressing:
		canaryPhase = Progressing
	case flaggerv1.CanaryPhaseWaitingPromotion:
		canaryPhase = WaitingPromotion
	case flaggerv1.CanaryPhasePromoting:
		canaryPhase = Promoting
	case flaggerv1.CanaryPhaseFinalising:
		canaryPhase = Finalising
	case flaggerv1.CanaryPhaseSucceeded:
		canaryPhase = Succeeded
	case flaggerv1.CanaryPhaseFailed:
		canaryPhase = Failed
	case flaggerv1.CanaryPhaseTerminating:
		canaryPhase = Terminating
	case flaggerv1.CanaryPhaseTerminated:
		canaryPhase = Terminated
	default:
		canaryPhase = Progressing
	}
	cr.phase.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Set(float64(canaryPhase))
}

// SetWeight sets the weight values for primary and canary destinations
func (cr *Recorder) SetWeight(cd *flaggerv1.Canary, primary int, canary int) {
	cr.weight.WithLabelValues(fmt.Sprintf("%s-primary", cd.Spec.TargetRef.Name), cd.Namespace).Set(float64(primary))
	cr.weight.WithLabelValues(cd.Spec.TargetRef.Name, cd.Namespace).Set(float64(canary))
}
