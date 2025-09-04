package bvmtcv

import (
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"math"

	"gocv.io/x/gocv"
)

type ShapeType int

const (
	ShapeCircle ShapeType = iota
	ShapeSquare
	ShapeTriangle
)

type ShapeScore struct {
	Score          float64
	Pass           bool
	Reasons        []string
	IoU            float64
	Circularity    float64
	AngleRMSE      float64
	SideCV         float64
	Center         image.Point
	BBox           image.Rectangle
	DebugPNGBase64 string
}

type Params struct {
	PassThreshold float64 // e.g., 70
}

func ScoreGeometricPNG(pngBytes []byte, expected ShapeType, ps Params) (*ShapeScore, error) {
	img, err := gocv.IMDecode(pngBytes, gocv.IMReadColor)
	if err != nil || img.Empty() {
		return nil, errors.New("decode error")
	}
	defer img.Close()
	return scoreGeometricMat(img, expected, ps)
}

func scoreGeometricMat(src gocv.Mat, expected ShapeType, ps Params) (*ShapeScore, error) {
	// --- Preprocess ---
	gray := gocv.NewMat()
	blur := gocv.NewMat()
	th := gocv.NewMat()
	defer gray.Close()
	defer blur.Close()
	defer th.Close()

	gocv.CvtColor(src, &gray, gocv.ColorBGRToGray)
	gocv.GaussianBlur(gray, &blur, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
	// Otsu; si tus trazos son oscuros, BinaryInv suele ir bien
	gocv.Threshold(blur, &th, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

	// Cierre para unir gaps
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.MorphologyEx(th, &th, gocv.MorphClose, kernel)

	// --- Contour ---
	contours := gocv.FindContours(th, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	if contours.Size() == 0 {
		return &ShapeScore{Pass: false, Reasons: []string{"No se detectó figura"}}, nil
	}
	bestIdx, bestArea := -1, 0.0
	for i := 0; i < contours.Size(); i++ {
		a := gocv.ContourArea(contours.At(i))
		if a > bestArea {
			bestArea = a
			bestIdx = i
		}
	}
	c := contours.At(bestIdx)
	bbox := gocv.BoundingRect(c)

	// Máscara usuario (contorno relleno)
	userMask := gocv.NewMatWithSize(th.Rows(), th.Cols(), gocv.MatTypeCV8U)
	defer userMask.Close()
	userMask.SetTo(gocv.NewScalar(0, 0, 0, 0))
	gocv.DrawContours(&userMask, contours, bestIdx, color.RGBA{255, 255, 255, 0}, -1)

	// --- Métricas por figura ---
	var iou, circ, angleRMSE, sideCV float64
	var center image.Point

	switch expected {
	case ShapeCircle:
		cx, cy, r := gocv.MinEnclosingCircle(c)
		center = image.Pt(int(cx), int(cy))
		per := gocv.ArcLength(c, true)
		area := gocv.ContourArea(c)
		if per > 0 {
			circ = 4 * math.Pi * area / (per * per)
			if circ > 1 {
				circ = 1
			} // numeric guard
		}
		ideal := drawIdealCircle(th.Rows(), th.Cols(), center, int(r))
		defer ideal.Close()
		iou = maskIoU(userMask, ideal)

	case ShapeSquare:
		approx := gocv.ApproxPolyDP(c, 0.02*gocv.ArcLength(c, true), true)
		pts := toPoints(approx)
		if len(pts) == 4 && isContourConvex(pts) {
			pts := toPoints(approx)
			angleRMSE = anglesRMSE(pts, 90.0)
			sideCV = sideLenCV(pts)
			ideal := drawIdealPoly(th.Rows(), th.Cols(), pts)
			defer ideal.Close()
			iou = maskIoU(userMask, ideal)
			center = image.Pt(bbox.Min.X+bbox.Dx()/2, bbox.Min.Y+bbox.Dy()/2)
		} else {
			return &ShapeScore{Pass: false, Reasons: []string{"No se detectó un cuadrado claro"}}, nil
		}

	case ShapeTriangle:
		approx := gocv.ApproxPolyDP(c, 0.02*gocv.ArcLength(c, true), true)
		pts := toPoints(approx)
		if approx.Size() == 3 && isContourConvex(pts) {
			pts := toPoints(approx)
			angleRMSE = anglesRMSE(pts, 60.0) // objetivo equilátero (ajustable)
			sideCV = sideLenCV(pts)
			ideal := drawIdealPoly(th.Rows(), th.Cols(), pts)
			defer ideal.Close()
			iou = maskIoU(userMask, ideal)
			center = image.Pt(bbox.Min.X+bbox.Dx()/2, bbox.Min.Y+bbox.Dy()/2)
		} else {
			return &ShapeScore{Pass: false, Reasons: []string{"No se detectó un triángulo claro"}}, nil
		}
	}

	// --- Score ---
	score, reasons := combineScore(expected, iou, circ, angleRMSE, sideCV)
	pass := score >= ps.PassThreshold

	// --- Debug overlay ---
	debug := overlayDebug(src, expected, c, center, bbox)

	// Encode PNG
	buf, _ := gocv.IMEncode(gocv.PNGFileExt, debug)
	debug.Close()
	b64 := base64.StdEncoding.EncodeToString(buf.GetBytes())
	buf.Close()

	return &ShapeScore{
		Score:          score,
		Pass:           pass,
		Reasons:        reasons,
		IoU:            iou,
		Circularity:    circ,
		AngleRMSE:      angleRMSE,
		SideCV:         sideCV,
		Center:         center,
		BBox:           bbox,
		DebugPNGBase64: b64,
	}, nil
}

// ===== Helpers =====

func drawIdealCircle(h, w int, center image.Point, r int) gocv.Mat {
	m := gocv.NewMatWithSize(h, w, gocv.MatTypeCV8U)
	m.SetTo(gocv.NewScalar(0, 0, 0, 0))
	gocv.Circle(&m, center, r, color.RGBA{255, 255, 255, 0}, -1)
	return m
}

func drawIdealPoly(h, w int, pts []image.Point) gocv.Mat {
	m := gocv.NewMatWithSize(h, w, gocv.MatTypeCV8U)
	m.SetTo(gocv.NewScalar(0, 0, 0, 0))

	// 1) Construimos un PointVector y añadimos los puntos uno a uno.
	pv := gocv.NewPointVector()
	for _, p := range pts {
		pv.Append(p)
	}
	defer pv.Close()

	// 2) Construimos el PointsVector y le añadimos el PointVector anterior.
	pvs := gocv.NewPointsVector()
	pvs.Append(pv)
	defer pvs.Close()

	// 3) Rellenamos el polígono en blanco.
	gocv.FillPoly(&m, pvs, color.RGBA{255, 255, 255, 0})
	return m
}

func maskIoU(a, b gocv.Mat) float64 {
	inter := gocv.NewMat()
	union := gocv.NewMat()
	defer inter.Close()
	defer union.Close()
	gocv.BitwiseAnd(a, b, &inter)
	gocv.BitwiseOr(a, b, &union)
	ai := float64(gocv.CountNonZero(inter))
	au := float64(gocv.CountNonZero(union))
	if au == 0 {
		return 0
	}
	return ai / au
}

func toPoints(pv gocv.PointVector) []image.Point {
	pts := make([]image.Point, pv.Size())
	for i := 0; i < pv.Size(); i++ {
		p := pv.At(i)
		pts[i] = image.Pt(p.X, p.Y)
	}
	return pts
}

func anglesRMSE(pts []image.Point, target float64) float64 {
	n := len(pts)
	if n < 3 {
		return 180
	}
	// cierra polígono
	sum := 0.0
	for i := 0; i < n; i++ {
		a := pts[(i-1+n)%n]
		b := pts[i]
		c := pts[(i+1)%n]
		ang := angleDeg(a, b, c)
		d := ang - target
		sum += d * d
	}
	return math.Sqrt(sum / float64(n))
}

func angleDeg(a, b, c image.Point) float64 {
	v1x, v1y := float64(a.X-b.X), float64(a.Y-b.Y)
	v2x, v2y := float64(c.X-b.X), float64(c.Y-b.Y)
	dot := v1x*v2x + v1y*v2y
	n1 := math.Hypot(v1x, v1y)
	n2 := math.Hypot(v2x, v2y)
	if n1 == 0 || n2 == 0 {
		return 180
	}
	cos := dot / (n1 * n2)
	if cos > 1 {
		cos = 1
	}
	if cos < -1 {
		cos = -1
	}
	return math.Acos(cos) * 180 / math.Pi
}

func sideLenCV(pts []image.Point) float64 {
	n := len(pts)
	if n == 0 {
		return 1
	}
	lengths := make([]float64, n)
	for i := 0; i < n; i++ {
		a, b := pts[i], pts[(i+1)%n]
		lengths[i] = math.Hypot(float64(b.X-a.X), float64(b.Y-a.Y))
	}
	mean := 0.0
	for _, v := range lengths {
		mean += v
	}
	mean /= float64(n)
	if mean == 0 {
		return 1
	}
	var sq float64
	for _, v := range lengths {
		d := v - mean
		sq += d * d
	}
	variance := sq / float64(n)
	return math.Sqrt(variance) / mean // coeficiente de variación
}

func combineScore(shape ShapeType, iou, circ, angleRMSE, sideCV float64) (float64, []string) {
	reasons := []string{}
	clip := func(x, lo, hi float64) float64 {
		if x < lo {
			return lo
		}
		if x > hi {
			return hi
		}
		return x
	}
	iou = clip(iou, 0, 1)
	circ = clip(circ, 0, 1)

	switch shape {
	case ShapeCircle:
		radScore := 1.0 // si implementas MAD radial, úsalo aquí
		score := 100 * (0.6*iou + 0.3*circ + 0.1*radScore)
		if iou < 0.7 {
			reasons = append(reasons, "IoU bajo para el círculo")
		}
		if circ < 0.8 {
			reasons = append(reasons, "Circularidad baja")
		}
		return score, reasons

	case ShapeSquare:
		angScore := clip(1-angleRMSE/20.0, 0, 1) // 20° como “máx. tolerable”
		sideScore := clip(1-sideCV/0.25, 0, 1)   // CV 0.25 ~ tolerancia
		score := 100 * (0.5*iou + 0.3*angScore + 0.2*sideScore)
		if iou < 0.65 {
			reasons = append(reasons, "IoU bajo para el cuadrado")
		}
		if angScore < 0.6 {
			reasons = append(reasons, "Ángulos alejados de 90°")
		}
		if sideScore < 0.6 {
			reasons = append(reasons, "Lados muy desiguales")
		}
		return score, reasons

	case ShapeTriangle:
		angScore := clip(1-angleRMSE/25.0, 0, 1) // triángulo algo más laxo
		sideScore := clip(1-sideCV/0.3, 0, 1)
		score := 100 * (0.5*iou + 0.3*angScore + 0.2*sideScore)
		if iou < 0.6 {
			reasons = append(reasons, "IoU bajo para el triángulo")
		}
		if angScore < 0.6 {
			reasons = append(reasons, "Ángulos alejados de 60°")
		}
		if sideScore < 0.6 {
			reasons = append(reasons, "Lados muy desiguales")
		}
		return score, reasons
	}
	return 0, []string{"Figura no soportada"}
}

// reemplaza tu overlayDebug por esto (compatible v0.33–v0.41 aprox.)
func overlayDebug(src gocv.Mat, shape ShapeType, contour gocv.PointVector, center image.Point, bbox image.Rectangle) gocv.Mat {
	out := src.Clone()

	// Contorno del usuario en rojo
	pv := gocv.NewPointsVector()
	defer pv.Close()
	pv.Append(contour)
	gocv.DrawContours(&out, pv, -1, color.RGBA{255, 0, 0, 0}, 2)

	// Overlay “ideal”/referencia en verde
	switch shape {
	case ShapeCircle:
		cx, cy, r := gocv.MinEnclosingCircle(contour)
		gocv.Circle(&out, image.Pt(int(cx), int(cy)), int(r), color.RGBA{0, 255, 0, 0}, 2)
	default:
		gocv.Rectangle(&out, bbox, color.RGBA{0, 255, 0, 0}, 2)
		gocv.Circle(&out, center, 3, color.RGBA{0, 255, 0, 0}, -1)
	}
	return out
}

// ...existing code...
func isContourConvex(pts []image.Point) bool {
	if len(pts) < 3 {
		return false
	}
	sign := 0
	n := len(pts)
	for i := 0; i < n; i++ {
		dx1 := pts[(i+1)%n].X - pts[i].X
		dy1 := pts[(i+1)%n].Y - pts[i].Y
		dx2 := pts[(i+2)%n].X - pts[(i+1)%n].X
		dy2 := pts[(i+2)%n].Y - pts[(i+1)%n].Y
		zcrossproduct := dx1*dy2 - dy1*dx2
		if zcrossproduct != 0 {
			if sign == 0 {
				sign = zcrossproduct
			} else if (sign > 0) != (zcrossproduct > 0) {
				return false
			}
		}
	}
	return true
}

// ...existing code...
