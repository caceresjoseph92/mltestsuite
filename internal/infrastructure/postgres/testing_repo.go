package postgres

import (
	"context"
	"errors"
	"time"

	domain "mltestsuite/internal/domain/testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -- Teams -------------------------------------------------------------------

type TeamRepository struct{ pool *pgxpool.Pool }

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository { return &TeamRepository{pool: pool} }

func (r *TeamRepository) Save(ctx context.Context, t *domain.Team) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO teams (id,name,description,active,created_at) VALUES ($1,$2,$3,$4,$5)`,
		t.ID, t.Name, t.Description, t.Active, t.CreatedAt)
	return err
}

func (r *TeamRepository) FindAll(ctx context.Context) ([]*domain.Team, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,name,description,active,created_at FROM teams WHERE active=true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []*domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Active, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, rows.Err()
}

func (r *TeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	var t domain.Team
	err := r.pool.QueryRow(ctx,
		`SELECT id,name,description,active,created_at FROM teams WHERE id=$1`, id).
		Scan(&t.ID, &t.Name, &t.Description, &t.Active, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepository) Update(ctx context.Context, t *domain.Team) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE teams SET name=$2,description=$3,active=$4 WHERE id=$1`,
		t.ID, t.Name, t.Description, t.Active)
	return err
}

// -- Reports -----------------------------------------------------------------

type ReportRepository struct{ pool *pgxpool.Pool }

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository { return &ReportRepository{pool: pool} }

func (r *ReportRepository) Save(ctx context.Context, rep *domain.Report) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reports (id,team_id,name,report_type,description,active,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rep.ID, rep.TeamID, rep.Name, rep.ReportType, rep.Description, rep.Active, rep.CreatedAt)
	return err
}

func (r *ReportRepository) FindAll(ctx context.Context) ([]*domain.Report, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rp.id,rp.team_id,t.name,rp.name,rp.report_type,rp.description,rp.active,rp.created_at
		FROM reports rp JOIN teams t ON t.id=rp.team_id
		WHERE rp.active=true ORDER BY t.name,rp.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Report
	for rows.Next() {
		var rep domain.Report
		if err := rows.Scan(&rep.ID, &rep.TeamID, &rep.TeamName, &rep.Name, &rep.ReportType, &rep.Description, &rep.Active, &rep.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &rep)
	}
	return list, rows.Err()
}

func (r *ReportRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	var rep domain.Report
	err := r.pool.QueryRow(ctx, `
		SELECT rp.id,rp.team_id,t.name,rp.name,rp.report_type,rp.description,rp.active,rp.created_at
		FROM reports rp JOIN teams t ON t.id=rp.team_id WHERE rp.id=$1`, id).
		Scan(&rep.ID, &rep.TeamID, &rep.TeamName, &rep.Name, &rep.ReportType, &rep.Description, &rep.Active, &rep.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *ReportRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*domain.Report, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rp.id,rp.team_id,t.name,rp.name,rp.report_type,rp.description,rp.active,rp.created_at
		FROM reports rp JOIN teams t ON t.id=rp.team_id
		WHERE rp.team_id=$1 AND rp.active=true ORDER BY rp.name`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Report
	for rows.Next() {
		var rep domain.Report
		if err := rows.Scan(&rep.ID, &rep.TeamID, &rep.TeamName, &rep.Name, &rep.ReportType, &rep.Description, &rep.Active, &rep.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &rep)
	}
	return list, rows.Err()
}

func (r *ReportRepository) Update(ctx context.Context, rep *domain.Report) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE reports SET team_id=$2,name=$3,report_type=$4,description=$5,active=$6 WHERE id=$1`,
		rep.ID, rep.TeamID, rep.Name, rep.ReportType, rep.Description, rep.Active)
	return err
}

func (r *ReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE reports SET active=false WHERE id=$1`, id)
	return err
}

// -- TestCases ---------------------------------------------------------------

type TestCaseRepository struct{ pool *pgxpool.Pool }

func NewTestCaseRepository(pool *pgxpool.Pool) *TestCaseRepository {
	return &TestCaseRepository{pool: pool}
}

func (r *TestCaseRepository) Save(ctx context.Context, tc *domain.TestCase) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO test_cases (id,report_id,title,preconditions,steps,expected_result,priority,reference_image_url,active,created_by,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		tc.ID, tc.ReportID, tc.Title, tc.Preconditions, tc.Steps, tc.ExpectedResult,
		string(tc.Priority), tc.ReferenceImageURL, tc.Active, tc.CreatedByID, tc.CreatedAt)
	return err
}

func (r *TestCaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.TestCase, error) {
	var tc domain.TestCase
	var priority string
	err := r.pool.QueryRow(ctx, `
		SELECT tc.id,tc.report_id,rp.name,rp.report_type,t.name,tc.title,tc.preconditions,tc.steps,
		       tc.expected_result,tc.priority,tc.reference_image_url,tc.active,tc.created_by,u.name,tc.created_at
		FROM test_cases tc
		JOIN reports rp ON rp.id=tc.report_id
		JOIN teams t ON t.id=rp.team_id
		JOIN users u ON u.id=tc.created_by
		WHERE tc.id=$1`, id).
		Scan(&tc.ID, &tc.ReportID, &tc.ReportName, &tc.ReportType, &tc.TeamName,
			&tc.Title, &tc.Preconditions, &tc.Steps, &tc.ExpectedResult,
			&priority, &tc.ReferenceImageURL, &tc.Active, &tc.CreatedByID, &tc.CreatedByName, &tc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	tc.Priority = domain.Priority(priority)
	fields, _ := r.FindFieldsByTestCaseID(ctx, tc.ID)
	tc.Fields = fields
	return &tc, nil
}

func (r *TestCaseRepository) FindByReportID(ctx context.Context, reportID uuid.UUID) ([]*domain.TestCase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tc.id,tc.report_id,rp.name,rp.report_type,t.name,tc.title,tc.preconditions,tc.steps,
		       tc.expected_result,tc.priority,tc.reference_image_url,tc.active,tc.created_by,u.name,tc.created_at
		FROM test_cases tc
		JOIN reports rp ON rp.id=tc.report_id
		JOIN teams t ON t.id=rp.team_id
		JOIN users u ON u.id=tc.created_by
		WHERE tc.report_id=$1 AND tc.active=true ORDER BY tc.created_at`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTestCases(rows)
}

func (r *TestCaseRepository) FindAll(ctx context.Context) ([]*domain.TestCase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tc.id,tc.report_id,rp.name,rp.report_type,t.name,tc.title,tc.preconditions,tc.steps,
		       tc.expected_result,tc.priority,tc.reference_image_url,tc.active,tc.created_by,u.name,tc.created_at
		FROM test_cases tc
		JOIN reports rp ON rp.id=tc.report_id
		JOIN teams t ON t.id=rp.team_id
		JOIN users u ON u.id=tc.created_by
		WHERE tc.active=true ORDER BY t.name,rp.name,tc.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTestCases(rows)
}

func scanTestCases(rows pgx.Rows) ([]*domain.TestCase, error) {
	var list []*domain.TestCase
	for rows.Next() {
		var tc domain.TestCase
		var priority string
		if err := rows.Scan(&tc.ID, &tc.ReportID, &tc.ReportName, &tc.ReportType, &tc.TeamName,
			&tc.Title, &tc.Preconditions, &tc.Steps, &tc.ExpectedResult,
			&priority, &tc.ReferenceImageURL, &tc.Active, &tc.CreatedByID, &tc.CreatedByName, &tc.CreatedAt); err != nil {
			return nil, err
		}
		tc.Priority = domain.Priority(priority)
		list = append(list, &tc)
	}
	return list, rows.Err()
}

func (r *TestCaseRepository) Update(ctx context.Context, tc *domain.TestCase) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE test_cases SET title=$2,preconditions=$3,steps=$4,expected_result=$5,priority=$6,reference_image_url=$7 WHERE id=$1`,
		tc.ID, tc.Title, tc.Preconditions, tc.Steps, tc.ExpectedResult, string(tc.Priority), tc.ReferenceImageURL)
	return err
}

func (r *TestCaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE test_cases SET active=false WHERE id=$1`, id)
	return err
}

func (r *TestCaseRepository) SaveField(ctx context.Context, f *domain.TestCaseField) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO test_case_fields (id,test_case_id,field_name,expected_json,created_at) VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (test_case_id,field_name) DO UPDATE SET expected_json=$4`,
		f.ID, f.TestCaseID, f.FieldName, f.ExpectedJSON, f.CreatedAt)
	return err
}

func (r *TestCaseRepository) DeleteField(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM test_case_fields WHERE id=$1`, id)
	return err
}

func (r *TestCaseRepository) FindFieldsByTestCaseID(ctx context.Context, tcID uuid.UUID) ([]domain.TestCaseField, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,test_case_id,field_name,expected_json,created_at FROM test_case_fields WHERE test_case_id=$1 ORDER BY field_name`, tcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fields []domain.TestCaseField
	for rows.Next() {
		var f domain.TestCaseField
		if err := rows.Scan(&f.ID, &f.TestCaseID, &f.FieldName, &f.ExpectedJSON, &f.CreatedAt); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

// -- Releases ----------------------------------------------------------------

type ReleaseRepository struct{ pool *pgxpool.Pool }

func NewReleaseRepository(pool *pgxpool.Pool) *ReleaseRepository {
	return &ReleaseRepository{pool: pool}
}

func (r *ReleaseRepository) Save(ctx context.Context, rel *domain.Release) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO releases (id,version,description,pr_link,created_by,status,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rel.ID, rel.Version, rel.Description, rel.PRLink, rel.CreatedByID, string(rel.Status), rel.CreatedAt)
	return err
}

func (r *ReleaseRepository) FindAll(ctx context.Context) ([]*domain.Release, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id,r.version,r.description,r.pr_link,r.created_by,u.name,r.status,r.created_at
		FROM releases r JOIN users u ON u.id=r.created_by
		ORDER BY r.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Release
	for rows.Next() {
		var rel domain.Release
		var status string
		if err := rows.Scan(&rel.ID, &rel.Version, &rel.Description, &rel.PRLink, &rel.CreatedByID, &rel.CreatedByName, &status, &rel.CreatedAt); err != nil {
			return nil, err
		}
		rel.Status = domain.ReleaseStatus(status)
		list = append(list, &rel)
	}
	return list, rows.Err()
}

func (r *ReleaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Release, error) {
	var rel domain.Release
	var status string
	err := r.pool.QueryRow(ctx, `
		SELECT r.id,r.version,r.description,r.pr_link,r.created_by,u.name,r.status,r.created_at
		FROM releases r JOIN users u ON u.id=r.created_by WHERE r.id=$1`, id).
		Scan(&rel.ID, &rel.Version, &rel.Description, &rel.PRLink, &rel.CreatedByID, &rel.CreatedByName, &status, &rel.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rel.Status = domain.ReleaseStatus(status)
	return &rel, nil
}

func (r *ReleaseRepository) Update(ctx context.Context, rel *domain.Release) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE releases SET version=$2,description=$3,pr_link=$4,status=$5 WHERE id=$1`,
		rel.ID, rel.Version, rel.Description, rel.PRLink, string(rel.Status))
	return err
}

// -- Executions --------------------------------------------------------------

type ExecutionRepository struct{ pool *pgxpool.Pool }

func NewExecutionRepository(pool *pgxpool.Pool) *ExecutionRepository {
	return &ExecutionRepository{pool: pool}
}

func (r *ExecutionRepository) Save(ctx context.Context, e *domain.Execution) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO executions (id,release_id,test_case_id,status,notes,screenshot_url,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.ID, e.ReleaseID, e.TestCaseID, string(e.Status), e.Notes, e.ScreenshotURL, e.CreatedAt)
	return err
}

func (r *ExecutionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Execution, error) {
	var e domain.Execution
	var status string
	var executedByID *uuid.UUID
	var executedByName *string
	var executedAt *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT e.id,e.release_id,e.test_case_id,tc.title,rp.name,rp.report_type,t.name,
		       e.status,e.notes,e.screenshot_url,e.executed_by,u2.name,e.executed_at,e.created_at
		FROM executions e
		JOIN test_cases tc ON tc.id=e.test_case_id
		JOIN reports rp ON rp.id=tc.report_id
		JOIN teams t ON t.id=rp.team_id
		LEFT JOIN users u2 ON u2.id=e.executed_by
		WHERE e.id=$1`, id).
		Scan(&e.ID, &e.ReleaseID, &e.TestCaseID, &e.TestCaseTitle, &e.ReportName, &e.ReportType, &e.TeamName,
			&status, &e.Notes, &e.ScreenshotURL, &executedByID, &executedByName, &executedAt, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e.Status = domain.ExecutionStatus(status)
	e.ExecutedByID = executedByID
	if executedByName != nil {
		e.ExecutedByName = *executedByName
	}
	e.ExecutedAt = executedAt
	fields, _ := r.FindFieldsByExecutionID(ctx, e.ID)
	e.Fields = fields
	return &e, nil
}

func (r *ExecutionRepository) FindByReleaseID(ctx context.Context, releaseID uuid.UUID) ([]*domain.Execution, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id,e.release_id,e.test_case_id,tc.title,rp.name,rp.report_type,t.name,
		       e.status,e.notes,e.screenshot_url,e.executed_by,u2.name,e.executed_at,e.created_at
		FROM executions e
		JOIN test_cases tc ON tc.id=e.test_case_id
		JOIN reports rp ON rp.id=tc.report_id
		JOIN teams t ON t.id=rp.team_id
		LEFT JOIN users u2 ON u2.id=e.executed_by
		WHERE e.release_id=$1 ORDER BY t.name,rp.name,tc.title`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Execution
	for rows.Next() {
		var e domain.Execution
		var status string
		var executedByID *uuid.UUID
		var executedByName *string
		var executedAt *time.Time
		if err := rows.Scan(&e.ID, &e.ReleaseID, &e.TestCaseID, &e.TestCaseTitle, &e.ReportName, &e.ReportType, &e.TeamName,
			&status, &e.Notes, &e.ScreenshotURL, &executedByID, &executedByName, &executedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Status = domain.ExecutionStatus(status)
		e.ExecutedByID = executedByID
		if executedByName != nil {
			e.ExecutedByName = *executedByName
		}
		e.ExecutedAt = executedAt
		list = append(list, &e)
	}
	return list, rows.Err()
}

func (r *ExecutionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ExecutionStatus, notes, screenshotURL string, executedByID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE executions SET status=$2,notes=$3,screenshot_url=CASE WHEN $4!='' THEN $4 ELSE screenshot_url END,executed_by=$5,executed_at=$6 WHERE id=$1`,
		id, string(status), notes, screenshotURL, executedByID, now)
	return err
}

func (r *ExecutionRepository) UpsertField(ctx context.Context, f *domain.ExecutionField) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO execution_fields (id,execution_id,field_name,actual_json,matches,created_at) VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (execution_id,field_name) DO UPDATE SET actual_json=$4,matches=$5`,
		f.ID, f.ExecutionID, f.FieldName, f.ActualJSON, f.Matches, f.CreatedAt)
	return err
}

func (r *ExecutionRepository) FindFieldsByExecutionID(ctx context.Context, execID uuid.UUID) ([]domain.ExecutionField, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ef.id,ef.execution_id,ef.field_name,COALESCE(tcf.expected_json,'{}'),ef.actual_json,ef.matches,ef.created_at
		FROM execution_fields ef
		LEFT JOIN executions e ON e.id=ef.execution_id
		LEFT JOIN test_case_fields tcf ON tcf.test_case_id=e.test_case_id AND tcf.field_name=ef.field_name
		WHERE ef.execution_id=$1 ORDER BY ef.field_name`, execID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fields []domain.ExecutionField
	for rows.Next() {
		var f domain.ExecutionField
		if err := rows.Scan(&f.ID, &f.ExecutionID, &f.FieldName, &f.ExpectedJSON, &f.ActualJSON, &f.Matches, &f.CreatedAt); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

func (r *ExecutionRepository) CountByRelease(ctx context.Context, releaseID uuid.UUID) (total, passed, failed, pending int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*),
		       SUM(CASE WHEN status='pass' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status='fail' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status='pending' THEN 1 ELSE 0 END)
		FROM executions WHERE release_id=$1`, releaseID).
		Scan(&total, &passed, &failed, &pending)
	return
}

// -- KnowledgeDocs -----------------------------------------------------------

type KnowledgeRepository struct{ pool *pgxpool.Pool }

func NewKnowledgeRepository(pool *pgxpool.Pool) *KnowledgeRepository {
	return &KnowledgeRepository{pool: pool}
}

func (r *KnowledgeRepository) Upsert(ctx context.Context, doc *domain.KnowledgeDoc) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO knowledge_docs (id,title,content,report_type,created_by,updated_at,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (id) DO UPDATE SET title=$2,content=$3,report_type=$4,updated_at=$6`,
		doc.ID, doc.Title, doc.Content, doc.ReportType, doc.CreatedByID, doc.UpdatedAt, doc.CreatedAt)
	return err
}

func (r *KnowledgeRepository) FindAll(ctx context.Context) ([]*domain.KnowledgeDoc, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT kd.id,kd.title,kd.content,kd.report_type,kd.created_by,u.name,kd.updated_at,kd.created_at
		FROM knowledge_docs kd JOIN users u ON u.id=kd.created_by
		ORDER BY kd.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.KnowledgeDoc
	for rows.Next() {
		var doc domain.KnowledgeDoc
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.ReportType, &doc.CreatedByID, &doc.CreatedByName, &doc.UpdatedAt, &doc.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &doc)
	}
	return list, rows.Err()
}

func (r *KnowledgeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.KnowledgeDoc, error) {
	var doc domain.KnowledgeDoc
	err := r.pool.QueryRow(ctx, `
		SELECT kd.id,kd.title,kd.content,kd.report_type,kd.created_by,u.name,kd.updated_at,kd.created_at
		FROM knowledge_docs kd JOIN users u ON u.id=kd.created_by WHERE kd.id=$1`, id).
		Scan(&doc.ID, &doc.Title, &doc.Content, &doc.ReportType, &doc.CreatedByID, &doc.CreatedByName, &doc.UpdatedAt, &doc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *KnowledgeRepository) FindMain(ctx context.Context) (*domain.KnowledgeDoc, error) {
	var doc domain.KnowledgeDoc
	err := r.pool.QueryRow(ctx, `
		SELECT kd.id,kd.title,kd.content,kd.report_type,kd.created_by,u.name,kd.updated_at,kd.created_at
		FROM knowledge_docs kd JOIN users u ON u.id=kd.created_by
		WHERE kd.title='BUSINESS_KNOWLEDGE'
		ORDER BY kd.updated_at DESC LIMIT 1`).
		Scan(&doc.ID, &doc.Title, &doc.Content, &doc.ReportType, &doc.CreatedByID, &doc.CreatedByName, &doc.UpdatedAt, &doc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *KnowledgeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM knowledge_docs WHERE id=$1`, id)
	return err
}
