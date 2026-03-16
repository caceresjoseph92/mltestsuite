package testing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("recurso no encontrado")
	ErrNotOwner  = errors.New("no tiene permiso para este recurso")
	ErrDuplicate = errors.New("ya existe un recurso con ese nombre")
)

// Priority for test cases
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// ExecutionStatus
type ExecutionStatus string

const (
	StatusPending ExecutionStatus = "pending"
	StatusPass    ExecutionStatus = "pass"
	StatusFail    ExecutionStatus = "fail"
	StatusBlocked ExecutionStatus = "blocked"
	StatusSkipped ExecutionStatus = "skipped"
)

func (s ExecutionStatus) Label() string {
	switch s {
	case StatusPass:
		return "Paso"
	case StatusFail:
		return "Fallo"
	case StatusBlocked:
		return "Bloqueado"
	case StatusSkipped:
		return "Omitido"
	default:
		return "Pendiente"
	}
}

func (s ExecutionStatus) Color() string {
	switch s {
	case StatusPass:
		return "green"
	case StatusFail:
		return "red"
	case StatusBlocked:
		return "yellow"
	case StatusSkipped:
		return "gray"
	default:
		return "gray"
	}
}

func AllStatuses() []ExecutionStatus {
	return []ExecutionStatus{StatusPending, StatusPass, StatusFail, StatusBlocked, StatusSkipped}
}

// ReleaseStatus
type ReleaseStatus string

const (
	ReleaseInProgress ReleaseStatus = "in_progress"
	ReleaseDone       ReleaseStatus = "done"
)

// Team
type Team struct {
	ID          uuid.UUID
	Name        string
	Description string
	Active      bool
	CreatedAt   time.Time
}

// Report
type Report struct {
	ID          uuid.UUID
	TeamID      uuid.UUID
	TeamName    string
	Name        string
	ReportType  string
	Description string
	Active      bool
	CreatedAt   time.Time
}

// TestCase
type TestCase struct {
	ID                uuid.UUID
	ReportID          uuid.UUID
	ReportName        string
	ReportType        string
	TeamName          string
	Title             string
	Preconditions     string
	Steps             string
	ExpectedResult    string
	Priority          Priority
	ReferenceImageURL string
	Active            bool
	CreatedByID       uuid.UUID
	CreatedByName     string
	CreatedAt         time.Time
	Fields            []TestCaseField
}

// TestCaseField - dynamic JSON comparison field
type TestCaseField struct {
	ID           uuid.UUID
	TestCaseID   uuid.UUID
	FieldName    string
	ExpectedJSON string
	CreatedAt    time.Time
}

// Release
type Release struct {
	ID            uuid.UUID
	Version       string
	Description   string
	PRLink        string
	CreatedByID   uuid.UUID
	CreatedByName string
	Status        ReleaseStatus
	CreatedAt     time.Time
	// computed
	TotalCases   int
	PassedCases  int
	FailedCases  int
	PendingCases int
}

func (r *Release) PercentComplete() int {
	if r.TotalCases == 0 {
		return 0
	}
	return (r.PassedCases * 100) / r.TotalCases
}

func (r *Release) IsComplete() bool {
	return r.TotalCases > 0 && r.PassedCases == r.TotalCases
}

// Execution - one per test_case per release
type Execution struct {
	ID             uuid.UUID
	ReleaseID      uuid.UUID
	TestCaseID     uuid.UUID
	TestCaseTitle  string
	ReportName     string
	ReportType     string
	TeamName       string
	Status         ExecutionStatus
	Notes          string
	ScreenshotURL  string
	ExecutedByID   *uuid.UUID
	ExecutedByName string
	ExecutedAt     *time.Time
	CreatedAt      time.Time
	Fields         []ExecutionField
}

// ExecutionField - result of comparing a JSON field
type ExecutionField struct {
	ID           uuid.UUID
	ExecutionID  uuid.UUID
	FieldName    string
	ExpectedJSON string // from test case field
	ActualJSON   string
	Matches      *bool
	CreatedAt    time.Time
}

// KnowledgeDoc
type KnowledgeDoc struct {
	ID            uuid.UUID
	Title         string
	Content       string
	ReportType    string
	CreatedByID   uuid.UUID
	CreatedByName string
	UpdatedAt     time.Time
	CreatedAt     time.Time
}

// TeamRepository
type TeamRepository interface {
	Save(ctx context.Context, team *Team) error
	FindAll(ctx context.Context) ([]*Team, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Team, error)
	Update(ctx context.Context, team *Team) error
}

// ReportRepository
type ReportRepository interface {
	Save(ctx context.Context, report *Report) error
	FindAll(ctx context.Context) ([]*Report, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Report, error)
	FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*Report, error)
	Update(ctx context.Context, report *Report) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TestCaseRepository
type TestCaseRepository interface {
	Save(ctx context.Context, tc *TestCase) error
	FindByID(ctx context.Context, id uuid.UUID) (*TestCase, error)
	FindByReportID(ctx context.Context, reportID uuid.UUID) ([]*TestCase, error)
	FindAll(ctx context.Context) ([]*TestCase, error)
	Update(ctx context.Context, tc *TestCase) error
	Delete(ctx context.Context, id uuid.UUID) error
	SaveField(ctx context.Context, f *TestCaseField) error
	DeleteField(ctx context.Context, id uuid.UUID) error
	FindFieldsByTestCaseID(ctx context.Context, tcID uuid.UUID) ([]TestCaseField, error)
}

// ReleaseRepository
type ReleaseRepository interface {
	Save(ctx context.Context, r *Release) error
	FindAll(ctx context.Context) ([]*Release, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Release, error)
	Update(ctx context.Context, r *Release) error
}

// ExecutionRepository
type ExecutionRepository interface {
	Save(ctx context.Context, e *Execution) error
	FindByID(ctx context.Context, id uuid.UUID) (*Execution, error)
	FindByReleaseID(ctx context.Context, releaseID uuid.UUID) ([]*Execution, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status ExecutionStatus, notes, screenshotURL string, executedByID uuid.UUID) error
	UpsertField(ctx context.Context, f *ExecutionField) error
	FindFieldsByExecutionID(ctx context.Context, execID uuid.UUID) ([]ExecutionField, error)
	CountByRelease(ctx context.Context, releaseID uuid.UUID) (total, passed, failed, pending int, err error)
}

// KnowledgeRepository
type KnowledgeRepository interface {
	Upsert(ctx context.Context, doc *KnowledgeDoc) error
	FindAll(ctx context.Context) ([]*KnowledgeDoc, error)
	FindByID(ctx context.Context, id uuid.UUID) (*KnowledgeDoc, error)
	FindMain(ctx context.Context) (*KnowledgeDoc, error) // BUSINESS_KNOWLEDGE
	Delete(ctx context.Context, id uuid.UUID) error
}
