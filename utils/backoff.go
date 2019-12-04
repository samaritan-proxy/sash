// Copyright 2019 Samaritan Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"time"

	"github.com/cenkalti/backoff/v3"
)

type ExponentialBackoffBuilder struct {
	initialInterval     *time.Duration
	maxInterval         *time.Duration
	maxElapsedTime      *time.Duration
	randomizationFactor *float64
	multiplier          *float64
	maxRetries          *uint64
	clock               backoff.Clock
}

func NewExponentialBackoffBuilder() *ExponentialBackoffBuilder {
	return &ExponentialBackoffBuilder{}
}

func (builder *ExponentialBackoffBuilder) InitialInterval(d time.Duration) *ExponentialBackoffBuilder {
	builder.initialInterval = &d
	return builder
}

func (builder *ExponentialBackoffBuilder) MaxInterval(d time.Duration) *ExponentialBackoffBuilder {
	builder.maxInterval = &d
	return builder
}

func (builder *ExponentialBackoffBuilder) MaxElapsedTime(d time.Duration) *ExponentialBackoffBuilder {
	builder.maxElapsedTime = &d
	return builder
}

func (builder *ExponentialBackoffBuilder) RandomizationFactor(f float64) *ExponentialBackoffBuilder {
	builder.randomizationFactor = &f
	return builder
}

func (builder *ExponentialBackoffBuilder) Multiplier(f float64) *ExponentialBackoffBuilder {
	builder.multiplier = &f
	return builder
}

func (builder *ExponentialBackoffBuilder) Clock(c backoff.Clock) *ExponentialBackoffBuilder {
	builder.clock = c
	return builder
}

func (builder *ExponentialBackoffBuilder) MaxRetries(i uint64) *ExponentialBackoffBuilder {
	builder.maxRetries = &i
	return builder
}

func (builder *ExponentialBackoffBuilder) Build() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	defer b.Reset()

	if builder.initialInterval != nil {
		b.InitialInterval = *builder.initialInterval
	}
	if builder.randomizationFactor != nil {
		b.RandomizationFactor = *builder.randomizationFactor
	}
	if builder.multiplier != nil {
		b.Multiplier = *builder.multiplier
	}
	if builder.maxInterval != nil {
		b.MaxInterval = *builder.maxInterval
	}
	if builder.maxElapsedTime != nil {
		b.MaxElapsedTime = *builder.maxElapsedTime
	}
	if builder.clock != nil {
		b.Clock = builder.clock
	}
	if builder.maxRetries != nil {
		return backoff.WithMaxRetries(b, *builder.maxRetries)
	}
	return b
}
