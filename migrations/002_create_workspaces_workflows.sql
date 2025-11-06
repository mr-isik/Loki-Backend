-- Create workspaces table
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create workflow_status enum
DO $$ BEGIN
    CREATE TYPE workflow_status AS ENUM ('draft', 'published', 'archived');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create workflows table
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL DEFAULT 'Untitled Workflow',
    status workflow_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_workspaces_owner ON workspaces(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_created_at ON workspaces(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workflows_workspace ON workflows(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflows_updated_at ON workflows(updated_at DESC);

-- Add comments
COMMENT ON TABLE workspaces IS 'User workspaces/teams table';
COMMENT ON COLUMN workspaces.id IS 'Unique workspace identifier (UUID)';
COMMENT ON COLUMN workspaces.owner_user_id IS 'Owner user ID (foreign key to users)';
COMMENT ON COLUMN workspaces.name IS 'Workspace name';

COMMENT ON TABLE workflows IS 'Workflows table';
COMMENT ON COLUMN workflows.id IS 'Unique workflow identifier (UUID)';
COMMENT ON COLUMN workflows.workspace_id IS 'Workspace ID (foreign key to workspaces)';
COMMENT ON COLUMN workflows.title IS 'Workflow title';
COMMENT ON COLUMN workflows.status IS 'Workflow status (draft, published, archived)';
