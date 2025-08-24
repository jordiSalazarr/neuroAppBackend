package services

import (
	"fmt"

	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	"neuro.app.jordi/internal/evaluation/services/bvmtcv"
)

type LocalTemplateResolver struct{ Root string }

func NewLocalTemplateResolver() LocalTemplateResolver {
	// var candidates []string

	// if wd, err := os.Getwd(); err == nil {
	// 	candidates = append(candidates, filepath.Join(wd, "images"))
	// }
	// if exe, err := os.Executable(); err == nil {
	// 	candidates = append(candidates, filepath.Join(filepath.Dir(exe), "images"))
	// }
	// candidates = append(candidates, "images") // fallback relativo

	// for _, p := range candidates {
	// 	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
	// 		return LocalTemplateResolver{Root: p}
	// 	}
	// }

	// Si ninguna existe, devuelve la primera candidata (creará errores claros en Resolve)
	return LocalTemplateResolver{Root: "/Users/jordisalazarbadia/Desktop/NeuroApp/back/images/"}
}

func (r LocalTemplateResolver) Resolve(name string) (string, error) {
	// if strings.TrimSpace(name) == "" {
	// 	return "", fmt.Errorf("figure_name vacío")
	// }

	// // Limpia y evita traversal: quita separador inicial y normaliza
	// clean := filepath.Clean(name)
	// clean = strings.TrimPrefix(clean, string(filepath.Separator))

	// // Directorios candidatos: raíz y, si existe, assets/bvmt dentro de raíz
	// candidates := []string{
	// 	r.Root,
	// }

	// // Probar nombre tal cual y con extensiones típicas
	// exts := []string{"", ".png", ".jpg", ".jpeg"}

	// for _, base := range candidates {
	// 	for _, ext := range exts {
	// 		p := filepath.Join(base, clean+ext)

	// 		// Seguridad: asegúrate de que p cae dentro de base
	// 		absBase, _ := filepath.Abs(base)
	// 		absP, _ := filepath.Abs(p)
	// 		if !strings.HasPrefix(absP, absBase+string(filepath.Separator)) && absP != absBase {
	// 			continue
	// 		}

	// 		// Existe y no es directorio
	// 		if info, err := os.Stat(p); err == nil && !info.IsDir() {
	// 			return p, nil
	// 		}
	// 	}
	// }

	url := fmt.Sprintf("%s%s", r.Root, name)
	return url, nil

}

type OpenCVBVMTScorer struct{}

func (OpenCVBVMTScorer) Score(tpl, pat string) (VIMdomain.BVMTScore, error) {

	return bvmtcv.ScoreBVMT(tpl, pat)
}
