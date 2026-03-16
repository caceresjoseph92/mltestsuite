package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	domain "mltestsuite/internal/domain/testing"

	"github.com/google/uuid"
)

type Service struct {
	teamRepo      domain.TeamRepository
	reportRepo    domain.ReportRepository
	testCaseRepo  domain.TestCaseRepository
	releaseRepo   domain.ReleaseRepository
	execRepo      domain.ExecutionRepository
	knowledgeRepo domain.KnowledgeRepository
}

func NewService(
	teamRepo domain.TeamRepository,
	reportRepo domain.ReportRepository,
	testCaseRepo domain.TestCaseRepository,
	releaseRepo domain.ReleaseRepository,
	execRepo domain.ExecutionRepository,
	knowledgeRepo domain.KnowledgeRepository,
) *Service {
	return &Service{teamRepo, reportRepo, testCaseRepo, releaseRepo, execRepo, knowledgeRepo}
}

// -- Teams -------------------------------------------------------------------

func (s *Service) ListTeams(ctx context.Context) ([]*domain.Team, error) {
	return s.teamRepo.FindAll(ctx)
}

func (s *Service) GetTeam(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	return s.teamRepo.FindByID(ctx, id)
}

func (s *Service) CreateTeam(ctx context.Context, name, description string) error {
	return s.teamRepo.Save(ctx, &domain.Team{
		ID: uuid.New(), Name: name, Description: description, Active: true, CreatedAt: time.Now(),
	})
}

func (s *Service) UpdateTeam(ctx context.Context, id uuid.UUID, name, description string) error {
	t, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	t.Name = name
	t.Description = description
	return s.teamRepo.Update(ctx, t)
}

// -- Reports -----------------------------------------------------------------

func (s *Service) ListReports(ctx context.Context) ([]*domain.Report, error) {
	return s.reportRepo.FindAll(ctx)
}

func (s *Service) GetReport(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	return s.reportRepo.FindByID(ctx, id)
}

func (s *Service) CreateReport(ctx context.Context, teamID uuid.UUID, name, reportType, description string) error {
	return s.reportRepo.Save(ctx, &domain.Report{
		ID: uuid.New(), TeamID: teamID, Name: name, ReportType: reportType,
		Description: description, Active: true, CreatedAt: time.Now(),
	})
}

func (s *Service) UpdateReport(ctx context.Context, id, teamID uuid.UUID, name, reportType, description string) error {
	rep, err := s.reportRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	rep.TeamID = teamID
	rep.Name = name
	rep.ReportType = reportType
	rep.Description = description
	return s.reportRepo.Update(ctx, rep)
}

// -- TestCases ---------------------------------------------------------------

func (s *Service) ListTestCases(ctx context.Context) ([]*domain.TestCase, error) {
	return s.testCaseRepo.FindAll(ctx)
}

func (s *Service) ListTestCasesByReport(ctx context.Context, reportID uuid.UUID) ([]*domain.TestCase, error) {
	return s.testCaseRepo.FindByReportID(ctx, reportID)
}

func (s *Service) GetTestCase(ctx context.Context, id uuid.UUID) (*domain.TestCase, error) {
	return s.testCaseRepo.FindByID(ctx, id)
}

type CreateTestCaseInput struct {
	ReportID       uuid.UUID
	Title          string
	Preconditions  string
	Steps          string
	ExpectedResult string
	Priority       domain.Priority
	CreatedByID    uuid.UUID
}

func (s *Service) CreateTestCase(ctx context.Context, input CreateTestCaseInput) (*domain.TestCase, error) {
	tc := &domain.TestCase{
		ID: uuid.New(), ReportID: input.ReportID, Title: input.Title,
		Preconditions: input.Preconditions, Steps: input.Steps,
		ExpectedResult: input.ExpectedResult, Priority: input.Priority,
		Active: true, CreatedByID: input.CreatedByID, CreatedAt: time.Now(),
	}
	if err := s.testCaseRepo.Save(ctx, tc); err != nil {
		return nil, err
	}
	go s.refreshKnowledge(context.Background())
	return tc, nil
}

func (s *Service) UpdateTestCase(ctx context.Context, id uuid.UUID, input CreateTestCaseInput) error {
	tc, err := s.testCaseRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	tc.ReportID = input.ReportID
	tc.Title = input.Title
	tc.Preconditions = input.Preconditions
	tc.Steps = input.Steps
	tc.ExpectedResult = input.ExpectedResult
	tc.Priority = input.Priority
	return s.testCaseRepo.Update(ctx, tc)
}

func (s *Service) DeleteTestCase(ctx context.Context, id uuid.UUID) error {
	if err := s.testCaseRepo.Delete(ctx, id); err != nil {
		return err
	}
	go s.refreshKnowledge(context.Background())
	return nil
}

func (s *Service) UpdateTestCaseImage(ctx context.Context, id uuid.UUID, imageURL string) error {
	tc, err := s.testCaseRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	tc.ReferenceImageURL = imageURL
	return s.testCaseRepo.Update(ctx, tc)
}

func (s *Service) SaveTestCaseField(ctx context.Context, tcID uuid.UUID, fieldName, expectedJSON string) error {
	if !json.Valid([]byte(expectedJSON)) {
		return fmt.Errorf("JSON inválido")
	}
	return s.testCaseRepo.SaveField(ctx, &domain.TestCaseField{
		ID: uuid.New(), TestCaseID: tcID, FieldName: fieldName,
		ExpectedJSON: expectedJSON, CreatedAt: time.Now(),
	})
}

func (s *Service) DeleteTestCaseField(ctx context.Context, id uuid.UUID) error {
	return s.testCaseRepo.DeleteField(ctx, id)
}

// -- Releases ----------------------------------------------------------------

func (s *Service) ListReleases(ctx context.Context) ([]*domain.Release, error) {
	releases, err := s.releaseRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, rel := range releases {
		total, passed, failed, pending, err := s.execRepo.CountByRelease(ctx, rel.ID)
		if err == nil {
			rel.TotalCases = total
			rel.PassedCases = passed
			rel.FailedCases = failed
			rel.PendingCases = pending
		}
	}
	return releases, nil
}

func (s *Service) GetRelease(ctx context.Context, id uuid.UUID) (*domain.Release, error) {
	rel, err := s.releaseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	total, passed, failed, pending, err := s.execRepo.CountByRelease(ctx, rel.ID)
	if err == nil {
		rel.TotalCases = total
		rel.PassedCases = passed
		rel.FailedCases = failed
		rel.PendingCases = pending
	}
	return rel, nil
}

func (s *Service) CreateRelease(ctx context.Context, version, description, prLink string, createdByID uuid.UUID) (*domain.Release, error) {
	rel := &domain.Release{
		ID: uuid.New(), Version: version, Description: description, PRLink: prLink,
		CreatedByID: createdByID, Status: domain.ReleaseInProgress, CreatedAt: time.Now(),
	}
	if err := s.releaseRepo.Save(ctx, rel); err != nil {
		return nil, err
	}

	// Auto-create executions for ALL active test cases
	tcs, err := s.testCaseRepo.FindAll(ctx)
	if err != nil {
		return rel, nil
	}
	for _, tc := range tcs {
		_ = s.execRepo.Save(ctx, &domain.Execution{
			ID: uuid.New(), ReleaseID: rel.ID, TestCaseID: tc.ID,
			Status: domain.StatusPending, CreatedAt: time.Now(),
		})
	}
	return rel, nil
}

// -- Executions --------------------------------------------------------------

func (s *Service) GetReleaseExecutions(ctx context.Context, releaseID uuid.UUID) ([]*domain.Execution, error) {
	return s.execRepo.FindByReleaseID(ctx, releaseID)
}

func (s *Service) GetExecution(ctx context.Context, id uuid.UUID) (*domain.Execution, error) {
	exec, err := s.execRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Enrich with expected JSON from test case fields
	tcFields, _ := s.testCaseRepo.FindFieldsByTestCaseID(ctx, exec.TestCaseID)
	execFieldMap := make(map[string]domain.ExecutionField)
	for _, ef := range exec.Fields {
		execFieldMap[ef.FieldName] = ef
	}
	// Merge: add expected JSON to execution fields
	var merged []domain.ExecutionField
	for _, tcf := range tcFields {
		ef, ok := execFieldMap[tcf.FieldName]
		if !ok {
			ef = domain.ExecutionField{FieldName: tcf.FieldName}
		}
		ef.ExpectedJSON = tcf.ExpectedJSON
		merged = append(merged, ef)
	}
	exec.Fields = merged
	return exec, nil
}

func (s *Service) UpdateExecutionStatus(ctx context.Context, id uuid.UUID, status domain.ExecutionStatus, notes, screenshotURL string, executedByID uuid.UUID) error {
	err := s.execRepo.UpdateStatus(ctx, id, status, notes, screenshotURL, executedByID)
	if err != nil {
		return err
	}
	// Check if release is now 100% complete
	exec, err := s.execRepo.FindByID(ctx, id)
	if err != nil {
		return nil
	}
	total, passed, _, _, err := s.execRepo.CountByRelease(ctx, exec.ReleaseID)
	if err == nil && total > 0 && passed == total {
		rel, err := s.releaseRepo.FindByID(ctx, exec.ReleaseID)
		if err == nil {
			rel.Status = domain.ReleaseDone
			_ = s.releaseRepo.Update(ctx, rel)
		}
	}
	return nil
}

func (s *Service) CompareJSON(ctx context.Context, execID uuid.UUID, fieldName, actualJSON string) (*domain.ExecutionField, error) {
	if !json.Valid([]byte(actualJSON)) {
		return nil, fmt.Errorf("JSON inválido")
	}

	// Get expected JSON from test case field
	exec, err := s.execRepo.FindByID(ctx, execID)
	if err != nil {
		return nil, err
	}
	tcFields, _ := s.testCaseRepo.FindFieldsByTestCaseID(ctx, exec.TestCaseID)
	var expectedJSON string
	for _, f := range tcFields {
		if f.FieldName == fieldName {
			expectedJSON = f.ExpectedJSON
			break
		}
	}

	matches := jsonDeepEqual(expectedJSON, actualJSON)
	ef := &domain.ExecutionField{
		ID: uuid.New(), ExecutionID: execID, FieldName: fieldName,
		ExpectedJSON: expectedJSON, ActualJSON: actualJSON, Matches: &matches, CreatedAt: time.Now(),
	}
	if err := s.execRepo.UpsertField(ctx, ef); err != nil {
		return nil, err
	}
	return ef, nil
}

func jsonDeepEqual(a, b string) bool {
	var objA, objB any
	if err := json.Unmarshal([]byte(a), &objA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &objB); err != nil {
		return false
	}
	aBytes, _ := json.Marshal(objA)
	bBytes, _ := json.Marshal(objB)
	return string(aBytes) == string(bBytes)
}

// -- Knowledge ---------------------------------------------------------------

func (s *Service) GetBusinessKnowledge(ctx context.Context) (*domain.KnowledgeDoc, error) {
	return s.knowledgeRepo.FindMain(ctx)
}

func (s *Service) SaveKnowledgeDoc(ctx context.Context, doc *domain.KnowledgeDoc) error {
	return s.knowledgeRepo.Upsert(ctx, doc)
}

func (s *Service) ListKnowledgeDocs(ctx context.Context) ([]*domain.KnowledgeDoc, error) {
	return s.knowledgeRepo.FindAll(ctx)
}

// refreshKnowledge regenerates BUSINESS_KNOWLEDGE.md content from all test cases
func (s *Service) refreshKnowledge(ctx context.Context) {
	tcs, err := s.testCaseRepo.FindAll(ctx)
	if err != nil {
		return
	}

	var sb strings.Builder
	sb.WriteString("# BUSINESS_KNOWLEDGE — MLTestSuite\n\n")
	sb.WriteString("> Documento generado automáticamente. Actualizado al crear/eliminar casos de prueba.\n\n")
	sb.WriteString(fmt.Sprintf("**Total de casos:** %d\n\n", len(tcs)))
	sb.WriteString("---\n\n")

	// Group by report
	byReport := make(map[string][]*domain.TestCase)
	for _, tc := range tcs {
		key := tc.ReportType
		byReport[key] = append(byReport[key], tc)
	}

	for reportType, cases := range byReport {
		sb.WriteString(fmt.Sprintf("## %s\n\n", reportType))
		if len(cases) > 0 {
			sb.WriteString(fmt.Sprintf("**Equipo:** %s | **Reporte:** %s\n\n", cases[0].TeamName, cases[0].ReportName))
		}
		for _, tc := range cases {
			sb.WriteString(fmt.Sprintf("### %s\n", tc.Title))
			sb.WriteString(fmt.Sprintf("- **Prioridad:** %s\n", tc.Priority))
			if tc.Preconditions != "" {
				sb.WriteString(fmt.Sprintf("- **Precondiciones:** %s\n", tc.Preconditions))
			}
			if tc.ExpectedResult != "" {
				sb.WriteString(fmt.Sprintf("- **Resultado esperado:** %s\n", tc.ExpectedResult))
			}
			sb.WriteString("\n")
		}
	}

	doc := &domain.KnowledgeDoc{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000099"),
		Title:      "BUSINESS_KNOWLEDGE",
		Content:    sb.String(),
		ReportType: "all",
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	_ = s.knowledgeRepo.Upsert(ctx, doc)
}

// ExportClaudePrompt generates a prompt for Claude to generate test cases
func (s *Service) ExportClaudePrompt(ctx context.Context, reportID uuid.UUID) (string, error) {
	report, err := s.reportRepo.FindByID(ctx, reportID)
	if err != nil {
		return "", err
	}

	existingCases, _ := s.testCaseRepo.FindByReportID(ctx, reportID)
	knowledge, _ := s.knowledgeRepo.FindMain(ctx)

	var knowledgeContent string
	if knowledge != nil {
		knowledgeContent = knowledge.Content
	}

	var existingTitles []string
	for _, tc := range existingCases {
		existingTitles = append(existingTitles, "- "+tc.Title)
	}

	prompt := fmt.Sprintf(`Eres un experto en pruebas de software para el sistema ADR de MercadoLibre (fury_mprc-available-detail-report).

## Conocimiento del negocio
%s

---

## Reporte objetivo
**Tipo:** %s
**Nombre:** %s
**Descripción:** %s

## Casos existentes (NO duplicar)
%s

## Tu tarea
Genera nuevos casos de prueba para el siguiente desarrollo:

[DESCRIBE AQUÍ TU DESARROLLO]

## Formato de salida REQUERIDO (para importar en MLTestSuite)

Usa exactamente este formato para cada caso:

---
## TC: [Título del caso]
priority: high|medium|low
preconditions: [Precondiciones necesarias]
expected_result: [Qué debe ocurrir]

### request_parameters_json
[JSON esperado]

### [otro_campo_json si aplica]
[JSON esperado]

---

Genera mínimo 3 casos cubriendo: flujo exitoso, fallos esperados, casos borde.`,
		knowledgeContent,
		report.ReportType,
		report.Name,
		report.Description,
		strings.Join(existingTitles, "\n"),
	)
	return prompt, nil
}

// ImportFromMarkdown parses a Claude-generated .md and creates test cases
func (s *Service) ImportFromMarkdown(ctx context.Context, reportID uuid.UUID, content string, createdByID uuid.UUID) (int, error) {
	cases := parseMarkdownCases(content)
	count := 0
	for _, tc := range cases {
		tc.ReportID = reportID
		tc.CreatedByID = createdByID
		tc.Active = true
		tc.CreatedAt = time.Now()
		if tc.ID == uuid.Nil {
			tc.ID = uuid.New()
		}
		if err := s.testCaseRepo.Save(ctx, tc); err != nil {
			continue
		}
		for _, f := range tc.Fields {
			f.TestCaseID = tc.ID
			if f.ID == uuid.Nil {
				f.ID = uuid.New()
			}
			_ = s.testCaseRepo.SaveField(ctx, &f)
		}
		count++
	}
	go s.refreshKnowledge(context.Background())
	return count, nil
}

// isTCHeader returns true if the line starts a new test case block.
func isTCHeader(line string) bool {
	s := strings.TrimSpace(line)
	return strings.HasPrefix(s, "## TC:") || strings.HasPrefix(s, "TC:")
}

// snakeCaseFieldRe matches a bare snake_case identifier like request_parameters_json
var snakeCaseFieldRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// parseMarkdownCases parses .md format — supports both ## TC: and TC: headers,
// and both --- separators and TC: header-based splitting.
func parseMarkdownCases(content string) []*domain.TestCase {
	lines := strings.Split(content, "\n")
	var sections [][]string
	var current []string
	for _, line := range lines {
		if isTCHeader(line) && len(current) > 0 {
			sections = append(sections, current)
			current = nil
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		sections = append(sections, current)
	}

	var cases []*domain.TestCase
	for _, sec := range sections {
		tc := parseOneCase(strings.Join(sec, "\n"))
		if tc != nil && tc.Title != "" {
			cases = append(cases, tc)
		}
	}
	return cases
}

func parseOneCase(section string) *domain.TestCase {
	lines := strings.Split(strings.TrimSpace(section), "\n")
	if len(lines) == 0 {
		return nil
	}
	tc := &domain.TestCase{Priority: domain.PriorityMedium}
	var currentField string
	var fieldContent strings.Builder

	flushField := func() {
		if currentField != "" {
			content := strings.TrimSpace(fieldContent.String())
			// Remove ```json ... ``` wrapper if present
			content = strings.TrimPrefix(content, "```json\n")
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimSuffix(content, "\n```")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
			if json.Valid([]byte(content)) {
				tc.Fields = append(tc.Fields, domain.TestCaseField{
					FieldName: currentField, ExpectedJSON: content,
				})
			}
			currentField = ""
			fieldContent.Reset()
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## TC:") {
			tc.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "## TC:"))
		} else if strings.HasPrefix(trimmed, "TC:") {
			tc.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "TC:"))
		} else if strings.HasPrefix(trimmed, "priority:") {
			p := strings.TrimSpace(strings.TrimPrefix(trimmed, "priority:"))
			tc.Priority = domain.Priority(p)
		} else if strings.HasPrefix(trimmed, "preconditions:") {
			tc.Preconditions = strings.TrimSpace(strings.TrimPrefix(trimmed, "preconditions:"))
		} else if strings.HasPrefix(trimmed, "expected_result:") {
			tc.ExpectedResult = strings.TrimSpace(strings.TrimPrefix(trimmed, "expected_result:"))
		} else if strings.HasPrefix(line, "### ") {
			flushField()
			currentField = strings.TrimSpace(strings.TrimPrefix(line, "### "))
		} else if trimmed != "" && snakeCaseFieldRe.MatchString(trimmed) && currentField == "" {
			// bare snake_case identifier on its own line → treat as field name
			flushField()
			currentField = trimmed
		} else if currentField != "" {
			fieldContent.WriteString(line + "\n")
		}
	}
	flushField()
	return tc
}
