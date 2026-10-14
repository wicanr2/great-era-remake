// Package atlasmap draws the presentation-only national relief map.
// Province adjacency, battlefield terrain and game rules remain in internal/game.
package atlasmap

import (
	_ "embed"
	"encoding/json"
	"image"
	"math"
	"sort"

	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
)

// Natural Earth 5.1.2, public domain. Source and hashes are recorded in
// ../../../../geography-sources.json. No Hearts of Iron assets are used.
//
//go:embed geography.json
var geographyBytes []byte

type Point [2]float64
type Geography struct {
	Bounds [4]float64
	Land   [][]Point
	Rivers [][]Point
}
type Camera struct{ Zoom, Tilt, PanX, PanY float64 }

func DefaultCamera() Camera { return Camera{1.9, 1.36, .12, .02} }
func (c *Camera) ZoomAt(factor, x, y float64, w, h int) {
	before := c.Zoom
	c.Zoom = math.Max(.65, math.Min(3.5, c.Zoom*factor))
	x = x/float64(w)*2 - 1
	y = 1 - y/float64(h)*2
	c.PanX = x - (x-c.PanX)*c.Zoom/before
	c.PanY = y - (y-c.PanY)*c.Zoom/before
}
func (c *Camera) Pan(dx, dy float64, w, h int) {
	c.PanX = math.Max(-2, math.Min(2, c.PanX+dx/float64(w)*2))
	c.PanY = math.Max(-2, math.Min(2, c.PanY-dy/float64(h)*2))
}

var Anchors = [...]Point{
	{123.9, 47.35}, {126.5, 43.9}, {130.4, 46.8}, {126.6, 50}, {123.9, 43.3}, {119.7, 49.2},
	{123.4, 41.8}, {128, 45.3}, {124.4, 40.1}, {118, 41.8}, {116.4, 39}, {115, 41},
	{106.9, 47.8}, {111.7, 40.8}, {106.2, 38.5}, {112.5, 37.9}, {103.8, 36.1}, {108.9, 34.3},
	{113.6, 34.7}, {117, 36.6}, {119.1, 33.5}, {117.3, 31.9}, {120.2, 30.3}, {119.3, 26.1},
	{115.9, 28.7}, {114.3, 30.6}, {113, 28.2}, {106.7, 26.6}, {104.1, 30.7}, {101.8, 36.6},
	{87.6, 43.8}, {91.1, 29.7}, {100.2, 30.1}, {102.7, 25}, {108.3, 23}, {113.3, 23.1},
	{121, 23.8}, {110.3, 19.6}, {96.1, 21.9},
}

const cols, rows, maskW, maskH = 160, 112, 768, 512

type vertex struct {
	lon, lat, height, land float64
	rgb                    [3]float64
	province               int
}
type surfacePoint struct {
	rgb      [3]byte
	province uint8
	border   bool
}

type Map struct {
	geo       Geography
	mask      []byte
	surface   []surfacePoint
	vertices  []vertex
	indices   []uint16
	projected []Point
	cached    *image.RGBA
	key       renderKey
	Renders   int
}
type renderKey struct {
	Camera           Camera
	W, H             int
	Retro, Political bool
	Factions         [39]int
}

func New() (*Map, error) {
	m := &Map{mask: make([]byte, maskW*maskH)}
	if err := json.Unmarshal(geographyBytes, &m.geo); err != nil {
		return nil, err
	}
	// Polygon scan conversion is done once, without allocating a large texture.
	for _, ring := range m.geo.Land {
		for y := 0; y < maskH; y++ {
			lat := m.geo.Bounds[3] - (float64(y)+.5)/maskH*(m.geo.Bounds[3]-m.geo.Bounds[2])
			xs := []float64{}
			for i, p := range ring {
				q := ring[(i+1)%len(ring)]
				if (p[1] > lat) != (q[1] > lat) {
					xs = append(xs, (p[0]+(lat-p[1])*(q[0]-p[0])/(q[1]-p[1])-m.geo.Bounds[0])/(m.geo.Bounds[1]-m.geo.Bounds[0])*maskW)
				}
			}
			sort.Float64s(xs)
			for i := 0; i+1 < len(xs); i += 2 {
				for x := max(0, int(math.Ceil(xs[i]-.5))); x < min(maskW, int(math.Ceil(xs[i+1]-.5))); x++ {
					m.mask[y*maskW+x] = 255
				}
			}
		}
	}
	for r := 0; r <= rows; r++ {
		for c := 0; c <= cols; c++ {
			lon := m.geo.Bounds[0] + float64(c)/cols*(m.geo.Bounds[1]-m.geo.Bounds[0])
			lat := m.geo.Bounds[2] + float64(r)/rows*(m.geo.Bounds[3]-m.geo.Bounds[2])
			e := m.Elevation(lon, lat)
			nx := -(m.Elevation(lon+.12, lat) - m.Elevation(lon-.12, lat)) / .24 * 5.5
			nz := -(m.Elevation(lon, lat+.12) - m.Elevation(lon, lat-.12)) / .24 * 5.5
			light := .33 + .78*math.Max(0, (-nx*.8+1+nz*.5)/(math.Sqrt(nx*nx+1+nz*nz)*1.375))
			color := mix([3]float64{.62, .55, .36}, [3]float64{.34, .45, .29}, clamp((lon-91)/30))
			color = mix(color, [3]float64{.51, .48, .40}, clamp((e-.18)/.42))
			color = mix(color, [3]float64{.89, .90, .85}, clamp((e-.53)/.35))
			for i := range color {
				color[i] = clamp(color[i]*light) * 255
			}
			nearest, best := 0, math.Inf(1)
			for i, p := range Anchors {
				d := (lon-p[0])*(lon-p[0]) + (lat-p[1])*(lat-p[1])*1.3
				if d < best {
					nearest, best = i, d
				}
			}
			m.vertices = append(m.vertices, vertex{lon, lat, e, m.landAt(lon, lat), color, nearest})
		}
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			p := uint16(r*(cols+1) + c)
			m.indices = append(m.indices, p, p+1, p+cols+1, p+1, p+cols+2, p+cols+1)
		}
	}
	m.projected = make([]Point, len(m.vertices))
	m.prepareSurface()
	return m, nil
}
func clamp(x float64) float64 { return math.Max(0, math.Min(1, x)) }
func mix(a, b [3]float64, t float64) [3]float64 {
	for i := range a {
		a[i] = a[i]*(1-t) + b[i]*t
	}
	return a
}
func hash(x, y int) float64 {
	n := uint32(x+1907)*374761393 ^ uint32(y+103)*668265263
	n = (n ^ (n >> 13)) * 1274126177
	return float64(n) / 4294967295
}
func noise(x, y float64) float64 {
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-math.Floor(x), y-math.Floor(y)
	fx = fx * fx * (3 - 2*fx)
	fy = fy * fy * (3 - 2*fy)
	return hash(ix, iy)*(1-fx)*(1-fy) + hash(ix+1, iy)*fx*(1-fy) + hash(ix, iy+1)*(1-fx)*fy + hash(ix+1, iy+1)*fx*fy
}
func gaussian(x, y, cx, cy, sx, sy float64) float64 {
	return math.Exp(-((x-cx)*(x-cx)/(sx*sx) + (y-cy)*(y-cy)/(sy*sy)))
}
func (m *Map) landAt(lon, lat float64) float64 {
	x := (lon - m.geo.Bounds[0]) / (m.geo.Bounds[1] - m.geo.Bounds[0]) * (maskW - 1)
	y := (m.geo.Bounds[3] - lat) / (m.geo.Bounds[3] - m.geo.Bounds[2]) * (maskH - 1)
	if x < 0 || y < 0 || x >= maskW-1 || y >= maskH-1 {
		return 0
	}
	ix, iy := int(x), int(y)
	fx, fy := x-float64(ix), y-float64(iy)
	at := func(dx, dy int) float64 { return float64(m.mask[(iy+dy)*maskW+ix+dx]) / 255 }
	return at(0, 0)*(1-fx)*(1-fy) + at(1, 0)*fx*(1-fy) + at(0, 1)*(1-fx)*fy + at(1, 1)*fx*fy
}
func (m *Map) Elevation(lon, lat float64) float64 {
	land := m.landAt(lon, lat)
	if land < .02 {
		return -.012
	}
	// Broad plateau and continuous ranges. Fine rock detail is shaded texture,
	// rather than independent peaks that visibly buckle the coast mesh.
	himalaya := 28.6 + math.Sin((lon-83)*.12)*1.3
	tianshan := 42.3 + math.Sin((lon-78)*.18)*1.1
	qinling := 33.6 + math.Sin((lon-102)*.3)*.6
	e := gaussian(lon, lat, 88, 33, 13, 6)*.48 +
		gaussian(lon, lat, 87, himalaya, 13, .85)*.42 +
		gaussian(lon, lat, 84, tianshan, 11, .85)*.34 +
		gaussian(lon, lat, 106, qinling, 7, .7)*.18 +
		gaussian(lon, lat, 103, 28.6, 3, 5)*.23 +
		gaussian(lon, lat, 127, 44, 2.6, 8)*.18 +
		gaussian(lon, lat, 117, 26, 3.5, 5)*.12 +
		gaussian(lon, lat, 108, 25, 9, 4)*.13
	ridges := (.5 + .5*math.Sin(lon*4+lat*1.7+noise(lon*.4, lat*.4)*6)) * .12
	return (.018 + e*(.9+ridges)) * land
}
func (m *Map) prepareSurface() {
	m.surface = make([]surfacePoint, maskW*maskH)
	for y := 0; y < maskH; y++ {
		for x := 0; x < maskW; x++ {
			lon := m.geo.Bounds[0] + float64(x)/(maskW-1)*(m.geo.Bounds[1]-m.geo.Bounds[0])
			lat := m.geo.Bounds[3] - float64(y)/(maskH-1)*(m.geo.Bounds[3]-m.geo.Bounds[2])
			e := m.Elevation(lon, lat)
			dx := (m.Elevation(lon+.08, lat) - m.Elevation(lon-.08, lat)) / .16
			dy := (m.Elevation(lon, lat+.08) - m.Elevation(lon, lat-.08)) / .16
			shade := 1.0
			_, _ = dx, dy // Approved HD map: no directional terrain shadow.
			kind := hdart.Plain
			base := [3]float64{151, 124, 88}
			if lon < 101 {
				kind = hdart.Desert
				base = [3]float64{181, 149, 103}
			}
			if e > .16 {
				kind = hdart.Mountain
				base = [3]float64{134, 116, 96}
			}
			if lon > 104 && lat < 32 && e < .30 {
				kind = hdart.Forest
				base = [3]float64{101, 85, 66}
			}
			if lon > 124 && lat > 40 {
				kind = hdart.Forest
				base = [3]float64{115, 96, 75}
			}
			if e > .52 {
				kind = hdart.Snow
				base = [3]float64{192, 183, 161}
			}
			p := hdart.Sample(kind, int(lon*37), int(lat*37))
			grain := (float64(p.R)+float64(p.G)+float64(p.B))/3 - 110
			macro := (noise(lon*.5, lat*.5) - .5) * 10
			for k := range base {
				base[k] = math.Max(0, math.Min(255, (base[k]+grain*.18+macro)*shade))
			}
			// Presentation-only regions use the existing province anchors. Their
			// approximate outline does not determine adjacency or battlefield terrain.
			wx, wy := lon+(noise(lon*.7, lat*.7)-.5)*.9, lat+(noise(lon*.6+14, lat*.6)-.5)*.7
			nearest, best := 0, math.Inf(1)
			for i, p := range Anchors {
				d := (wx-p[0])*(wx-p[0]) + (wy-p[1])*(wy-p[1])*1.3
				if d < best {
					nearest, best = i, d
				}
			}
			m.surface[y*maskW+x] = surfacePoint{rgb: [3]byte{byte(base[0]), byte(base[1]), byte(base[2])}, province: uint8(nearest)}
		}
	}
	for y := 1; y < maskH-1; y++ {
		for x := 1; x < maskW-1; x++ {
			i := y*maskW + x
			p := m.surface[i].province
			m.surface[i].border = m.surface[i+1].province != p || m.surface[i+maskW].province != p
		}
	}
}
func (m *Map) surfaceAt(lon, lat float64, political bool, factions *[39]int) [3]float64 {
	x := clamp((lon-m.geo.Bounds[0])/(m.geo.Bounds[1]-m.geo.Bounds[0])) * (maskW - 1)
	y := clamp((m.geo.Bounds[3]-lat)/(m.geo.Bounds[3]-m.geo.Bounds[2])) * (maskH - 1)
	ix, iy := min(maskW-2, int(x)), min(maskH-2, int(y))
	fx, fy := x-float64(ix), y-float64(iy)
	p := m.surface[iy*maskW+ix]
	rgb := [3]float64{}
	for k := range rgb {
		rgb[k] = float64(m.surface[iy*maskW+ix].rgb[k])*(1-fx)*(1-fy) + float64(m.surface[iy*maskW+ix+1].rgb[k])*fx*(1-fy) + float64(m.surface[(iy+1)*maskW+ix].rgb[k])*(1-fx)*fy + float64(m.surface[(iy+1)*maskW+ix+1].rgb[k])*fx*fy
	}
	if political {
		tints := [10][3]float64{{132, 103, 78}, {162, 136, 92}, {143, 115, 99}, {108, 119, 139}, {155, 121, 83}, {124, 102, 91}, {153, 91, 74}, {174, 139, 98}, {132, 112, 132}, {166, 119, 86}}
		tint := tints[max(0, factions[p.province])%10]
		// Preserve terrain luminance beneath faction pigment.
		light := (rgb[0] + rgb[1] + rgb[2]) / 390
		for k := range rgb {
			rgb[k] = rgb[k]*.46 + tint[k]*light*.54
		}
		if p.border {
			for k := range rgb {
				rgb[k] *= .68
			}
		}
	}
	return rgb
}
func (m *Map) Project(p Point, c Camera, w, h int) Point {
	return project(p, m.Elevation(p[0], p[1]), c, w, h)
}
func project(p Point, height float64, c Camera, w, h int) Point {
	e, x, z := height*.28, (p[0]-106)*.04, (p[1]-35)*.04
	y := e*math.Cos(c.Tilt) + z*math.Sin(c.Tilt)
	d := 3.5 - e*math.Sin(c.Tilt) + z*math.Cos(c.Tilt)
	return Point{(x*2.5*c.Zoom/(float64(w)/float64(h))/d + c.PanX + 1) * float64(w) / 2, (1 - (y*2.5*c.Zoom/d + c.PanY)) * float64(h) / 2}
}
func (m *Map) Hit(x, y float64, c Camera, w, h int) int {
	best, id := 22.0, 0
	for i, p := range Anchors {
		q := m.Project(p, c, w, h)
		d := math.Hypot(x-q[0], y-q[1])
		if d < best {
			best, id = d, i+1
		}
	}
	return id
}
func (m *Map) Render(c Camera, w, h int, retro, political bool, factions [39]int) *image.RGBA {
	w = max(1, min(w, 1280))
	h = max(1, min(h, 720))
	key := renderKey{c, w, h, retro, political, factions}
	if m.cached != nil && key == m.key {
		return m.cached
	}
	if m.cached == nil || m.cached.Bounds().Dx() != w || m.cached.Bounds().Dy() != h {
		m.cached = image.NewRGBA(image.Rect(0, 0, w, h))
	}
	dst := m.cached
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := hdart.Sample(hdart.Ocean, x/2, y/2)
			grain := (float64(p.R) + float64(p.G) + float64(p.B)) / 3
			o := y*dst.Stride + x*4
			if retro {
				dst.Pix[o], dst.Pix[o+1], dst.Pix[o+2] = 33, 78, 134
			} else {
				depth := float64(y) / float64(h) * 4
				dst.Pix[o], dst.Pix[o+1], dst.Pix[o+2] = byte(18+grain*.08), byte(38+grain*.13+depth), byte(59+grain*.18+depth)
			}
			dst.Pix[o+3] = 255
		}
	}
	for i, v := range m.vertices {
		m.projected[i] = project(Point{v.lon, v.lat}, v.height, c, w, h)
	}
	for i := len(m.indices) - 3; i >= 0; i -= 3 {
		ia, ib, ic := m.indices[i], m.indices[i+1], m.indices[i+2]
		va, vb, vc := m.vertices[ia], m.vertices[ib], m.vertices[ic]
		if va.land < .02 && vb.land < .02 && vc.land < .02 {
			continue
		}
		a, b, c := m.projected[ia], m.projected[ib], m.projected[ic]
		det := (b[1]-c[1])*(a[0]-c[0]) + (c[0]-b[0])*(a[1]-c[1])
		if math.Abs(det) < .001 {
			continue
		}
		minX, maxX := max(0, int(math.Floor(math.Min(a[0], math.Min(b[0], c[0]))))), min(w-1, int(math.Ceil(math.Max(a[0], math.Max(b[0], c[0])))))
		minY, maxY := max(0, int(math.Floor(math.Min(a[1], math.Min(b[1], c[1]))))), min(h-1, int(math.Ceil(math.Max(a[1], math.Max(b[1], c[1])))))
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				u := ((b[1]-c[1])*(float64(x)+.5-c[0]) + (c[0]-b[0])*(float64(y)+.5-c[1])) / det
				v := ((c[1]-a[1])*(float64(x)+.5-c[0]) + (a[0]-c[0])*(float64(y)+.5-c[1])) / det
				t := 1 - u - v
				if u < -.00001 || v < -.00001 || t < -.00001 {
					continue
				}
				lon, lat := u*va.lon+v*vb.lon+t*vc.lon, u*va.lat+v*vb.lat+t*vc.lat
				land := m.landAt(lon, lat)
				if land < .04 {
					continue
				}
				rgb := m.surfaceAt(lon, lat, political, &factions)
				offset := y*dst.Stride + x*4
				alpha := clamp((land - .04) * 2.1)
				for k := 0; k < 3; k++ {
					value := rgb[k]
					if retro {
						value = math.Round(value/42.5) * 42.5
					}
					dst.Pix[offset+k] = byte(math.Max(0, math.Min(255, value*alpha+float64(dst.Pix[offset+k])*(1-alpha))))
				}
			}
		}
	}
	for _, river := range m.geo.Rivers {
		for i := 1; i < len(river); i++ {
			m.line(dst, m.Project(river[i-1], key.Camera, w, h), m.Project(river[i], key.Camera, w, h), [3]byte{112, 144, 164})
		}
	}
	m.key = key
	m.Renders++
	return dst
}
func (m *Map) line(dst *image.RGBA, a, b Point, rgb [3]byte) {
	dx, dy := b[0]-a[0], b[1]-a[1]
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
	if steps > 2000 {
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(max(1, steps))
		x, y := int(a[0]+dx*t), int(a[1]+dy*t)
		if image.Pt(x, y).In(dst.Bounds()) {
			offset := y*dst.Stride + x*4
			for k := 0; k < 3; k++ {
				dst.Pix[offset+k] = byte((int(dst.Pix[offset+k]) + int(rgb[k])) / 2)
			}
		}
	}
}
