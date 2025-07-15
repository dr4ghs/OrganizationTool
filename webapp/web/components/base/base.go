package base

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

type ComponentType string

const (
	Primary    ComponentType = "primary"
	Variant                  = "variant"
	Secondary                = "secondary"
	Background               = "background"
	Surface                  = "surface"
	Error                    = "error"
)

type Builder struct {
	id         *string
	typ        ComponentType
	elevation  Elevation
	shadow     bool
	rounded    *Rounded
	outerClass []string
	innerClass []string
	attrs      templ.Attributes
}

func New() (b *Builder) {
	return &Builder{
		outerClass: make([]string, 0),
		innerClass: make([]string, 0),
		attrs:      templ.Attributes{},
	}
}

func (b *Builder) Copy() *Builder {
	result := new(Builder)
	*result = *b

	return result
}

func (b *Builder) Type(typ ComponentType) *Builder {
	b.typ = typ

	return b
}

func (b *Builder) Id(id string) *Builder {
	b.id = &id

	return b
}

func (b *Builder) Elevation(val int) *Builder {
	b.elevation = NewElevation(val)

	return b
}

func (b *Builder) Shadow() *Builder {
	b.shadow = true

	return b
}

func (b *Builder) Rounded() *RoundedBuilder {
	b.rounded = &Rounded{}

	return &RoundedBuilder{
		builder: b,
	}
}

func (b *Builder) AddInnerClass(class string) *Builder {
	b.innerClass = append(b.innerClass, class)

	return b
}

func (b *Builder) AddOuterClass(class string) *Builder {
	b.outerClass = append(b.outerClass, class)

	return b
}

func (b *Builder) Attribute(key string, value any) *Builder {
	b.attrs[key] = value

	return b
}

func (b Builder) OuterClass() string {
	class := fmt.Sprintf("group relative shadow-black bg-%s text-on-%s", b.typ, b.typ)

	if b.rounded != nil {
		class = fmt.Sprintf("%s %s", class, b.rounded)
	}

	if b.shadow {
		class = fmt.Sprintf("%s %s", class, b.elevation.Shadow())
	}

	if len(b.outerClass) > 0 {
		class = fmt.Sprintf("%s %s", class, strings.Join(b.outerClass, " "))
	}

	return class
}

func (b Builder) InnerClass() string {
	class := fmt.Sprintf(
		"bg-white w-full h-full %s absolute pointer-events-none",
		b.elevation.Opacity(),
	)

	if b.rounded != nil {
		class = fmt.Sprintf("%s %s", class, b.rounded)
	}

	if len(b.innerClass) > 0 {
		class = fmt.Sprintf("%s %s", class, strings.Join(b.innerClass, " "))
	}

	return class
}
