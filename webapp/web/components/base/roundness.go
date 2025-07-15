package base

import (
	"fmt"
	"strings"
)

type Roundness string

const (
	RoundedXs   Roundness = "xs"
	RoundedSm             = "sm"
	RoundedMd             = "md"
	RoundedLg             = "lg"
	RoundedXl             = "xl"
	Rounded2Xl            = "2xl"
	Rounded3Xl            = "3xl"
	Rounded4Xl            = "4xl"
	RoundedFull           = "full"
)

var corners = []string{"tl", "tr", "br", "bl"}

type Rounded struct {
	All     *Roundness
	corners [4]*Roundness
}

func (r Rounded) String() string {
	if r.All != nil {
		return fmt.Sprintf("rounded-%v", *r.All)
	}

	classes := []string{}
	for i, c := range r.corners {
		if c != nil {
			classes = append(classes, fmt.Sprintf("rounded-%s-%v", corners[i], *c))
		}
	}

	return strings.Join(classes, " ")
}

type RoundedBuilder struct {
	builder *Builder
	rounded Rounded
}

func (b *RoundedBuilder) All(r Roundness) *Builder {
	b.rounded.All = new(Roundness)
	b.rounded.All = &r

	b.builder.rounded = &b.rounded

	return b.builder
}

func (b *RoundedBuilder) TopLeft(r Roundness) *RoundedBuilder {
	b.rounded.corners[0] = new(Roundness)
	b.rounded.corners[0] = &r

	return b
}

func (b *RoundedBuilder) TopRight(r Roundness) *RoundedBuilder {
	b.rounded.corners[1] = new(Roundness)
	b.rounded.corners[1] = &r

	return b
}

func (b *RoundedBuilder) BottomRight(r Roundness) *RoundedBuilder {
	b.rounded.corners[2] = new(Roundness)
	b.rounded.corners[2] = &r

	return b
}

func (b *RoundedBuilder) BottomLeft(r Roundness) *RoundedBuilder {
	b.rounded.corners[3] = new(Roundness)
	b.rounded.corners[3] = &r

	return b
}

func (b *RoundedBuilder) Set() *Builder {
	b.builder.rounded = &b.rounded

	return b.builder
}
