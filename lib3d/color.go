package lib3d

import "strconv"

type colorState struct {
	R, G, B, A float32
}

func parseColorHex(hex string) colorState {
	c := colorState{1, 1, 1, 1}
	h := hex
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}

	if len(h) == 3 {
		buf := string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		n, err := strconv.ParseUint(buf, 16, 32)
		if err != nil {
			return c
		}
		c.R = float32((n>>16)&0xff) / 255
		c.G = float32((n>>8)&0xff) / 255
		c.B = float32(n&0xff) / 255
		return c
	}

	if len(h) == 6 {
		n, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			return c
		}
		c.R = float32((n>>16)&0xff) / 255
		c.G = float32((n>>8)&0xff) / 255
		c.B = float32(n&0xff) / 255
		return c
	}

	return c
}

func parseColorRGB(r, g, b float32) colorState {
	return colorState{r / 255, g / 255, b / 255, 1}
}

func parseColorRGBA(r, g, b, a float32) colorState {
	return colorState{r / 255, g / 255, b / 255, a / 255}
}
