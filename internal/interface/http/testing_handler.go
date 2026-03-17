package http

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	apptesting "mltestsuite/internal/application/testing"
	appuser "mltestsuite/internal/application/user"
	domain "mltestsuite/internal/domain/testing"
	"mltestsuite/internal/infrastructure/cloudinary"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// TestingHandler maneja todos los endpoints de testing (teams, reports, testcases, releases, executions, knowledge).
type TestingHandler struct {
	service     *apptesting.Service
	userService *appuser.Service
	uploader    *cloudinary.Uploader
	renderer    *Renderer
}

// NewTestingHandler crea el handler de testing.
func NewTestingHandler(service *apptesting.Service, userService *appuser.Service, uploader *cloudinary.Uploader, renderer *Renderer) *TestingHandler {
	return &TestingHandler{service: service, userService: userService, uploader: uploader, renderer: renderer}
}

// parseUUID parsea un path value como UUID, retorna 400 si es invalido.
func parseUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	idStr := r.PathValue(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

// -- Teams -------------------------------------------------------------------

type teamVM struct {
	ID          interface{}
	Name        string
	Description string
	Members     []string
}

func (h *TestingHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.service.ListTeams(r.Context())
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	users, _ := h.userService.ListUsers(r.Context())

	// Group member names by team ID
	membersByTeam := map[uuid.UUID][]string{}
	for _, u := range users {
		if u.TeamID != nil {
			membersByTeam[*u.TeamID] = append(membersByTeam[*u.TeamID], u.Name)
		}
	}

	vms := make([]teamVM, len(teams))
	for i, t := range teams {
		vms[i] = teamVM{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Members:     membersByTeam[t.ID],
		}
	}
	h.renderer.ExecuteTemplate(w, "teams/list.html", withFlash(w, r, map[string]any{
		"Teams": vms,
	}))
}

func (h *TestingHandler) ShowCreateTeam(w http.ResponseWriter, r *http.Request) {
	h.renderer.ExecuteTemplate(w, "teams/new.html", withFlash(w, r, map[string]any{}))
}

func (h *TestingHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	description := r.FormValue("description")
	if name == "" {
		setFlash(w, "error", "El nombre es obligatorio")
		http.Redirect(w, r, "/teams/new", http.StatusSeeOther)
		return
	}
	if err := h.service.CreateTeam(r.Context(), name, description); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/teams/new", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Equipo creado correctamente")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

func (h *TestingHandler) ShowEditTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	team, err := h.service.GetTeam(r.Context(), id)
	if err != nil {
		http.Error(w, "Equipo no encontrado", http.StatusNotFound)
		return
	}
	h.renderer.ExecuteTemplate(w, "teams/edit.html", withFlash(w, r, map[string]any{
		"Team": team,
	}))
}

func (h *TestingHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	name := r.FormValue("name")
	description := r.FormValue("description")
	if name == "" {
		setFlash(w, "error", "El nombre es obligatorio")
		http.Redirect(w, r, fmt.Sprintf("/teams/%s/edit", id), http.StatusSeeOther)
		return
	}
	if err := h.service.UpdateTeam(r.Context(), id, name, description); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, fmt.Sprintf("/teams/%s/edit", id), http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Equipo actualizado")
	http.Redirect(w, r, "/teams", http.StatusSeeOther)
}

// -- Reports -----------------------------------------------------------------

func (h *TestingHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.ListReports(r.Context())
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	h.renderer.ExecuteTemplate(w, "reports/list.html", withFlash(w, r, map[string]any{
		"Reports": reports,
	}))
}

func (h *TestingHandler) ShowCreateReport(w http.ResponseWriter, r *http.Request) {
	teams, _ := h.service.ListTeams(r.Context())
	h.renderer.ExecuteTemplate(w, "reports/new.html", withFlash(w, r, map[string]any{
		"Teams": teams,
	}))
}

func (h *TestingHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(r.FormValue("team_id"))
	if err != nil {
		setFlash(w, "error", "Equipo inválido")
		http.Redirect(w, r, "/reports/new", http.StatusSeeOther)
		return
	}
	name := r.FormValue("name")
	reportType := r.FormValue("report_type")
	description := r.FormValue("description")
	if err := h.service.CreateReport(r.Context(), teamID, name, reportType, description); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/reports/new", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Reporte creado correctamente")
	http.Redirect(w, r, "/reports", http.StatusSeeOther)
}

func (h *TestingHandler) ShowEditReport(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	report, err := h.service.GetReport(r.Context(), id)
	if err != nil {
		http.Error(w, "Reporte no encontrado", http.StatusNotFound)
		return
	}
	teams, _ := h.service.ListTeams(r.Context())
	h.renderer.ExecuteTemplate(w, "reports/edit.html", withFlash(w, r, map[string]any{
		"Report": report,
		"Teams":  teams,
	}))
}

func (h *TestingHandler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	teamID, err := uuid.Parse(r.FormValue("team_id"))
	if err != nil {
		setFlash(w, "error", "Equipo inválido")
		http.Redirect(w, r, fmt.Sprintf("/reports/%s/edit", id), http.StatusSeeOther)
		return
	}
	name := r.FormValue("name")
	reportType := r.FormValue("report_type")
	description := r.FormValue("description")
	if err := h.service.UpdateReport(r.Context(), id, teamID, name, reportType, description); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, fmt.Sprintf("/reports/%s/edit", id), http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Reporte actualizado")
	http.Redirect(w, r, "/reports", http.StatusSeeOther)
}

// -- TestCases ---------------------------------------------------------------

type reportSummaryVM struct {
	ReportID   uuid.UUID
	ReportName string
	ReportType string
	TeamName   string
	CaseCount  int
	Creators   string
}

func (h *TestingHandler) ListTestCases(w http.ResponseWriter, r *http.Request) {
	tcs, err := h.service.ListTestCases(r.Context())
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	// Group by report
	orderMap := []uuid.UUID{}
	byReport := map[uuid.UUID]*reportSummaryVM{}
	creatorsByReport := map[uuid.UUID]map[string]struct{}{}
	for _, tc := range tcs {
		if _, ok := byReport[tc.ReportID]; !ok {
			orderMap = append(orderMap, tc.ReportID)
			byReport[tc.ReportID] = &reportSummaryVM{
				ReportID:   tc.ReportID,
				ReportName: tc.ReportName,
				ReportType: tc.ReportType,
				TeamName:   tc.TeamName,
			}
			creatorsByReport[tc.ReportID] = map[string]struct{}{}
		}
		byReport[tc.ReportID].CaseCount++
		if tc.CreatedByName != "" {
			creatorsByReport[tc.ReportID][tc.CreatedByName] = struct{}{}
		}
	}

	summaries := make([]reportSummaryVM, 0, len(orderMap))
	for _, id := range orderMap {
		vm := byReport[id]
		names := make([]string, 0, len(creatorsByReport[id]))
		for n := range creatorsByReport[id] {
			names = append(names, n)
		}
		vm.Creators = strings.Join(names, ", ")
		summaries = append(summaries, *vm)
	}

	h.renderer.ExecuteTemplate(w, "testcases/list.html", withFlash(w, r, map[string]any{
		"Reports": summaries,
	}))
}

func (h *TestingHandler) ListTestCasesByReport(w http.ResponseWriter, r *http.Request) {
	reportID, ok := parseUUID(w, r, "reportID")
	if !ok {
		return
	}
	tcs, err := h.service.ListTestCasesByReport(r.Context(), reportID)
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	// Infer report meta from first case (or empty)
	var reportName, reportType, teamName string
	if len(tcs) > 0 {
		reportName = tcs[0].ReportName
		reportType = tcs[0].ReportType
		teamName = tcs[0].TeamName
	} else {
		// Fallback: fetch report directly
		rep, err := h.service.GetReport(r.Context(), reportID)
		if err == nil {
			reportName = rep.Name
			reportType = rep.ReportType
		}
	}

	h.renderer.ExecuteTemplate(w, "testcases/by_report.html", withFlash(w, r, map[string]any{
		"TestCases":  tcs,
		"ReportName": reportName,
		"ReportType": reportType,
		"TeamName":   teamName,
	}))
}

func (h *TestingHandler) ShowCreateTestCase(w http.ResponseWriter, r *http.Request) {
	reports, _ := h.service.ListReports(r.Context())
	h.renderer.ExecuteTemplate(w, "testcases/new.html", withFlash(w, r, map[string]any{
		"Reports":    reports,
		"Priorities": []domain.Priority{domain.PriorityHigh, domain.PriorityMedium, domain.PriorityLow},
	}))
}

func (h *TestingHandler) CreateTestCase(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(r.FormValue("report_id"))
	if err != nil {
		setFlash(w, "error", "Reporte inválido")
		http.Redirect(w, r, "/testcases/new", http.StatusSeeOther)
		return
	}
	userID, err := uuid.Parse(GetUserID(r.Context()))
	if err != nil {
		http.Error(w, "Usuario inválido", http.StatusBadRequest)
		return
	}
	input := apptesting.CreateTestCaseInput{
		ReportID:       reportID,
		Title:          r.FormValue("title"),
		Preconditions:  r.FormValue("preconditions"),
		Steps:          r.FormValue("steps"),
		ExpectedResult: r.FormValue("expected_result"),
		Priority:       domain.Priority(r.FormValue("priority")),
		CreatedByID:    userID,
	}
	if _, err := h.service.CreateTestCase(r.Context(), input); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/testcases/new", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Caso de prueba creado")
	http.Redirect(w, r, "/testcases", http.StatusSeeOther)
}

func (h *TestingHandler) ShowTestCase(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	tc, err := h.service.GetTestCase(r.Context(), id)
	if err != nil {
		http.Error(w, "Caso no encontrado", http.StatusNotFound)
		return
	}
	h.renderer.ExecuteTemplate(w, "testcases/show.html", withFlash(w, r, map[string]any{
		"TestCase": tc,
	}))
}

func (h *TestingHandler) ShowEditTestCase(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	tc, err := h.service.GetTestCase(r.Context(), id)
	if err != nil {
		http.Error(w, "Caso no encontrado", http.StatusNotFound)
		return
	}
	reports, _ := h.service.ListReports(r.Context())
	h.renderer.ExecuteTemplate(w, "testcases/edit.html", withFlash(w, r, map[string]any{
		"TestCase":      tc,
		"Reports":       reports,
		"Priorities":    []domain.Priority{domain.PriorityHigh, domain.PriorityMedium, domain.PriorityLow},
		"FromExecution": r.URL.Query().Get("from_execution"),
	}))
}

func (h *TestingHandler) UpdateTestCase(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	reportID, err := uuid.Parse(r.FormValue("report_id"))
	if err != nil {
		setFlash(w, "error", "Reporte inválido")
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s/edit", id), http.StatusSeeOther)
		return
	}
	userID, _ := uuid.Parse(GetUserID(r.Context()))
	input := apptesting.CreateTestCaseInput{
		ReportID:       reportID,
		Title:          r.FormValue("title"),
		Preconditions:  r.FormValue("preconditions"),
		Steps:          r.FormValue("steps"),
		ExpectedResult: r.FormValue("expected_result"),
		Priority:       domain.Priority(r.FormValue("priority")),
		CreatedByID:    userID,
	}
	if err := h.service.UpdateTestCase(r.Context(), id, input); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s/edit", id), http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Caso de prueba actualizado")
	if execID := r.FormValue("from_execution"); execID != "" {
		http.Redirect(w, r, fmt.Sprintf("/executions/%s", execID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
}

func (h *TestingHandler) DeleteTestCase(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteTestCase(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *TestingHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		setFlash(w, "error", "Error al procesar archivo")
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		setFlash(w, "error", "No se encontró el archivo")
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
		return
	}
	defer file.Close()

	result, err := h.uploader.Upload(r.Context(), file, "mltestsuite")
	if err != nil {
		slog.Error("cloudinary upload error", "error", err)
		setFlash(w, "error", "Error al subir la imagen")
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
		return
	}

	if err := h.service.UpdateTestCaseImage(r.Context(), id, result.SecureURL); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Imagen subida correctamente")
	http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
}

func (h *TestingHandler) SaveField(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	fieldName := r.FormValue("field_name")
	expectedJSON := r.FormValue("expected_json")
	if fieldName == "" || expectedJSON == "" {
		setFlash(w, "error", "Nombre del campo y JSON son obligatorios")
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
		return
	}
	if err := h.service.SaveTestCaseField(r.Context(), id, fieldName, expectedJSON); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Campo JSON guardado")
	http.Redirect(w, r, fmt.Sprintf("/testcases/%s", id), http.StatusSeeOther)
}

func (h *TestingHandler) DeleteField(w http.ResponseWriter, r *http.Request) {
	fieldID, ok := parseUUID(w, r, "fieldID")
	if !ok {
		return
	}
	if err := h.service.DeleteTestCaseField(r.Context(), fieldID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// -- Import/Export -----------------------------------------------------------

func (h *TestingHandler) ShowImport(w http.ResponseWriter, r *http.Request) {
	reports, _ := h.service.ListReports(r.Context())
	h.renderer.ExecuteTemplate(w, "testcases/import.html", withFlash(w, r, map[string]any{
		"Reports": reports,
	}))
}

func (h *TestingHandler) ImportMarkdown(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(r.FormValue("report_id"))
	if err != nil {
		setFlash(w, "error", "Reporte inválido")
		http.Redirect(w, r, "/testcases/import", http.StatusSeeOther)
		return
	}
	userID, _ := uuid.Parse(GetUserID(r.Context()))
	content := r.FormValue("markdown")
	if content == "" {
		setFlash(w, "error", "El contenido Markdown es obligatorio")
		http.Redirect(w, r, "/testcases/import", http.StatusSeeOther)
		return
	}
	count, err := h.service.ImportFromMarkdown(r.Context(), reportID, content, userID)
	if err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/testcases/import", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", fmt.Sprintf("Se importaron %d casos de prueba", count))
	http.Redirect(w, r, "/testcases", http.StatusSeeOther)
}

func (h *TestingHandler) ExportClaudePrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	prompt, err := h.service.ExportClaudePrompt(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=claude_prompt.md")
	fmt.Fprint(w, prompt)
}

// -- Releases ----------------------------------------------------------------

func (h *TestingHandler) ListReleases(w http.ResponseWriter, r *http.Request) {
	releases, err := h.service.ListReleases(r.Context())
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	h.renderer.ExecuteTemplate(w, "releases/list.html", withFlash(w, r, map[string]any{
		"Releases": releases,
	}))
}

func (h *TestingHandler) ShowCreateRelease(w http.ResponseWriter, r *http.Request) {
	h.renderer.ExecuteTemplate(w, "releases/new.html", withFlash(w, r, map[string]any{}))
}

func (h *TestingHandler) CreateRelease(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(GetUserID(r.Context()))
	version := r.FormValue("version")
	description := r.FormValue("description")
	prLink := r.FormValue("pr_link")
	if version == "" {
		setFlash(w, "error", "La versión es obligatoria")
		http.Redirect(w, r, "/releases/new", http.StatusSeeOther)
		return
	}
	rel, err := h.service.CreateRelease(r.Context(), version, description, prLink, userID)
	if err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/releases/new", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Release creado correctamente")
	http.Redirect(w, r, fmt.Sprintf("/releases/%s", rel.ID), http.StatusSeeOther)
}

func (h *TestingHandler) ShowRelease(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	rel, err := h.service.GetRelease(r.Context(), id)
	if err != nil {
		http.Error(w, "Release no encontrado", http.StatusNotFound)
		return
	}
	executions, _ := h.service.GetReleaseExecutions(r.Context(), id)
	statuses := domain.AllStatuses()
	h.renderer.ExecuteTemplate(w, "releases/show.html", withFlash(w, r, map[string]any{
		"Release":    rel,
		"Executions": executions,
		"Statuses":   statuses,
	}))
}

// -- Executions --------------------------------------------------------------

func (h *TestingHandler) ShowExecution(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	exec, err := h.service.GetExecution(r.Context(), id)
	if err != nil {
		http.Error(w, "Ejecución no encontrada", http.StatusNotFound)
		return
	}
	// Also get the test case for reference
	tc, _ := h.service.GetTestCase(r.Context(), exec.TestCaseID)
	statuses := domain.AllStatuses()
	h.renderer.ExecuteTemplate(w, "testcases/show.html", withFlash(w, r, map[string]any{
		"Execution": exec,
		"TestCase":  tc,
		"Statuses":  statuses,
	}))
}

func (h *TestingHandler) UpdateExecution(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	userID, _ := uuid.Parse(GetUserID(r.Context()))
	status := domain.ExecutionStatus(r.FormValue("status"))
	notes := r.FormValue("notes")

	// Handle optional screenshot upload
	var screenshotURL string
	file, _, err := r.FormFile("screenshot")
	if err == nil {
		defer file.Close()
		result, err := h.uploader.Upload(r.Context(), file, "mltestsuite/screenshots")
		if err == nil {
			screenshotURL = result.SecureURL
		}
	}

	if err := h.service.UpdateExecutionStatus(r.Context(), id, status, notes, screenshotURL, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the execution to find the release ID for redirect
	exec, _ := h.service.GetExecution(r.Context(), id)
	if exec != nil {
		setFlash(w, "success", "Estado actualizado")
		http.Redirect(w, r, fmt.Sprintf("/releases/%s", exec.ReleaseID), http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *TestingHandler) CompareJSON(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	fieldName := r.FormValue("field_name")
	actualJSON := r.FormValue("actual_json")
	if fieldName == "" || actualJSON == "" {
		http.Error(w, "field_name y actual_json son obligatorios", http.StatusBadRequest)
		return
	}
	ef, err := h.service.CompareJSON(r.Context(), id, fieldName, actualJSON)
	if err != nil {
		expectedJSON := ""
		if exec, execErr := h.service.GetExecution(r.Context(), id); execErr == nil {
			for _, f := range exec.Fields {
				if f.FieldName == fieldName {
					expectedJSON = f.ExpectedJSON
					break
				}
			}
		}
		h.renderer.ExecuteTemplate(w, "partials/execution_row.html", map[string]any{
			"Field": &domain.ExecutionField{
				ExecutionID:  id,
				FieldName:    fieldName,
				ExpectedJSON: expectedJSON,
				ActualJSON:   actualJSON,
				ErrorMsg:     err.Error(),
			},
		})
		return
	}
	// Return the result as a partial HTML row
	h.renderer.ExecuteTemplate(w, "partials/execution_row.html", map[string]any{
		"Field": ef,
	})
}

// -- Knowledge ---------------------------------------------------------------

func (h *TestingHandler) ShowKnowledge(w http.ResponseWriter, r *http.Request) {
	doc, _ := h.service.GetBusinessKnowledge(r.Context())
	docs, _ := h.service.ListKnowledgeDocs(r.Context())
	reports, _ := h.service.ListReports(r.Context())
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ReportType < reports[j].ReportType
	})
	h.renderer.ExecuteTemplate(w, "knowledge/show.html", withFlash(w, r, map[string]any{
		"Document":  doc,
		"Documents": docs,
		"Reports":   reports,
	}))
}

func (h *TestingHandler) UpdateKnowledge(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(GetUserID(r.Context()))
	title := r.FormValue("title")
	content := r.FormValue("content")
	reportType := r.FormValue("report_type")

	idStr := r.FormValue("id")
	var docID uuid.UUID
	if idStr != "" {
		docID, _ = uuid.Parse(idStr)
	}
	if docID == uuid.Nil {
		docID = uuid.New()
	}

	doc := &domain.KnowledgeDoc{
		ID:          docID,
		Title:       title,
		Content:     content,
		ReportType:  reportType,
		CreatedByID: userID,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
	}
	if err := h.service.SaveKnowledgeDoc(r.Context(), doc); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/knowledge", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Documento guardado")
	http.Redirect(w, r, "/knowledge", http.StatusSeeOther)
}

// -- Export Excel -------------------------------------------------------------

func (h *TestingHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	rel, err := h.service.GetRelease(r.Context(), id)
	if err != nil {
		http.Error(w, "Release no encontrado", http.StatusNotFound)
		return
	}
	executions, err := h.service.GetReleaseExecutions(r.Context(), id)
	if err != nil {
		http.Error(w, "Error obteniendo ejecuciones", http.StatusInternalServerError)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Ejecuciones"
	idx, _ := f.NewSheet(sheet)
	f.SetActiveSheet(idx)
	// Delete default Sheet1 if it exists
	f.DeleteSheet("Sheet1")

	// Headers
	headers := []string{"Equipo", "Reporte", "Caso de Prueba", "Prioridad", "Estado", "Notas", "Ejecutado Por", "Ejecutado En"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Style for header row
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#DDDDDD"}},
	})
	f.SetRowStyle(sheet, 1, 1, style)

	// Data rows
	for i, exec := range executions {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), exec.TeamName)
		f.SetCellValue(sheet, cellName(2, row), exec.ReportName)
		f.SetCellValue(sheet, cellName(3, row), exec.TestCaseTitle)
		f.SetCellValue(sheet, cellName(4, row), exec.ReportType)
		f.SetCellValue(sheet, cellName(5, row), exec.Status.Label())
		f.SetCellValue(sheet, cellName(6, row), exec.Notes)
		f.SetCellValue(sheet, cellName(7, row), exec.ExecutedByName)
		if exec.ExecutedAt != nil {
			f.SetCellValue(sheet, cellName(8, row), exec.ExecutedAt.Format("2006-01-02 15:04"))
		}
	}

	// Auto-width columns
	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 20)
	}

	filename := fmt.Sprintf("release_%s.xlsx", rel.Version)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if _, err := f.WriteTo(w); err != nil {
		slog.Error("excel write error", "error", err)
	}
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

// Ensure io.Reader is used (compile check for uploader interface)
var _ io.Reader
