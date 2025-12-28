-- Workflow table
CREATE TABLE IF NOT EXISTS workflow (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_created_at ON workflow(created_at DESC);

-- NodeDefinition table (belongs to Workflow)
CREATE TABLE IF NOT EXISTS node_definition (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    node_template_id VARCHAR(26) NOT NULL REFERENCES node_template(id),
    name VARCHAR(255) NOT NULL,
    config JSONB,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_node_definition_workflow_id ON node_definition(workflow_id);

-- Edge table (connects NodeDefinitions within a Workflow)
CREATE TABLE IF NOT EXISTS edge (
    id VARCHAR(26) PRIMARY KEY,
    workflow_id VARCHAR(26) NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    from_node_definition_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    to_node_definition_id VARCHAR(26) NOT NULL REFERENCES node_definition(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Prevent duplicate edges
    CONSTRAINT unique_edge UNIQUE (workflow_id, from_node_definition_id, to_node_definition_id),
    -- Prevent self-loops
    CONSTRAINT no_self_loop CHECK (from_node_definition_id != to_node_definition_id)
);

CREATE INDEX IF NOT EXISTS idx_edge_workflow_id ON edge(workflow_id);
CREATE INDEX IF NOT EXISTS idx_edge_from_node ON edge(from_node_definition_id);
CREATE INDEX IF NOT EXISTS idx_edge_to_node ON edge(to_node_definition_id);
