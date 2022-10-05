// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package adapter

import (
	"context"
	"github.com/aws/private-amazon-cloudwatch-agent-staging/receiver/adapter/accumulator"
	"github.com/influxdata/telegraf/models"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type AdaptedReceiver struct {
	logger      *zap.Logger
	input       *models.RunningInput
	accumulator accumulator.OtelAccumulator
}

func newAdaptedReceiver(input *models.RunningInput, logger *zap.Logger) *AdaptedReceiver {
	return &AdaptedReceiver{
		input:  input,
		logger: logger,
	}
}

// Adapter Receiver uses Scrape Controller to scrape metric and has three phases:
// Start: Start the accumulator to initialize the logger and resources metric
// Scrape: Gather metrics using accumulator
// (e.g CPU https://github.com/influxdata/telegraf/blob/6e924fcd5cc2ce79a024b7275d865d7a19c455ed/plugins/inputs/cpu/cpu.go)
// Shutdown Stop the scarpper and flush the remaining metrics  before shutting down the scraper.
func (r *AdaptedReceiver) start(_ context.Context, _ component.Host) error {
	// TODO: Add Set Precision based on agent precision and agent interval
	// https://github.com/influxdata/telegraf/blob/3b3584b40b7c9ea10ae9cb02137fc072da202704/agent/agent.go#L316-L317
	r.accumulator = accumulator.NewAccumulator(r.input, r.logger)
	return nil
}

func (r *AdaptedReceiver) scrape(_ context.Context) (pmetric.Metrics, error) {
	r.logger.Debug("Begining scraping metrics with adapter", zap.String("receiver", r.input.Config.Name))

	if err := r.input.Input.Gather(r.accumulator); err != nil {
		r.accumulator.AddError(err)
		return pmetric.Metrics{}, err
	}

	return r.accumulator.GetOtelMetrics(), nil
}

func (r *AdaptedReceiver) shutdown(_ context.Context) error {
	r.logger.Debug("Shutdown adapter", zap.String("receiver", r.input.Config.Name))
	return nil
}