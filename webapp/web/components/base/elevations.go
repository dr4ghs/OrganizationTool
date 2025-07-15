package base

import (
	"fmt"
	"log"
)

var elevations = map[int]Elevation{
	0:  {0, 0, "none"},
	1:  {1, 5, "xs"},
	2:  {2, 7, "sm"},
	3:  {3, 8, "md"},
	4:  {4, 9, "md"},
	6:  {6, 11, "lg"},
	8:  {8, 12, "lg"},
	12: {12, 14, "xl"},
	16: {16, 15, "xl"},
	24: {24, 16, "2xl"},
}

type Elevation struct {
	Dp           int
	OpacityValue int
	ShadowValue  string
}

func NewElevation(dp int) Elevation {
	if _, ok := elevations[dp]; !ok {
		log.Panicf("Cannot have an elevation of %d", dp)
	}

	return elevations[dp]
}

func (e *Elevation) Opacity() string {
	return fmt.Sprintf("opacity-[.%02d]", e.OpacityValue)
}

func (e *Elevation) Shadow() string {
	return fmt.Sprintf("drop-shadow-%s", e.ShadowValue)
}
