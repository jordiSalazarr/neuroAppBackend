package visualspatialsubtest

import (
	"errors"
	"image"
	"image/color"
	"math"
	"sort"

	"gocv.io/x/gocv"
	createvisualspatialsubtest "neuro.app.jordi/internal/evaluation/application/commands/create-visual-spatial-subtest"
)

// Implementa app.Analyzer
type GoCVClockAnalyzer struct{}

func NewGoCVClockAnalyzer() *GoCVClockAnalyzer { return &GoCVClockAnalyzer{} }

// ---------- tipo nombrado para líneas candidatas ----------
type cand struct {
	p1, p2 image.Point
	len    float64
	ang    float64
}

func (a *GoCVClockAnalyzer) Analyze(imageBytes []byte, expectedHour, expectedMin int) (createvisualspatialsubtest.Analysis, []byte, error) {
	var out createvisualspatialsubtest.Analysis
	if len(imageBytes) == 0 {
		return out, nil, errors.New("imageBytes vacío")
	}

	buf := gocv.NewMat()
	defer buf.Close()
	buf, err := gocv.IMDecode(imageBytes, gocv.IMReadColor)
	if err != nil || buf.Empty() {
		return out, nil, errors.New("no se pudo decodificar la imagen")
	}

	src := buf.Clone()
	defer src.Close()
	debug := src.Clone()
	defer debug.Close()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(src, &gray, gocv.ColorBGRToGray)

	blur := gocv.NewMat()
	defer blur.Close()
	gocv.GaussianBlur(gray, &blur, image.Pt(5, 5), 0, 0, gocv.BorderDefault)

	// --- DIAL (contorno principal) ---
	th := gocv.NewMat()
	defer th.Close()
	gocv.AdaptiveThreshold(blur, &th, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinaryInv, 31, 5)

	contours := gocv.FindContours(th, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	if contours.Size() == 0 {
		out.Reasons = append(out.Reasons, "No se detectaron contornos para el dial.")
		return finalize(out, expectedHour, expectedMin), encodePNG(debug), nil
	}

	type dialCand struct {
		center      image.Point
		radius      float64
		area        float64
		circularity float64
	}
	var best dialCand
	imgArea := float64(src.Rows() * src.Cols())
	maxArea := 0.0
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := math.Abs(gocv.ContourArea(c))
		if area < imgArea*0.02 {
			continue
		}
		center, r := centerAndRadius(c)
		if r <= 0 {
			continue
		}
		circ := area / (math.Pi * r * r)
		if area > maxArea {
			maxArea = area
			best = dialCand{center: center, radius: r, area: area, circularity: circ}
		}
	}
	if maxArea == 0 {
		out.Reasons = append(out.Reasons, "No se encontró un dial candidato con tamaño suficiente.")
		return finalize(out, expectedHour, expectedMin), encodePNG(debug), nil
	}

	out.CenterX = best.center.X
	out.CenterY = best.center.Y
	out.Radius = best.radius
	out.DialCircularity = best.circularity

	// Dibujo del dial
	gocv.Circle(&debug, best.center, int(best.radius), color.RGBA{0, 255, 0, 0}, 2)
	gocv.Circle(&debug, best.center, 4, color.RGBA{0, 0, 255, 0}, -1)

	// --- EDGES + FALLBACKS ---
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(blur, &edges, 25, 75)

	// Otsu + Dilate para reforzar trazos finos
	bin := gocv.NewMat()
	defer bin.Close()
	gocv.Threshold(gray, &bin, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	binDil := gocv.NewMat()
	defer binDil.Close()
	gocv.Dilate(bin, &binDil, kernel)

	// Closing morfológico sobre edges para unir pequeños gaps
	edgesClose := gocv.NewMat()
	defer edgesClose.Close()
	gocv.MorphologyEx(edges, &edgesClose, gocv.MorphClose, kernel)

	// --- MÁSCARA ANULAR: [0.07·R .. 0.85·R] ---
	// Evita el perímetro (números) y deja un agujero central pequeño (no corta agujas)
	ring := gocv.Zeros(edges.Rows(), edges.Cols(), gocv.MatTypeCV8U)
	defer ring.Close()
	gocv.Circle(&ring, best.center, int(best.radius*0.85), color.RGBA{255, 255, 255, 0}, -1) // exterior blanco
	gocv.Circle(&ring, best.center, int(best.radius*0.07), color.RGBA{0, 0, 0, 0}, -1)       // agujero pequeño

	// Aplica anillo
	edgesIn := gocv.NewMat()
	defer edgesIn.Close()
	gocv.BitwiseAnd(edgesClose, ring, &edgesIn)
	binIn := gocv.NewMat()
	defer binIn.Close()
	gocv.BitwiseAnd(binDil, ring, &binIn)

	// --- Hough + filtros geométricos ---
	var rawCands []cand
	trySets := []struct {
		src         *gocv.Mat
		houghThr    int
		minLenMul   float64
		maxGapMul   float64
		centerSlack float64
	}{
		{&edgesIn, 28, 0.50, 0.10, 0.18}, // estricto
		{&edgesIn, 22, 0.45, 0.12, 0.22}, // medio
		{&binIn, 24, 0.45, 0.12, 0.22},   // bin dilatada
		{&binIn, 18, 0.35, 0.15, 0.26},   // permisivo (horaria corta)
		{&edgesIn, 12, 0.25, 0.20, 0.35}, // muy permisivo
	}

	for _, ts := range trySets {
		lines := gocv.NewMat()
		gocv.HoughLinesPWithParams(
			*ts.src, &lines,
			1.0, math.Pi/180.0, ts.houghThr,
			float32(best.radius*ts.minLenMul), // minLineLength
			float32(best.radius*ts.maxGapMul), // maxLineGap
		)

		// Dibuja todas las líneas crudas (amarillo) para depurar
		for i := 0; i < lines.Rows(); i++ {
			l := lines.GetVeciAt(i, 0)
			p1 := image.Pt(int(l[0]), int(l[1]))
			p2 := image.Pt(int(l[2]), int(l[3]))
			gocv.Line(&debug, p1, p2, color.RGBA{200, 200, 0, 0}, 1)
		}

		// Filtrado geométrico robusto: pasa por el centro, nace en el centro y sale hacia fuera
		for i := 0; i < lines.Rows(); i++ {
			l := lines.GetVeciAt(i, 0)
			p1 := image.Pt(int(l[0]), int(l[1]))
			p2 := image.Pt(int(l[2]), int(l[3]))

			// Distancia del centro al segmento (debe ser pequeña)
			if distPointToSegment(best.center, p1, p2) > best.radius*ts.centerSlack {
				continue
			}
			// r1: extremo más cercano al centro; r2: extremo más lejano
			r1 := math.Min(distPts(best.center, p1), distPts(best.center, p2))
			r2 := math.Max(distPts(best.center, p1), distPts(best.center, p2))
			// Debe nacer del centro (r1 <= 0.18R) y salir hacia el dial (r2 >= 0.40R)
			if r1 > best.radius*0.18 {
				continue
			}
			if r2 < best.radius*0.40 {
				continue
			}

			length := distPts(p1, p2)
			if length < best.radius*ts.minLenMul {
				continue
			}

			v := farVectorFromCenter(best.center, p1, p2)
			ang := vectorAngleClock(v)
			rawCands = append(rawCands, cand{p1: p1, p2: p2, len: length, ang: ang})
		}
		lines.Close()

		if len(rawCands) >= 2 {
			break
		}
	}

	// Fallback por componentes (puede añadir 1 o más candidatos)
	if len(rawCands) < 2 {
		compCands := componentFallback(binIn, best.center, best.radius)
		if len(compCands) > 0 {
			rawCands = append(rawCands, compCands...)
		}
	}

	// Fallback radial + extensión artificial (fabrica 2 rayos si aún no hay 2)
	if len(rawCands) < 2 {
		a1, a2, ok := radialTwoPeaksFocused(binIn, best.center, best.radius)
		if ok {
			minLen := best.radius * 0.95
			hrLen := best.radius * 0.60
			rawCands = append(rawCands,
				cand{p1: best.center, p2: pointOnCircle(best.center, a1, minLen), len: minLen, ang: a1},
				cand{p1: best.center, p2: pointOnCircle(best.center, a2, hrLen), len: hrLen, ang: a2},
			)
		}
	}

	if len(rawCands) < 2 {
		out.Reasons = append(out.Reasons, "No se hallaron dos direcciones de aguja (Hough/Componentes/Radial).")
		out = finalize(out, expectedHour, expectedMin)
		return out, encodePNG(debug), nil
	}

	// --- Deduplicación por ángulo (más laxa) + guard ---
	cands := dedupByAngle(rawCands, 6.0)
	if len(cands) < 2 {
		// Último intento: forzar con radial una segunda dirección y deduplicar de nuevo
		a1, a2, ok := radialTwoPeaksFocused(binIn, best.center, best.radius)
		if ok {
			minLen := best.radius * 0.95
			hrLen := best.radius * 0.60
			cands = append(cands,
				cand{p1: best.center, p2: pointOnCircle(best.center, a1, minLen), len: minLen, ang: a1},
				cand{p1: best.center, p2: pointOnCircle(best.center, a2, hrLen), len: hrLen, ang: a2},
			)
			cands = dedupByAngle(cands, 6.0)
		}
	}
	if len(cands) < 2 {
		out.Reasons = append(out.Reasons, "Solo se detectó una aguja tras deduplicar por ángulo (con radial).")
		out = finalize(out, expectedHour, expectedMin)
		return out, encodePNG(debug), nil
	}
	if len(cands) > 6 {
		cands = cands[:6]
	}

	// Elegir mejor par por mínimo error total
	_, _, minAng, hourAng := selectBestPair(cands, expectedHour, expectedMin)

	// --- Dibujo final extendido (homogeneiza visualmente) ---
	pMinExt := pointOnCircle(best.center, minAng, best.radius*0.95)
	pHrExt := pointOnCircle(best.center, hourAng, best.radius*0.60)
	gocv.Line(&debug, best.center, pMinExt, color.RGBA{0, 255, 255, 0}, 3) // cian
	gocv.Line(&debug, best.center, pHrExt, color.RGBA{255, 0, 255, 0}, 3)  // magenta

	out.MinuteAngleDeg = normalizeDeg(minAng)
	out.HourAngleDeg = normalizeDeg(hourAng)
	out = finalize(out, expectedHour, expectedMin)

	// --- Reglas de aprobación (ajústalas si tu dataset es muy manuscrito) ---
	minuteTol := 15.0
	hourTol := 20.0
	pass := true
	if out.DialCircularity < 0.70 {
		pass = false
		out.Reasons = append(out.Reasons, "Dial poco circular (< 0.70).")
	}
	if out.MinuteAngularErrorDeg > minuteTol {
		pass = false
		out.Reasons = append(out.Reasons, "Minutero fuera de tolerancia (±15°).")
	}
	if out.HourAngularErrorDeg > hourTol {
		pass = false
		out.Reasons = append(out.Reasons, "Horario fuera de tolerancia (±20°).")
	}
	out.Pass = pass
	if pass {
		out.Reasons = append(out.Reasons, "Dial razonablemente circular y agujas en posiciones esperadas.")
	}
	return out, encodePNG(debug), nil
}

// —— utilidades geométricas y de ángulos —— //

func centerAndRadius(contour gocv.PointVector) (image.Point, float64) {
	n := contour.Size()
	if n == 0 {
		return image.Pt(0, 0), 0
	}
	var sx, sy float64
	for i := 0; i < n; i++ {
		p := contour.At(i)
		sx += float64(p.X)
		sy += float64(p.Y)
	}
	cx := sx / float64(n)
	cy := sy / float64(n)

	var sumr float64
	for i := 0; i < n; i++ {
		p := contour.At(i)
		sumr += math.Hypot(float64(p.X)-cx, float64(p.Y)-cy)
	}
	r := sumr / float64(n)
	return image.Pt(int(cx), int(cy)), r
}

func distPts(a, b image.Point) float64 {
	return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y))
}

func distPointToSegment(p, a, b image.Point) float64 {
	px, py := float64(p.X), float64(p.Y)
	ax, ay := float64(a.X), float64(a.Y)
	bx, by := float64(b.X), float64(b.Y)
	vx, vy := bx-ax, by-ay
	wx, wy := px-ax, py-ay

	c1 := vx*wx + vy*wy
	if c1 <= 0 {
		return math.Hypot(px-ax, py-ay)
	}
	c2 := vx*vx + vy*vy
	if c2 <= c1 {
		return math.Hypot(px-bx, py-by)
	}
	t := c1 / c2
	projx := ax + t*vx
	projy := ay + t*vy
	return math.Hypot(px-projx, py-projy)
}

func farVectorFromCenter(c, a, b image.Point) image.Point {
	if distPts(c, a) >= distPts(c, b) {
		return image.Pt(a.X-c.X, a.Y-c.Y)
	}
	return image.Pt(b.X-c.X, b.Y-c.Y)
}

// 0° arriba, sentido horario creciente
func vectorAngleClock(v image.Point) float64 {
	dx := float64(v.X)
	dy := float64(v.Y)
	ang := math.Atan2(-dy, dx) * 180 / math.Pi // 0° a la derecha
	ang = normalizeDeg(ang - 90.0)             // 0° arriba
	return normalizeDeg(-ang)                  // horario
}

func normalizeDeg(d float64) float64 {
	for d < 0 {
		d += 360
	}
	for d >= 360 {
		d -= 360
	}
	return d
}

func angError(a, b float64) float64 {
	diff := math.Abs(normalizeDeg(a) - normalizeDeg(b))
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

func finalize(out createvisualspatialsubtest.Analysis, expectedHour, expectedMin int) createvisualspatialsubtest.Analysis {
	out.ExpectedMinuteAngle = float64((expectedMin % 60) * 6)
	h := expectedHour % 12
	out.ExpectedHourAngle = float64(h*30) + float64(expectedMin)*0.5
	out.MinuteAngularErrorDeg = angError(out.MinuteAngleDeg, out.ExpectedMinuteAngle)
	out.HourAngularErrorDeg = angError(out.HourAngleDeg, out.ExpectedHourAngle)
	return out
}

func encodePNG(m gocv.Mat) []byte {
	if m.Empty() {
		return nil
	}
	b, _ := gocv.IMEncode(".png", m)
	defer b.Close()
	return b.GetBytes()
}

// Un extremo del segmento debe nacer "cerca" del centro
func nearCenter(p image.Point, c image.Point, r float64, k float64) bool {
	return distPts(p, c) <= r*k
}

// Deduplica por ángulo para evitar coger dos segmentos casi colineales
func dedupByAngle(in []cand, minSepDeg float64) []cand {
	out := make([]cand, 0, len(in))
	for _, c := range in {
		keep := true
		for _, k := range out {
			if angError(c.ang, k.ang) < minSepDeg {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, c)
		}
	}
	return out
}

// Selecciona el mejor par (minutero/horario) por mínimo error total con la hora esperada
func selectBestPair(cands []cand, expectedHour, expectedMin int) (minCand, hrCand cand, minuteAngle, hourAngle float64) {
	// Si no hay 2 candidatos, devolvemos ceros sin pánico.
	n := len(cands)
	if n < 2 {
		return
	}

	mExp := float64((expectedMin % 60) * 6)
	hExp := float64((expectedHour%12)*30) + float64(expectedMin)*0.5

	bestErr := math.MaxFloat64
	bestI, bestJ := -1, -1
	swap := false

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			a1 := normalizeDeg(cands[i].ang)
			a2 := normalizeDeg(cands[j].ang)

			errA := angError(a1, mExp) + angError(a2, hExp) // i=minute, j=hour
			errB := angError(a2, mExp) + angError(a1, hExp) // j=minute, i=hour

			if errA < bestErr {
				bestErr = errA
				bestI, bestJ = i, j
				swap = false
			}
			if errB < bestErr {
				bestErr = errB
				bestI, bestJ = i, j
				swap = true
			}
		}
	}

	// Si por lo que sea no se eligió pareja válida, usa las 2 más largas
	if bestI == -1 || bestJ == -1 || bestI >= n || bestJ >= n || bestI == bestJ {
		sort.Slice(cands, func(i, j int) bool { return cands[i].len > cands[j].len })
		// n >= 2 garantizado por el guard del inicio
		minCand, hrCand = cands[0], cands[1]
		a1 := normalizeDeg(minCand.ang)
		a2 := normalizeDeg(hrCand.ang)
		// asigna por mínimo error total
		if angError(a1, mExp)+angError(a2, hExp) <= angError(a2, mExp)+angError(a1, hExp) {
			minuteAngle, hourAngle = a1, a2
		} else {
			minuteAngle, hourAngle = a2, a1
			minCand, hrCand = hrCand, minCand
		}
		return
	}

	// Camino normal
	if !swap {
		minCand, hrCand = cands[bestI], cands[bestJ]
		minuteAngle, hourAngle = normalizeDeg(minCand.ang), normalizeDeg(hrCand.ang)
	} else {
		minCand, hrCand = cands[bestJ], cands[bestI]
		minuteAngle, hourAngle = normalizeDeg(minCand.ang), normalizeDeg(hrCand.ang)
	}
	return
}

// Busca componentes conectados "negros" dentro del disco 0.93R,
// que toquen el centro y sean lo bastante grandes; para cada uno
// calcula el ángulo hacia su punto más alejado del centro.
func componentFallback(bin gocv.Mat, center image.Point, radius float64) []cand {
	out := []cand{}
	if bin.Empty() {
		return out
	}
	// Contornos de blobs (bin es 255 en objetos, 0 en fondo)
	contours := gocv.FindContours(bin, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := math.Abs(gocv.ContourArea(c))
		// descarta ruido muy pequeño (0.2% del área del dial)
		if area < (math.Pi*radius*radius)*0.002 {
			continue
		}
		// Debe tocar el centro: algún punto del contorno dentro de 0.25R
		touches := false
		minDist := 1e9
		var farP image.Point
		maxDist := -1.0
		for j := 0; j < c.Size(); j++ {
			p := c.At(j)
			d := distPts(center, image.Pt(p.X, p.Y))
			if d < minDist {
				minDist = d
			}
			if d > maxDist {
				maxDist = d
				farP = image.Pt(p.X, p.Y)
			}
			if d <= radius*0.25 {
				touches = true
			}
		}
		if !touches {
			continue
		}
		if maxDist < radius*0.25 {
			continue // demasiado corto
		}
		v := image.Pt(farP.X-center.X, farP.Y-center.Y)
		ang := vectorAngleClock(v)
		out = append(out, cand{
			p1:  center,
			p2:  farP,
			len: maxDist,
			ang: ang,
		})
	}
	return out
}

// Escaneo radial “centrado”: busca 2 picos solo entre 0.12·R y 0.60·R, con supresión ±15°
func radialTwoPeaksFocused(img gocv.Mat, center image.Point, radius float64) (float64, float64, bool) {
	if img.Empty() {
		return 0, 0, false
	}
	scores := make([]float64, 360)
	rMin := radius * 0.12
	rMax := radius * 0.60
	if rMin < 1 {
		rMin = 1
	}

	for a := 0; a < 360; a++ {
		scores[a] = rayScoreWeighted(img, center, float64(a), rMin, rMax, 1.0)
	}
	smoothScores(scores, 2)

	i1 := argMax(scores)
	if i1 < 0 {
		return 0, 0, false
	}
	supressAround(scores, i1, 15) // ±15° para separar de verdad del minutero
	i2 := argMax(scores)
	if i2 < 0 {
		return 0, 0, false
	}
	return float64(i1), float64(i2), true
}

// Igual que rayScore, pero pondera más lo cercano al centro (mejor para horaria corta)
func rayScoreWeighted(img gocv.Mat, center image.Point, angDeg, rMin, rMax, step float64) float64 {
	sum := 0.0
	rad := (90.0 - angDeg) * (math.Pi / 180.0)
	for r := rMin; r <= rMax; r += step {
		x := float64(center.X) + r*math.Cos(rad)
		y := float64(center.Y) - r*math.Sin(rad)
		xi, yi := int(math.Round(x)), int(math.Round(y))
		if yi >= 0 && yi < img.Rows() && xi >= 0 && xi < img.Cols() {
			v := img.GetUCharAt(yi, xi)
			w := 1.0 / (1.0 + 0.02*r) // peso ↓ con r (favorece la horaria)
			sum += float64(v) * w
		}
	}
	return sum
}

// Punto a “distancia” len desde el centro en un ángulo (0° arriba, horario)
func pointOnCircle(center image.Point, angDeg, length float64) image.Point {
	rad := (90.0 - angDeg) * (math.Pi / 180.0)
	x := float64(center.X) + length*math.Cos(rad)
	y := float64(center.Y) - length*math.Sin(rad)
	return image.Pt(int(math.Round(x)), int(math.Round(y)))
}

// --- Helpers para el escaneo radial ---

// Suaviza el array "a" con media móvil circular de ventana ±wnd.
func smoothScores(a []float64, wnd int) {
	if wnd <= 0 || len(a) == 0 {
		return
	}
	n := len(a)
	out := make([]float64, n)
	den := float64(2*wnd + 1)
	for i := 0; i < n; i++ {
		s := 0.0
		for k := -wnd; k <= wnd; k++ {
			j := (i + k + n) % n // circular
			s += a[j]
		}
		out[i] = s / den
	}
	copy(a, out)
}

// Devuelve el índice del máximo; -1 si el vector está vacío o todo es 0.
func argMax(a []float64) int {
	if len(a) == 0 {
		return -1
	}
	maxIdx := 0
	maxVal := a[0]
	for i := 1; i < len(a); i++ {
		if a[i] > maxVal {
			maxVal = a[i]
			maxIdx = i
		}
	}
	if maxVal <= 0 {
		return -1
	}
	return maxIdx
}

// Anula (pone a 0) la energía alrededor de idx en una ventana ±width (circular).
func supressAround(a []float64, idx int, width int) {
	if len(a) == 0 {
		return
	}
	n := len(a)
	for k := -width; k <= width; k++ {
		j := (idx + k + n) % n
		a[j] = 0
	}
}
