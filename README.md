# 🧠 NeuroApp — Batería Neurológica Digital

> Evaluaciones neurocognitivas validadas en español, scoring automatizado y reportes clínicos listos para compartir.

<p align="center">
  <img src="./.github/hero.png" alt="NeuroApp cover" width="800"/>
</p>

<p align="center">
  <a href="#arquitectura"><img alt="Arquitectura" src="https://img.shields.io/badge/arquitectura-hexagonal-blue"></a>
  <a href="#stack"><img alt="Go" src="https://img.shields.io/badge/Go-1.22-00ADD8?logo=go"></a>
  <a href="#seguridad--cumplimiento"><img alt="GDPR" src="https://img.shields.io/badge/GDPR-ready-0A0"></a>
  <a href="#ci--test"><img alt="Tests" src="https://img.shields.io/badge/tests-race%20%2B%20coverage-informational"></a>
  <a href="#licencia"><img alt="License" src="https://img.shields.io/badge/license-MIT-lightgrey"></a>
</p>

---

## ✨ TL;DR

NeuroApp digitaliza subtests neuropsicológicos (HVLT-R, BVMT-R, TMT, Clock Drawing, Letter Cancellation, Fluencia Semántica, etc.), automatiza el scoring, genera PDFs clínicos y permite flujo paciente–especialista con vídeo corto y mensajería.

---

## 🚀 Características

* **Batería neuropsicológica validada en población española.**
* **Scoring automático** con reglas clínicas y rangos normativos.
* **Reportes PDF** con firma y envío por **email (SES)**; almacenamiento en **S3**.
* **Vídeos breves** para apoyo diagnóstico (upload seguro y auditado).
* **Perfiles** Paciente / Especialista / Admin.
* **Historial cronológico** de evaluaciones y materiales.
* **Multi‑idioma** (ES/CAT/EN) y **accesibilidad**.
* **Arquitectura hexagonal + DDD**, tests unitarios y de dominio.

> **Ámbito clínico**: Parkinson, deterioro cognitivo, cribado, seguimiento.

---

## 🧩 Stack

* **Backend:** Go (Gin), DDD + hexagonal, SQLBoiler, migraciones.
* **DB:** PostgreSQL.
* **Frontend:** React Native (Expo) y Next.js (web) *(si aplica en este repo)*.
* **Infra:** AWS (Fargate para API), **Lambda** para generación de PDF, **S3** (archivos), **SES** (emails), **Cognito** (auth).
* **CI:** GitHub Actions (build + test + coverage).

---

## 🗂️ Estructura (ejemplo)

```
.
├── cmd/
│   └── api/                  # Entrypoint servidor HTTP
├── internal/                 # Dominio y aplicación (DDD)
│   ├── evaluation/           # Casos de uso y servicios
│   ├── ...
│   └── platform/             # Adapters (db, email, storage, auth)
├── pkg/                      # Utilidades compartidas
├── migrations/               # SQL
├── web/ or apps/mobile/      # Frontend si está en monorepo
├── .github/workflows/        # CI
└── README.md
```

---

## 🏗️ Arquitectura

```mermaid
flowchart LR
  UI[Paciente / Especialista\nWeb & Mobile] --> API
  subgraph Backend (Go)
    API[HTTP API (Gin)] --> APP[Application Layer\nUse Cases]
    APP --> DOM[Domain (DDD)]
    APP --> ADP[(Adapters)]
    ADP --> DB[(PostgreSQL)]
    ADP --> S3[(S3 - Storage)]
    ADP --> SES[(SES - Email)]
    ADP --> COG[(Cognito - Auth)]
    APP --> EVT[(Events/Queue optional)]
  end
  PDF[Lambda PDF\n(S3 + HTML->PDF)] --> S3
```

---

## ⚙️ Configuración local

### Requisitos

* Go **1.22+**
* MYSQL **13+**
* Node + npm **18/20** (frontend)

### Variables de entorno (ejemplo)

Crea `.env` (o usa tu gestor de secretos):

```
ENV=local
DB_DSN=postgres://user:pass@localhost:5432/neuroapp?sslmode=disable
AWS_REGION=eu-west-1
S3_BUCKET=neuroapp-reports
SES_SENDER=noreply@tu-dominio.com
COGNITO_POOL_ID=eu-west-1_XXXX
COGNITO_CLIENT_ID=YYYY
```

### Migraciones & arranque

```bash
# 1) Levanta Postgres y aplica migraciones
make migrate          # o tu herramienta favorita

# 2) Ejecuta API
go run ./cmd/api

# 3) (Opcional) Frontend
cd web && npm i && npm run dev
```

---

## 🧪 CI & Test

### Ejecutar tests con race + coverage **solo** sobre `internal`

```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./internal/...
```

Generar HTML del coverage y abrirlo:

```bash
go tool cover -html=coverage.out -o coverage.html
# Abre coverage.html en tu navegador
```

### GitHub Actions (ejemplo mínimo)

```yaml
name: CI
on:
  push:
    branches: [ "main", "develop" ]
permissions:
  contents: read
env:
  GO_VERSION: "1.22.x"
  CGO_ENABLED: "1"   # para -race
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      - name: Download modules
        run: go mod download
      - name: Lint (opcional)
        run: |
          go vet ./...
      - name: Test with coverage (internal only)
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./internal/...
      - name: Upload coverage artifact
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out
```

> Si quieres **badge de cobertura**, integra Codecov o Coveralls y añade su action.

---

## 🔌 Endpoints & Docs

* Documentación **OpenAPI/Swagger**: `./api/openapi.yaml` *(placeholder)*.
* Postman/Insomnia collections en `./docs/api/` *(placeholder).*

---

## 🔐 Seguridad & Cumplimiento

* **GDPR**: minimización de datos, consentimiento informado, **DPA** con proveedores, derecho de acceso/rectificación/borrado.
* **Cifrado** en tránsito (TLS) y en reposo (S3 SSE, RDS KMS si aplica).
* **Logs sin datos clínicos** (PII/PHI redactada/anónima), niveles de acceso y auditoría.
* **Backups** y retención configurable.
* Preparado para **marcado CE** (MDR) como software de apoyo *(orientativo; requiere validación clínica y proceso regulatorio formal).*

---

## 🗺️ Roadmap (extracto)

* [ ] Escalas adicionales (ej. NBACE módulos opcionales)
* [ ] Panel admin y auditoría avanzada
* [ ] Exportación FHIR/HL7
* [ ] Modelos de ayuda a decisión (LLM + RAG) con trazabilidad
* [ ] Modo offline en móvil

---

## 🤝 Contribuir

1. Haz fork y crea rama desde `main`.
2. Asegura **tests** y estilo (lint/vet).
3. Pull Request con descripción clara (**qué**, **por qué**, **cómo testear**).

Consulta `CONTRIBUTING.md` y `CODE_OF_CONDUCT.md` *(placeholders).*

---

## 📸 Capturas / Demo

* `./.github/screenshots/` *(añade tus PNG/GIF)*
* Demo corta (≤60s) en `./.github/demo.mp4`

---

## 📄 Licencia

MIT © 2025 — *Cámbiala si el proyecto es propietario/privado.*

---

## 📬 Contacto

**Equipo NeuroApp**
Soporte: [support@tu-dominio.com](mailto:support@tu-dominio.com)
Clínico: [dr.salazar@tu-dominio.com](mailto:dr.salazar@tu-dominio.com)

> ¿Quieres que personalicemos este README con tu logo, enlaces reales y badges de cobertura? Abre un issue o contáctanos.
