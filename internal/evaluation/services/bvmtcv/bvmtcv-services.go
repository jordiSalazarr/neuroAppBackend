package bvmtcv

import (
	"errors"
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"

	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	"neuro.app.jordi/internal/evaluation/utils"
)

// ScoreBVMT calcula un score 0..100 usando métricas sobre BORDES (robusto a grosor).
func ScoreBVMT(templatePath, patientPath string) (VIMdomain.BVMTScore, error) {
	const outW, outH = 512, 512

	// 1) Carga + preprocesado (gris, máscara binaria, bordes dilatados)
	refGray, refMask, refEdges, err := loadAndPrepAll(templatePath, outW, outH)
	if err != nil {
		return VIMdomain.BVMTScore{}, fmt.Errorf("prep template: %w", err)
	}
	defer refGray.Close()
	defer refMask.Close()
	defer refEdges.Close()

	testGray, testMask, testEdges, err := loadAndPrepAll(patientPath, outW, outH)
	if err != nil {
		return VIMdomain.BVMTScore{}, fmt.Errorf("prep patient: %w", err)
	}
	defer testGray.Close()
	defer testMask.Close()
	defer testEdges.Close()

	// 2) Alineación: primero ORB; si falla o pocos inliers → fallback por bbox
	size := image.Pt(refGray.Cols(), refGray.Rows())

	alignedGray, H, _, inliers, err := alignORB(refGray, testGray)
	if err != nil || inliers < 10 {
		// Fallback geométrico: centra y escala por bounding box
		M, ok := similarityByBBox(refMask, testMask)
		if !ok {
			// último recurso: sin alineación
			alignedGray = testGray.Clone()
		} else {
			if !alignedGray.Empty() {
				alignedGray.Close()
			}
			alignedGray = gocv.NewMat()
			gocv.WarpAffine(testGray, &alignedGray, M, size)
		}
		H = gocv.NewMat()
	}
	defer alignedGray.Close()

	// Warp de bordes con lo que tengamos
	alignedEdges := gocv.NewMat()
	if !H.Empty() {
		gocv.WarpPerspective(testEdges, &alignedEdges, H, size)
	} else {
		// si venimos del fallback, intenta con M (bbox)
		M, ok := similarityByBBox(refMask, testMask)
		if ok {
			gocv.WarpAffine(testEdges, &alignedEdges, M, size)
		} else {
			alignedEdges = testEdges.Clone()
		}
	}
	defer alignedEdges.Close()

	// 3) Métricas sobre bordes
	dice := diceBinary(refEdges, alignedEdges) // 0..1 (preferible a IoU en bordes finos)
	ssim := ssimGlobal(refEdges, alignedEdges) // ~[-1..1]
	psnr := psnrGray(refEdges, alignedEdges)   // dB

	ssim01 := utils.Clamp01((ssim + 1.0) / 2.0)
	psnr01 := utils.Clamp01(psnr / 40.0)

	// 4) Score final (pondera Dice)
	score01 := utils.Clamp01(0.70*dice + 0.20*ssim01 + 0.10*psnr01)
	final := int(math.Round(100.0 * score01))

	return VIMdomain.BVMTScore{
		IoU:        dice, // guardamos DICE en el campo IoU (documenta este matiz si hace falta)
		SSIM:       ssim, // crudo; si prefieres 0..1, guarda ssim01
		PSNR:       psnr,
		FinalScore: final,
	}, nil
}

/* ---------------------- Nuevos/actualizados helpers ---------------------- */

// Carga cuidando alpha; devuelve: gris, máscara binaria (figura), y bordes dilatados.
func loadAndPrepAll(path string, w, h int) (gray, mask, edges gocv.Mat, err error) {
	img := gocv.IMRead(path, gocv.IMReadUnchanged) // conserva alpha si existe
	if img.Empty() {
		return gocv.NewMat(), gocv.NewMat(), gocv.NewMat(), fmt.Errorf("cannot read %s", path)
	}
	defer img.Close()

	// resize
	res := gocv.NewMat()
	gocv.Resize(img, &res, image.Pt(w, h), 0, 0, gocv.InterpolationArea)
	defer res.Close()

	// Si BGRA, usa alpha como máscara; si no, umbraliza gris.
	if res.Channels() == 4 {
		planes := gocv.Split(res)
		defer func() {
			for _, p := range planes {
				p.Close()
			}
		}()

		// gray desde BGR
		gray = gocv.NewMat()
		gocv.CvtColor(res, &gray, gocv.ColorBGRAToBGR) //ColorBGRA2BGR
		g := gocv.NewMat()
		gocv.CvtColor(gray, &g, gocv.ColorBGRToGray)
		gray.Close()
		gray = g

		// máscara desde alpha (>0 → 255)
		mask = gocv.NewMat()
		gocv.Threshold(planes[3], &mask, 0, 255, gocv.ThresholdBinary)
	} else {
		gray = gocv.NewMat()
		gocv.CvtColor(res, &gray, gocv.ColorBGRToGray)

		// Otsu invertido → figura blanca (255) sobre negro
		mask = gocv.NewMat()
		blur := gocv.NewMat()
		gocv.GaussianBlur(gray, &blur, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
		gocv.Threshold(blur, &mask, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)
		blur.Close()
	}

	// Limpieza morfológica ligera
	k3 := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	gocv.MorphologyEx(mask, &mask, gocv.MorphClose, k3)
	k3.Close()

	// Bordes (Canny) + pequeña dilatación → robusto a grosor
	edges = gocv.NewMat()
	gocv.Canny(gray, &edges, 50, 150)
	kd := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	gocv.Dilate(edges, &edges, kd)
	kd.Close()

	// Opcional: constriñelos con la máscara (por si hay ruido fuera)
	tmp := gocv.NewMat()
	gocv.BitwiseAnd(edges, mask, &tmp)
	edges.Close()
	edges = tmp

	return gray, mask, edges, nil
}

// Fallback: transforma por similitud (escala + traslación) usando bounding boxes mayores.
func similarityByBBox(refMask, testMask gocv.Mat) (M gocv.Mat, ok bool) {
	rc, rw, rh, rok := bboxCenter(refMask)
	tc, tw, th, tok := bboxCenter(testMask)
	if !rok || !tok || rw == 0 || rh == 0 || tw == 0 || th == 0 {
		return gocv.NewMat(), false
	}

	// Escala por el máximo lado (evita distorsión por aspect ratio)
	refSize := math.Max(rw, rh)
	testSize := math.Max(tw, th)
	scale := refSize / testSize

	// Matriz de similitud (rotación 0): [ s 0 tx; 0 s ty ]
	M = gocv.NewMatWithSize(2, 3, gocv.MatTypeCV64F)
	M.SetDoubleAt(0, 0, scale)
	M.SetDoubleAt(0, 1, 0)
	M.SetDoubleAt(1, 0, 0)
	M.SetDoubleAt(1, 1, scale)

	// tx, ty para llevar el centro de test al centro de ref
	tx := float64(rc.X) - (scale * float64(tc.X))
	ty := float64(rc.Y) - (scale * float64(tc.Y))
	M.SetDoubleAt(0, 2, tx)
	M.SetDoubleAt(1, 2, ty)

	return M, true
}

// Encuentra centro y ancho/alto del mayor contorno en una máscara binaria.
func bboxCenter(mask gocv.Mat) (center image.Point, w, h float64, ok bool) {
	if mask.Empty() {
		return image.Point{}, 0, 0, false
	}

	contours := gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	// según versión de GoCV, PointsVector puede tener Close(); si existe, puedes defer contours.Close()

	n := contours.Size()
	if n == 0 {
		return image.Point{}, 0, 0, false
	}

	maxA := 0.0
	var best image.Rectangle

	for i := 0; i < n; i++ {
		c := contours.At(i)      // c es gocv.PointVector
		a := gocv.ContourArea(c) // área del contorno i
		if a > maxA {
			maxA = a
			best = gocv.BoundingRect(c) // rectángulo envolvente del contorno i
		}
	}

	if maxA <= 0 {
		return image.Point{}, 0, 0, false
	}

	cx := best.Min.X + best.Dx()/2
	cy := best.Min.Y + best.Dy()/2
	return image.Pt(cx, cy), float64(best.Dx()), float64(best.Dy()), true
}

// Dice (0..1) para binarios.
func diceBinary(a, b gocv.Mat) float64 {
	if a.Empty() || b.Empty() || a.Rows() != b.Rows() || a.Cols() != b.Cols() {
		return 0
	}
	and := gocv.NewMat()
	defer and.Close()
	gocv.BitwiseAnd(a, b, &and)

	na := float64(gocv.CountNonZero(a))
	nb := float64(gocv.CountNonZero(b))
	nab := float64(gocv.CountNonZero(and))
	if na+nb == 0 {
		return 1
	}
	return (2.0 * nab) / (na + nb)
}

func iouBinary(a, b gocv.Mat) float64 {
	if a.Empty() || b.Empty() || a.Rows() != b.Rows() || a.Cols() != b.Cols() {
		return 0
	}
	and := gocv.NewMat()
	or := gocv.NewMat()
	defer and.Close()
	defer or.Close()

	gocv.BitwiseAnd(a, b, &and)
	gocv.BitwiseOr(a, b, &or)

	inter := float64(gocv.CountNonZero(and))
	union := float64(gocv.CountNonZero(or))
	if union == 0 {
		if inter == 0 {
			return 1 // ambas vacías → IoU perfecto
		}
		return 0
	}
	return inter / union
}

func ssimGlobal(a, b gocv.Mat) float64 {
	if a.Empty() || b.Empty() || a.Rows() != b.Rows() || a.Cols() != b.Cols() {
		return 0
	}
	rows, cols := a.Rows(), a.Cols()
	N := float64(rows * cols)

	var sumX, sumY, sumXX, sumYY, sumXY float64
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			ax := float64(a.GetUCharAt(y, x))
			by := float64(b.GetUCharAt(y, x))
			sumX += ax
			sumY += by
			sumXX += ax * ax
			sumYY += by * by
			sumXY += ax * by
		}
	}
	muX := sumX / N
	muY := sumY / N
	varX := sumXX/N - muX*muX
	varY := sumYY/N - muY*muY
	covXY := sumXY/N - muX*muY

	L := 255.0
	c1 := (0.01 * L) * (0.01 * L)
	c2 := (0.03 * L) * (0.03 * L)

	num := (2*muX*muY + c1) * (2*covXY + c2)
	den := (muX*muX + muY*muY + c1) * (varX + varY + c2)
	if den == 0 {
		return 1
	}
	return num / den
}

func psnrGray(a, b gocv.Mat) float64 {
	mse := mseGray(a, b)
	if mse <= 1e-12 {
		return 99.0
	}
	return 10.0 * math.Log10((255.0*255.0)/mse)
}

func mseGray(a, b gocv.Mat) float64 {
	if a.Empty() || b.Empty() || a.Rows() != b.Rows() || a.Cols() != b.Cols() {
		return 1e9
	}
	rows, cols := a.Rows(), a.Cols()
	N := float64(rows * cols)
	var sum float64
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			d := float64(a.GetUCharAt(y, x)) - float64(b.GetUCharAt(y, x))
			sum += d * d
		}
	}
	return sum / N
}

func Clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func alignORB(refGray, testGray gocv.Mat) (aligned gocv.Mat, H gocv.Mat, matches, inliers int, err error) {
	orb := gocv.NewORB()
	defer orb.Close()

	kp1, desc1 := orb.DetectAndCompute(refGray, gocv.NewMat())
	kp2, desc2 := orb.DetectAndCompute(testGray, gocv.NewMat())
	defer desc1.Close()
	defer desc2.Close()

	if desc1.Empty() || desc2.Empty() {
		return gocv.NewMat(), gocv.NewMat(), 0, 0, errors.New("no features")
	}

	bf := gocv.NewBFMatcher()
	defer bf.Close()

	mm := bf.KnnMatch(desc1, desc2, 2)
	if len(mm) == 0 {
		return gocv.NewMat(), gocv.NewMat(), 0, 0, errors.New("no matches")
	}

	var srcPts, dstPts []gocv.Point2f
	for _, pair := range mm {
		if len(pair) < 2 {
			continue
		}
		m, n := pair[0], pair[1]
		if m.Distance < 0.75*n.Distance {
			p1 := gocv.Point2f{X: float32(kp1[m.QueryIdx].X), Y: float32(kp1[m.QueryIdx].Y)}
			p2 := gocv.Point2f{X: float32(kp2[m.TrainIdx].X), Y: float32(kp2[m.TrainIdx].Y)}
			srcPts = append(srcPts, p1)
			dstPts = append(dstPts, p2)
		}
	}

	if len(srcPts) < 4 {
		return gocv.NewMat(), gocv.NewMat(), len(mm), 0, errors.New("not enough inliers")
	}

	srcMat := gocv.NewMatWithSize(len(srcPts), 2, gocv.MatTypeCV32F)
	dstMat := gocv.NewMatWithSize(len(dstPts), 2, gocv.MatTypeCV32F)
	for i, p := range srcPts {
		srcMat.SetFloatAt(i, 0, p.X)
		srcMat.SetFloatAt(i, 1, p.Y)
	}
	for i, p := range dstPts {
		dstMat.SetFloatAt(i, 0, p.X)
		dstMat.SetFloatAt(i, 1, p.Y)
	}
	defer srcMat.Close()
	defer dstMat.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	// Homografía con RANSAC (umbral 3.0, 2000 iteraciones, confianza 0.995)
	H = gocv.FindHomography(
		srcMat,
		dstMat,
		gocv.HomographyMethodRANSAC, // según versión también puede ser gocv.RANSAC
		3.0,                         // ransacReprojThreshold
		&mask,
		2000,  // maxIters
		0.995, // confidence
	)

	inliers = gocv.CountNonZero(mask)
	matches = len(srcPts)

	aligned = gocv.NewMat()
	gocv.WarpPerspective(testGray, &aligned, H, image.Pt(refGray.Cols(), refGray.Rows()))

	return aligned, H, matches, inliers, nil
}
