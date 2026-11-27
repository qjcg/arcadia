package pilot

import "math"

// stepCamera spirals around the origin while moving radially in the
// selected zoom direction. The frontend projects the resulting radius.
func (a *AutoZoom) stepCamera(v View) View {
	if a.speed <= 0 || a.speed == 1 {
		return v
	}
	if a.direction > 0 {
		v.Zoom *= a.speed
	} else {
		v.Zoom /= a.speed
	}
	c := &v.Camera
	angle := math.Atan2(c.Z, c.X) + 0.012
	radius := math.Hypot(c.X, c.Z)
	if radius == 0 {
		radius = 1
	}
	c.X, c.Z = radius*math.Cos(angle), radius*math.Sin(angle)
	c.Y += (0.3*radius*math.Sin(angle) - c.Y) * 0.05
	c.Yaw = math.Atan2(-c.X, -c.Z)
	c.Pitch = -math.Atan2(c.Y, radius)
	return v
}
