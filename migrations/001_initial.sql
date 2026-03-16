-- Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    active BOOLEAN NOT NULL DEFAULT true,
    notification_emails TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Teams
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Reports
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id),
    name TEXT NOT NULL,
    report_type TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Test Cases
CREATE TABLE IF NOT EXISTS test_cases (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL REFERENCES reports(id),
    title TEXT NOT NULL,
    preconditions TEXT NOT NULL DEFAULT '',
    steps TEXT NOT NULL DEFAULT '',
    expected_result TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'medium',
    reference_image_url TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Test Case JSON Fields (dynamic comparison fields)
CREATE TABLE IF NOT EXISTS test_case_fields (
    id UUID PRIMARY KEY,
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL,
    expected_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(test_case_id, field_name)
);

-- Releases
CREATE TABLE IF NOT EXISTS releases (
    id UUID PRIMARY KEY,
    version TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    pr_link TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'in_progress',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Executions (one per test_case per release, shared)
CREATE TABLE IF NOT EXISTS executions (
    id UUID PRIMARY KEY,
    release_id UUID NOT NULL REFERENCES releases(id),
    test_case_id UUID NOT NULL REFERENCES test_cases(id),
    status TEXT NOT NULL DEFAULT 'pending',
    notes TEXT NOT NULL DEFAULT '',
    screenshot_url TEXT NOT NULL DEFAULT '',
    executed_by UUID REFERENCES users(id),
    executed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(release_id, test_case_id)
);

-- Execution JSON Field Results
CREATE TABLE IF NOT EXISTS execution_fields (
    id UUID PRIMARY KEY,
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL,
    actual_json TEXT NOT NULL DEFAULT '{}',
    matches BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(execution_id, field_name)
);

-- Knowledge Documents (BUSINESS_KNOWLEDGE stored in DB)
CREATE TABLE IF NOT EXISTS knowledge_docs (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    report_type TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
